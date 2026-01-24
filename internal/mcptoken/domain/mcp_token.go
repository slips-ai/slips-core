package domain

import (
	"time"

	"github.com/google/uuid"
)

// MCPToken represents an MCP authentication token
type MCPToken struct {
	ID         uuid.UUID
	Token      uuid.UUID
	UserID     string
	Name       string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	IsActive   bool
}

// IsExpired checks if the token has expired
func (t *MCPToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// IsValid checks if the token is valid (active and not expired)
func (t *MCPToken) IsValid() bool {
	return t.IsActive && !t.IsExpired()
}
