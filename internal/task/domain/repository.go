package domain

import (
	"context"

	"github.com/google/uuid"
)

// ListOptions defines options for listing tasks
type ListOptions struct {
	IncludeArchived bool
	ArchivedOnly    bool
}

// Repository defines the interface for task persistence
type Repository interface {
	Create(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID, ownerID string) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID, ownerID string) error
	List(ctx context.Context, ownerID string, filterTagIDs []uuid.UUID, limit, offset int, opts ListOptions) ([]*Task, error)
	Archive(ctx context.Context, id uuid.UUID, ownerID string) (*Task, error)
	Unarchive(ctx context.Context, id uuid.UUID, ownerID string) (*Task, error)
}
