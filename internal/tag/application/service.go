package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/slips-ai/slips-core/internal/tag/domain"
	"github.com/slips-ai/slips-core/pkg/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("tag-service")

// Service provides tag business logic
type Service struct {
	repo   domain.Repository
	logger *slog.Logger
}

// NewService creates a new tag service
func NewService(repo domain.Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateTag creates a new tag
func (s *Service) CreateTag(ctx context.Context, name string) (*domain.Tag, error) {
	ctx, span := tracer.Start(ctx, "CreateTag", trace.WithAttributes(
		attribute.String("name", name),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	tag := domain.NewTag(name, userID)
	if err := s.repo.Create(ctx, tag); err != nil {
		s.logger.ErrorContext(ctx, "failed to create tag", "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "tag created", "id", tag.ID, "owner_id", userID)
	return tag, nil
}

// GetTag retrieves a tag by ID
func (s *Service) GetTag(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	ctx, span := tracer.Start(ctx, "GetTag", trace.WithAttributes(
		attribute.String("id", id.String()),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	tag, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tag", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	return tag, nil
}

// UpdateTag updates a tag
func (s *Service) UpdateTag(ctx context.Context, id uuid.UUID, name string) (*domain.Tag, error) {
	ctx, span := tracer.Start(ctx, "UpdateTag", trace.WithAttributes(
		attribute.String("id", id.String()),
		attribute.String("name", name),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	tag, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tag for update", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	tag.Update(name)
	if err := s.repo.Update(ctx, tag); err != nil {
		s.logger.ErrorContext(ctx, "failed to update tag", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "tag updated", "id", tag.ID)
	return tag, nil
}

// DeleteTag deletes a tag
func (s *Service) DeleteTag(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "DeleteTag", trace.WithAttributes(
		attribute.String("id", id.String()),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return err
	}

	if err := s.repo.Delete(ctx, id, userID); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete tag", "id", id, "error", err)
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "tag deleted", "id", id)
	return nil
}

// ListTags lists tags
func (s *Service) ListTags(ctx context.Context, limit, offset int) ([]*domain.Tag, error) {
	ctx, span := tracer.Start(ctx, "ListTags", trace.WithAttributes(
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	tags, err := s.repo.List(ctx, userID, limit, offset)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list tags", "error", err)
		span.RecordError(err)
		return nil, err
	}

	return tags, nil
}
