package auth

import (
	"context"
	"errors"
)

type contextKey string

const userIDKey contextKey = "user_id"

var (
	// ErrMissingUserID is returned when user ID is not found in context
	ErrMissingUserID = errors.New("user ID not found in context")
)

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", ErrMissingUserID
	}
	return userID, nil
}
