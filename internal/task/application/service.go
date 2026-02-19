package application

import (
	"context"
	"log/slog"
	"slices"
	"strings"
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
func (s *Service) CreateTask(ctx context.Context, title, notes string, tagNames []string, startDate *time.Time, checklistItems []string) (*domain.Task, error) {
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
	task.Checklist = make([]domain.ChecklistItem, 0, len(checklistItems))
	for i, content := range checklistItems {
		task.Checklist = append(task.Checklist, domain.ChecklistItem{
			Content:   content,
			Completed: false,
			SortOrder: int32(i),
		})
	}

	// Set start date if provided; nil means inbox
	task.SetStartDate(startDate)

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
func (s *Service) UpdateTask(ctx context.Context, id uuid.UUID, title, notes string, tagNames []string, startDateProvided bool, startDate *time.Time) (*domain.Task, error) {
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

	// Update start date only when provided in request.
	if startDateProvided {
		task.SetStartDate(startDate)
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

// AddChecklistItem adds a checklist item to a task.
func (s *Service) AddChecklistItem(ctx context.Context, taskID uuid.UUID, content string) (*domain.ChecklistItem, error) {
	ctx, span := tracer.Start(ctx, "AddChecklistItem", trace.WithAttributes(
		attribute.String("task_id", taskID.String()),
	))
	defer span.End()

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	item, err := s.repo.AddChecklistItem(ctx, taskID, userID, content)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to add checklist item", "task_id", taskID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	return item, nil
}

// UpdateChecklistItemContent updates checklist item text.
func (s *Service) UpdateChecklistItemContent(ctx context.Context, itemID uuid.UUID, content string) (*domain.ChecklistItem, error) {
	ctx, span := tracer.Start(ctx, "UpdateChecklistItemContent", trace.WithAttributes(
		attribute.String("item_id", itemID.String()),
	))
	defer span.End()

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	item, err := s.repo.UpdateChecklistItemContent(ctx, itemID, userID, content)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update checklist item", "item_id", itemID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	return item, nil
}

// SetChecklistItemCompleted sets checklist item completion state.
func (s *Service) SetChecklistItemCompleted(ctx context.Context, itemID uuid.UUID, completed bool) (*domain.ChecklistItem, error) {
	ctx, span := tracer.Start(ctx, "SetChecklistItemCompleted", trace.WithAttributes(
		attribute.String("item_id", itemID.String()),
		attribute.Bool("completed", completed),
	))
	defer span.End()

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	item, err := s.repo.SetChecklistItemCompleted(ctx, itemID, userID, completed)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set checklist item completion", "item_id", itemID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	return item, nil
}

// DeleteChecklistItem deletes a checklist item.
func (s *Service) DeleteChecklistItem(ctx context.Context, itemID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "DeleteChecklistItem", trace.WithAttributes(
		attribute.String("item_id", itemID.String()),
	))
	defer span.End()

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return err
	}

	if err := s.repo.DeleteChecklistItem(ctx, itemID, userID); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete checklist item", "item_id", itemID, "error", err)
		span.RecordError(err)
		return err
	}

	return nil
}

// ReorderChecklistItems sets a new checklist order for all task items.
func (s *Service) ReorderChecklistItems(ctx context.Context, taskID uuid.UUID, itemIDs []uuid.UUID) ([]domain.ChecklistItem, error) {
	ctx, span := tracer.Start(ctx, "ReorderChecklistItems", trace.WithAttributes(
		attribute.String("task_id", taskID.String()),
		attribute.Int("item_count", len(itemIDs)),
	))
	defer span.End()

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	existingItems, err := s.repo.ListChecklistItems(ctx, taskID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list checklist items", "task_id", taskID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	if len(existingItems) != len(itemIDs) {
		return nil, domain.ErrInvalidChecklistOrder
	}

	existingIDs := make([]uuid.UUID, len(existingItems))
	for i := range existingItems {
		existingIDs[i] = existingItems[i].ID
	}

	slices.SortFunc(existingIDs, func(a, b uuid.UUID) int {
		return strings.Compare(a.String(), b.String())
	})
	sortedRequested := append([]uuid.UUID(nil), itemIDs...)
	slices.SortFunc(sortedRequested, func(a, b uuid.UUID) int {
		return strings.Compare(a.String(), b.String())
	})
	if !slices.Equal(existingIDs, sortedRequested) {
		return nil, domain.ErrInvalidChecklistOrder
	}

	if err := s.repo.ReorderChecklistItems(ctx, taskID, userID, itemIDs); err != nil {
		s.logger.ErrorContext(ctx, "failed to reorder checklist items", "task_id", taskID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	items, err := s.repo.ListChecklistItems(ctx, taskID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list reordered checklist items", "task_id", taskID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	return items, nil
}
