package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	tagdomain "github.com/slips-ai/slips-core/internal/tag/domain"
	"github.com/slips-ai/slips-core/internal/task/domain"
	"github.com/slips-ai/slips-core/pkg/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("task-service")

// Service provides task business logic
type Service struct {
	repo    domain.Repository
	tagRepo tagdomain.Repository
	logger  *slog.Logger
}

// NewService creates a new task service
func NewService(repo domain.Repository, tagRepo tagdomain.Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:    repo,
		tagRepo: tagRepo,
		logger:  logger,
	}
}

// CreateTask creates a new task
func (s *Service) CreateTask(ctx context.Context, title, notes string, tagNames []string, startDateKind string, startDate *time.Time) (*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "CreateTask", trace.WithAttributes(
		attribute.String("title", title),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	// Convert tag names to tag IDs (create tags if they don't exist)
	tagIDs := make([]uuid.UUID, 0, len(tagNames))
	for _, tagName := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, tagName, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to get or create tag", "tag_name", tagName, "error", err)
			span.RecordError(err)
			return nil, err
		}
		tagIDs = append(tagIDs, tag.ID)
	}

	task := domain.NewTask(title, notes, userID, tagIDs)

	// Set start date if provided
	task.SetStartDate(domain.StartDateKind(startDateKind), startDate)
	if err := task.ValidateStartDate(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err := s.repo.Create(ctx, task); err != nil {
		s.logger.ErrorContext(ctx, "failed to create task", "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "task created", "id", task.ID, "owner_id", userID)
	return task, nil
}

// GetTask retrieves a task by ID
func (s *Service) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "GetTask", trace.WithAttributes(
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

	task, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get task", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	return task, nil
}

// UpdateTask updates a task
func (s *Service) UpdateTask(ctx context.Context, id uuid.UUID, title, notes string, tagNames []string, startDateKind *string, startDate *time.Time) (*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "UpdateTask", trace.WithAttributes(
		attribute.String("id", id.String()),
		attribute.String("title", title),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	task, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get task for update", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	// Convert tag names to tag IDs (create tags if they don't exist)
	tagIDs := make([]uuid.UUID, 0, len(tagNames))
	for _, tagName := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, tagName, userID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to get or create tag", "tag_name", tagName, "error", err)
			span.RecordError(err)
			return nil, err
		}
		tagIDs = append(tagIDs, tag.ID)
	}

	task.Update(title, notes, tagIDs)

	// Set start date only if provided (non-nil)
	if startDateKind != nil {
		task.SetStartDate(domain.StartDateKind(*startDateKind), startDate)
		if err := task.ValidateStartDate(); err != nil {
			span.RecordError(err)
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, task); err != nil {
		s.logger.ErrorContext(ctx, "failed to update task", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	// Clean up orphaned tags
	if err := s.tagRepo.DeleteOrphans(ctx, userID); err != nil {
		s.logger.WarnContext(ctx, "failed to clean up orphan tags", "error", err)
		// Don't fail the update if tag cleanup fails
	}

	s.logger.InfoContext(ctx, "task updated", "id", task.ID)
	return task, nil
}

// DeleteTask deletes a task
func (s *Service) DeleteTask(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "DeleteTask", trace.WithAttributes(
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
		s.logger.ErrorContext(ctx, "failed to delete task", "id", id, "error", err)
		span.RecordError(err)
		return err
	}

	// Clean up orphaned tags
	if err := s.tagRepo.DeleteOrphans(ctx, userID); err != nil {
		s.logger.WarnContext(ctx, "failed to clean up orphan tags", "error", err)
		// Don't fail the delete if tag cleanup fails
	}

	s.logger.InfoContext(ctx, "task deleted", "id", id)
	return nil
}

// ListTasks lists tasks
func (s *Service) ListTasks(ctx context.Context, filterTagIDs []uuid.UUID, limit, offset int, includeArchived, archivedOnly bool) ([]*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "ListTasks", trace.WithAttributes(
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
		attribute.Bool("include_archived", includeArchived),
		attribute.Bool("archived_only", archivedOnly),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	opts := domain.ListOptions{
		IncludeArchived: includeArchived,
		ArchivedOnly:    archivedOnly,
	}

	tasks, err := s.repo.List(ctx, userID, filterTagIDs, limit, offset, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list tasks", "error", err)
		span.RecordError(err)
		return nil, err
	}

	return tasks, nil
}

// ArchiveTask archives a task
func (s *Service) ArchiveTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "ArchiveTask", trace.WithAttributes(
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

	task, err := s.repo.Archive(ctx, id, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to archive task", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "task archived", "id", id)
	return task, nil
}

// UnarchiveTask unarchives a task
func (s *Service) UnarchiveTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "UnarchiveTask", trace.WithAttributes(
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

	task, err := s.repo.Unarchive(ctx, id, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to unarchive task", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "task unarchived", "id", id)
	return task, nil
}
