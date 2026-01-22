package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/slips-ai/slips-core/internal/task/domain"
	"github.com/slips-ai/slips-core/pkg/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("task-service")

// Service provides task business logic
type Service struct {
	repo   domain.Repository
	logger *slog.Logger
}

// NewService creates a new task service
func NewService(repo domain.Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateTask creates a new task
func (s *Service) CreateTask(ctx context.Context, title, notes string) (*domain.Task, error) {
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

	task := domain.NewTask(title, notes, userID)
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
func (s *Service) UpdateTask(ctx context.Context, id uuid.UUID, title, notes string) (*domain.Task, error) {
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

	task.Update(title, notes)
	if err := s.repo.Update(ctx, task); err != nil {
		s.logger.ErrorContext(ctx, "failed to update task", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
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

	s.logger.InfoContext(ctx, "task deleted", "id", id)
	return nil
}

// ListTasks lists tasks
func (s *Service) ListTasks(ctx context.Context, limit, offset int) ([]*domain.Task, error) {
	ctx, span := tracer.Start(ctx, "ListTasks", trace.WithAttributes(
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

	tasks, err := s.repo.List(ctx, userID, limit, offset)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list tasks", "error", err)
		span.RecordError(err)
		return nil, err
	}

	return tasks, nil
}
