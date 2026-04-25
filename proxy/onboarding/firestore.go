package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

func createUser(ctx context.Context, user User) error {
	_, err := fsClient.Collection("users").Doc(user.OrgID).Set(ctx, user)
	return err
}

func getUserByOrgID(ctx context.Context, orgID string) (*User, error) {
	doc, err := fsClient.Collection("users").Doc(orgID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	var user User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	return &user, nil
}

func getUserByGoogleSub(ctx context.Context, sub string) (*User, error) {
	docs, err := fsClient.Collection("users").Where("google_sub", "==", sub).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("get user by sub: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	var user User
	if err := docs[0].DataTo(&user); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	return &user, nil
}

func getAllUsers(ctx context.Context) ([]User, error) {
	docs, err := fsClient.Collection("users").Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]User, 0, len(docs))
	for _, doc := range docs {
		var u User
		if err := doc.DataTo(&u); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

func createTask(ctx context.Context, task Task) error {
	_, err := fsClient.Collection("tasks").Doc(task.TaskID).Set(ctx, task)
	return err
}

func getTask(ctx context.Context, taskID string) (*Task, error) {
	doc, err := fsClient.Collection("tasks").Doc(taskID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	var task Task
	if err := doc.DataTo(&task); err != nil {
		return nil, fmt.Errorf("parse task: %w", err)
	}
	return &task, nil
}

func updateTaskStatus(ctx context.Context, taskID, status string) error {
	_, err := fsClient.Collection("tasks").Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

func listTasksByOrg(ctx context.Context, orgID string) ([]Task, error) {
	docs, err := fsClient.Collection("tasks").Where("org_id", "==", orgID).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	tasks := make([]Task, 0, len(docs))
	for _, doc := range docs {
		var t Task
		if err := doc.DataTo(&t); err == nil {
			tasks = append(tasks, t)
		}
	}
	// Sort newest first in Go to avoid needing a composite Firestore index.
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[j].CreatedAt.After(tasks[i].CreatedAt) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
	return tasks, nil
}
