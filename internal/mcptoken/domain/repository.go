package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for MCP token persistence
type Repository interface {
	// Create creates a new MCP token
	Create(ctx context.Context, token *MCPToken) error

	// GetByToken retrieves an MCP token by its token value
	GetByToken(ctx context.Context, token uuid.UUID) (*MCPToken, error)

	// GetByID retrieves an MCP token by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*MCPToken, error)

	// ListByUserID retrieves all MCP tokens for a user
	ListByUserID(ctx context.Context, userID string) ([]*MCPToken, error)

	// UpdateLastUsedAt updates the last used timestamp
	UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error

	// Revoke revokes (deactivates) an MCP token
	Revoke(ctx context.Context, id uuid.UUID) error

	// Delete permanently deletes an MCP token
	Delete(ctx context.Context, id uuid.UUID) error
}
