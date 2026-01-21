package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for task persistence
type Repository interface {
	Create(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Task, error)
}
