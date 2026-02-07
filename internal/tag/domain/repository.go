package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for tag persistence
type Repository interface {
	Create(ctx context.Context, tag *Tag) error
	Get(ctx context.Context, id uuid.UUID, ownerID string) (*Tag, error)
	GetByName(ctx context.Context, name, ownerID string) (*Tag, error)
	GetOrCreate(ctx context.Context, name, ownerID string) (*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uuid.UUID, ownerID string) error
	DeleteOrphans(ctx context.Context, ownerID string) error
	List(ctx context.Context, ownerID string, limit, offset int) ([]*Tag, error)
}
