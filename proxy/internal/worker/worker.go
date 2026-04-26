package worker

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"proxy/internal/mcp"
	"proxy/internal/models"
	"proxy/internal/pubsub"
	"proxy/internal/store"

	"github.com/google/uuid"
)

// Worker polls each org's DLQ subscription and creates repair tasks.
type Worker struct {
	RepairSA  string
	Store     *store.Store
	MCPClient *mcp.Client
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
		current := map[string]bool{}
		for _, u := range users {
			current[u.OrgID] = true
		}
		// Cancel goroutines for orgs no longer in Firestore.
		for orgID, cancel := range active {
			if !current[orgID] {
				log.Printf("worker: org %s removed from Firestore, stopping goroutine", orgID)
				cancel()
				delete(active, orgID)
			}
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
	// Skip messages already repaired by us to prevent feedback loops. redeploy trigger
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

	result, err := w.MCPClient.Call(ctx, user.OrgID, msg.Message.MessageID, rawPayload)
	if err != nil {
		return err
	}

	taskID := uuid.New().String()
	attrs := make(map[string]string, len(msg.Message.Attributes)+1)
	for k, v := range msg.Message.Attributes {
		attrs[k] = v
	}
	attrs["_deadlift_confidence"] = fmt.Sprintf("%d", result.ConfidenceScore)

	task := models.Task{
		TaskID:       taskID,
		OrgID:        user.OrgID,
		MessageID:    msg.Message.MessageID,
		RawPayload:   rawPayload,
		Attributes:   attrs,
		FixedPayload: result.FixedPayload,
		ErrorClass:   result.ErrorClass,
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

	if allAutoRepublish(user.AutoRepublish) {
		outAttrs := make(map[string]string, len(msg.Message.Attributes)+1)
		for k, v := range msg.Message.Attributes {
			if k != "_deadlift_confidence" && k != "simulate_failure" {
				outAttrs[k] = v
			}
		}
		outAttrs["_deadlift_repaired"] = "true"
		if err := pubsub.PublishMessage(ctx, token, user.MainTopic, result.FixedPayload, outAttrs); err != nil {
			log.Printf("worker[%s]: auto-publish error: %v", user.OrgID, err)
			_ = w.Store.UpdateTaskStatus(ctx, taskID, "failed")
			return nil
		}
		_ = w.Store.UpdateTaskStatus(ctx, taskID, "approved")
		log.Printf("worker[%s]: auto-approved task %s (class: %s)", user.OrgID, taskID, result.ErrorClass)
	} else {
		log.Printf("worker[%s]: task %s pending approval (class: %s)", user.OrgID, taskID, result.ErrorClass)
	}

	return nil
}
