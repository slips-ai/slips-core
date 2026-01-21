package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for tag persistence
type Repository interface {
	Create(ctx context.Context, tag *Tag) error
	Get(ctx context.Context, id uuid.UUID) (*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Tag, error)
}
