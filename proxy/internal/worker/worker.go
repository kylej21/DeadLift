package worker

import (
	"context"
	"encoding/base64"
	"log"
	"time"

	"github.com/google/uuid"
	"proxy/internal/mcp"
	"proxy/internal/models"
	"proxy/internal/pubsub"
	"proxy/internal/store"
)

// Worker polls each org's DLQ subscription and creates repair tasks.
type Worker struct {
	RepairSA string
	Store    *store.Store
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("worker: starting DLQ polling")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	active := map[string]context.CancelFunc{}

	launch := func() {
		users, err := w.Store.GetAllUsers(ctx)
		if err != nil {
			log.Printf("worker: failed to load orgs: %v", err)
			return
		}
		for _, u := range users {
			if _, running := active[u.OrgID]; !running {
				orgCtx, cancel := context.WithCancel(ctx)
				active[u.OrgID] = cancel
				go w.runOrgWorker(orgCtx, u)
			}
		}
	}

	launch()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			launch()
		}
	}
}

func (w *Worker) runOrgWorker(ctx context.Context, user models.User) {
	log.Printf("worker[%s]: polling %s", user.OrgID, user.DLQSubscription)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		token, err := pubsub.GetRepairSAToken(ctx, w.RepairSA)
		if err != nil {
			log.Printf("worker[%s]: get token error: %v — retrying in 30s", user.OrgID, err)
			time.Sleep(30 * time.Second)
			continue
		}

		msgs, err := pubsub.PullMessages(ctx, token, user.DLQSubscription)
		if err != nil {
			log.Printf("worker[%s]: pull error: %v — retrying in 10s", user.OrgID, err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, msg := range msgs {
			if err := w.processMessage(ctx, user, token, msg); err != nil {
				log.Printf("worker[%s]: process message %s error: %v", user.OrgID, msg.Message.MessageID, err)
			}
		}

		if len(msgs) == 0 {
			time.Sleep(3 * time.Second)
		}
	}
}

// subscriptionShortName extracts a friendly label from a full resource path or bare name.
func subscriptionShortName(sub string) string {
	for i := len(sub) - 1; i >= 0; i-- {
		if sub[i] == '/' {
			return sub[i+1:]
		}
	}
	return sub
}

func allAutoRepublish(m map[string]bool) bool {
	if len(m) == 0 {
		return false
	}
	for _, v := range m {
		if !v {
			return false
		}
	}
	return true
}

func (w *Worker) processMessage(ctx context.Context, user models.User, token string, msg models.PubSubMessage) error {
	// Skip messages already repaired by us to prevent feedback loops.
	if msg.Message.Attributes["_deadlift_repaired"] == "true" {
		log.Printf("worker[%s]: skipping already-repaired message %s", user.OrgID, msg.Message.MessageID)
		_ = pubsub.AckMessages(ctx, token, user.DLQSubscription, []string{msg.AckID})
		return nil
	}

	rawBytes, err := base64.StdEncoding.DecodeString(msg.Message.Data)
	if err != nil {
		return err
	}
	rawPayload := string(rawBytes)

	fixedPayload, err := mcp.CallMCP(ctx, rawPayload)
	if err != nil {
		return err
	}

	taskID := uuid.New().String()
	task := models.Task{
		TaskID:       taskID,
		OrgID:        user.OrgID,
		MessageID:    msg.Message.MessageID,
		RawPayload:   rawPayload,
		Attributes:   msg.Message.Attributes,
		FixedPayload: fixedPayload,
		Status:       "pending_approval",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := w.Store.CreateTask(ctx, task); err != nil {
		return err
	}

	// Ack immediately — Firestore task is now the source of truth.
	if err := pubsub.AckMessages(ctx, token, user.DLQSubscription, []string{msg.AckID}); err != nil {
		log.Printf("worker[%s]: ack error: %v", user.OrgID, err)
	}

	// Batch detection: if threshold is set, group pending tasks from the same subscription.
	if user.BatchingThreshold > 0 {
		w.maybeAssignBatch(ctx, user, taskID)
	}

	if allAutoRepublish(user.AutoRepublish) {
		outAttrs := make(map[string]string, len(msg.Message.Attributes)+1)
		for k, v := range msg.Message.Attributes {
			outAttrs[k] = v
		}
		outAttrs["_deadlift_repaired"] = "true"
		if err := pubsub.PublishMessage(ctx, token, user.MainTopic, fixedPayload, outAttrs); err != nil {
			log.Printf("worker[%s]: auto-publish error: %v", user.OrgID, err)
			_ = w.Store.UpdateTaskStatus(ctx, taskID, "failed")
			return nil
		}
		_ = w.Store.UpdateTaskStatus(ctx, taskID, "approved")
		log.Printf("worker[%s]: auto-approved task %s", user.OrgID, taskID)
	} else {
		log.Printf("worker[%s]: task %s pending human approval", user.OrgID, taskID)
	}

	return nil
}

func (w *Worker) maybeAssignBatch(ctx context.Context, user models.User, taskID string) {
	sub := user.DLQSubscription
	existing, err := w.Store.GetPendingBatchBySubscription(ctx, user.OrgID, sub)
	if err != nil {
		log.Printf("worker[%s]: batch lookup error: %v", user.OrgID, err)
		return
	}

	if existing != nil {
		// Add to existing batch and stamp the task.
		if err := w.Store.AddTaskToBatch(ctx, existing.BatchID, taskID); err != nil {
			log.Printf("worker[%s]: add to batch error: %v", user.OrgID, err)
			return
		}
		if err := w.Store.UpdateTaskBatchID(ctx, taskID, existing.BatchID); err != nil {
			log.Printf("worker[%s]: stamp task batch_id error: %v", user.OrgID, err)
		}
		log.Printf("worker[%s]: task %s added to batch %s (now %d tasks)", user.OrgID, taskID, existing.BatchID, existing.TaskCount+1)
		return
	}

	// Count pending tasks for this subscription to decide whether to create a batch.
	pendingTasks, err := w.Store.ListTasksByOrg(ctx, user.OrgID)
	if err != nil {
		log.Printf("worker[%s]: list tasks for batch check error: %v", user.OrgID, err)
		return
	}
	pendingForSub := 0
	for _, t := range pendingTasks {
		if t.Status == "pending_approval" && t.BatchID == "" && (t.Attributes["dlq_subscription"] == sub || t.OrgID == user.OrgID) {
			pendingForSub++
		}
	}

	if pendingForSub < user.BatchingThreshold {
		return
	}

	// Threshold crossed — create a new batch and backfill existing unbatched pending tasks.
	batchID := uuid.New().String()
	topic := user.MainTopic
	batch := models.Batch{
		BatchID:      batchID,
		OrgID:        user.OrgID,
		Subscription: sub,
		Topic:        topic,
		TaskIDs:      []string{},
		TaskCount:    0,
		Status:       "pending",
		FirstSeen:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := w.Store.CreateBatch(ctx, batch); err != nil {
		log.Printf("worker[%s]: create batch error: %v", user.OrgID, err)
		return
	}

	count := 0
	for _, t := range pendingTasks {
		if t.Status == "pending_approval" && t.BatchID == "" {
			_ = w.Store.AddTaskToBatch(ctx, batchID, t.TaskID)
			_ = w.Store.UpdateTaskBatchID(ctx, t.TaskID, batchID)
			count++
		}
	}
	log.Printf("worker[%s]: created batch %s for subscription %s with %d tasks", user.OrgID, batchID, subscriptionShortName(sub), count)
}
