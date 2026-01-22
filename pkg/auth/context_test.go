package auth

import (
	"context"
	"testing"
)

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-123"

	ctx = WithUserID(ctx, userID)

	extractedID, err := GetUserID(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if extractedID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, extractedID)
	}
}

func TestGetUserID_MissingUserID(t *testing.T) {
	ctx := context.Background()

	_, err := GetUserID(ctx)
	if err != ErrMissingUserID {
		t.Fatalf("expected ErrMissingUserID, got %v", err)
	}
}

func TestGetUserID_EmptyUserID(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "")

	_, err := GetUserID(ctx)
	if err != ErrMissingUserID {
		t.Fatalf("expected ErrMissingUserID for empty user ID, got %v", err)
	}
}
