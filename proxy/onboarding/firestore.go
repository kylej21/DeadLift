package main

import (
	"context"
	"fmt"
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
