package domain

import (
	"context"
)

// Repository defines the interface for user persistence
type Repository interface {
	// UpsertUser creates or updates a user
	// Only updates username and avatar_url if they are currently NULL
	UpsertUser(ctx context.Context, user *User) (*User, error)

	// GetUserByUserID retrieves a user by their user ID (from JWT claims)
	GetUserByUserID(ctx context.Context, userID string) (*User, error)

	// GetUserByID retrieves a user by their database ID
	GetUserByID(ctx context.Context, id int64) (*User, error)

	// UpdateUserTavilyMCPToken updates Tavily MCP token for the given user ID
	UpdateUserTavilyMCPToken(ctx context.Context, userID, tavilyMCPToken string) (*User, error)
}
