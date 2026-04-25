package main

import (
	"context"
	"encoding/base64"
	"log"
	"time"

	"github.com/google/uuid"
)

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

func startWorker(ctx context.Context) {
	log.Println("worker: starting DLQ polling")
	// Refresh orgs every 5 minutes to pick up new onboards without restarting.
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	active := map[string]context.CancelFunc{}

	launch := func() {
		users, err := getAllUsers(ctx)
		if err != nil {
			log.Printf("worker: failed to load orgs: %v", err)
			return
		}
		for _, u := range users {
			if _, running := active[u.OrgID]; !running {
				orgCtx, cancel := context.WithCancel(ctx)
				active[u.OrgID] = cancel
				go runOrgWorker(orgCtx, u)
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

func runOrgWorker(ctx context.Context, user User) {
	log.Printf("worker[%s]: polling %s", user.OrgID, user.DLQSubscription)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		token, err := getRepairSAToken(ctx)
		if err != nil {
			log.Printf("worker[%s]: get token error: %v — retrying in 30s", user.OrgID, err)
			time.Sleep(30 * time.Second)
			continue
		}

		msgs, err := pullMessages(ctx, token, user.DLQSubscription)
		if err != nil {
			log.Printf("worker[%s]: pull error: %v — retrying in 10s", user.OrgID, err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, msg := range msgs {
			if err := processMessage(ctx, user, token, msg); err != nil {
				log.Printf("worker[%s]: process message %s error: %v", user.OrgID, msg.Message.MessageID, err)
			}
		}

		if len(msgs) == 0 {
			time.Sleep(3 * time.Second)
		}
	}
}

func processMessage(ctx context.Context, user User, token string, msg pubsubMessage) error {
	// Skip messages we already repaired — prevents feedback loops.
	if msg.Message.Attributes["_deadlift_repaired"] == "true" {
		log.Printf("worker[%s]: skipping already-repaired message %s", user.OrgID, msg.Message.MessageID)
		_ = ackMessages(ctx, token, user.DLQSubscription, []string{msg.AckID})
		return nil
	}

	rawBytes, err := base64.StdEncoding.DecodeString(msg.Message.Data)
	if err != nil {
		return err
	}
	rawPayload := string(rawBytes)

	fixedPayload, err := callMCP(ctx, rawPayload)
	if err != nil {
		return err
	}

	taskID := uuid.New().String()
	task := Task{
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

	if err := createTask(ctx, task); err != nil {
		return err
	}

	// Ack immediately — Firestore task is now the source of truth.
	if err := ackMessages(ctx, token, user.DLQSubscription, []string{msg.AckID}); err != nil {
		log.Printf("worker[%s]: ack error: %v", user.OrgID, err)
	}

	if allAutoRepublish(user.AutoRepublish) {
		outAttrs := make(map[string]string, len(msg.Message.Attributes)+1)
		for k, v := range msg.Message.Attributes {
			outAttrs[k] = v
		}
		outAttrs["_deadlift_repaired"] = "true"
		if err := publishMessage(ctx, token, user.MainTopic, fixedPayload, outAttrs); err != nil {
			log.Printf("worker[%s]: auto-publish error: %v", user.OrgID, err)
			_ = updateTaskStatus(ctx, taskID, "failed")
			return nil
		}
		_ = updateTaskStatus(ctx, taskID, "approved")
		log.Printf("worker[%s]: auto-approved task %s", user.OrgID, taskID)
	} else {
		log.Printf("worker[%s]: task %s pending human approval", user.OrgID, taskID)
	}

	return nil
}
