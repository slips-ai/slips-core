package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for task persistence
type Repository interface {
	Create(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID, ownerID string) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID, ownerID string) error
	List(ctx context.Context, ownerID string, filterTagIDs []uuid.UUID, limit, offset int) ([]*Task, error)
}
