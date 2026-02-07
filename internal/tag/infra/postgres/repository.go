package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slips-ai/slips-core/internal/tag/domain"
)

// TagRepository implements domain.Repository using PostgreSQL
type TagRepository struct {
	queries *Queries
}

// NewTagRepository creates a new tag repository
func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{
		queries: New(pool),
	}
}

// Create creates a new tag
func (r *TagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	result, err := r.queries.CreateTag(ctx, CreateTagParams{
		Name:    tag.Name,
		OwnerID: tag.OwnerID,
	})
	if err != nil {
		return err
	}

	tagID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return err
	}
	tag.ID = tagID
	tag.CreatedAt = result.CreatedAt.Time
	tag.UpdatedAt = result.UpdatedAt.Time
	return nil
}

// Get retrieves a tag by ID
func (r *TagRepository) Get(ctx context.Context, id uuid.UUID, ownerID string) (*domain.Tag, error) {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	result, err := r.queries.GetTag(ctx, GetTagParams{
		ID:      pgID,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, err
	}

	tagID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	return &domain.Tag{
		ID:        tagID,
		Name:      result.Name,
		OwnerID:   result.OwnerID,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

// Update updates a tag
func (r *TagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	pgID := pgtype.UUID{
		Bytes: tag.ID,
		Valid: true,
	}

	result, err := r.queries.UpdateTag(ctx, UpdateTagParams{
		ID:      pgID,
		Name:    tag.Name,
		OwnerID: tag.OwnerID,
	})
	if err != nil {
		return err
	}

	tag.UpdatedAt = result.UpdatedAt.Time
	return nil
}

// Delete deletes a tag
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID, ownerID string) error {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
	return r.queries.DeleteTag(ctx, DeleteTagParams{
		ID:      pgID,
		OwnerID: ownerID,
	})
}

// List lists tags with pagination
func (r *TagRepository) List(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Tag, error) {
	// Validate parameters to prevent negative values and potential overflow
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	// Convert to int32 (validation is done at gRPC layer)
	results, err := r.queries.ListTags(ctx, ListTagsParams{
		OwnerID: ownerID,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}

	tags := make([]*domain.Tag, len(results))
	for i, result := range results {
		tagID, err := uuid.FromBytes(result.ID.Bytes[:])
		if err != nil {
			return nil, err
		}
		tags[i] = &domain.Tag{
			ID:        tagID,
			Name:      result.Name,
			OwnerID:   result.OwnerID,
			CreatedAt: result.CreatedAt.Time,
			UpdatedAt: result.UpdatedAt.Time,
		}
	}

	return tags, nil
}
