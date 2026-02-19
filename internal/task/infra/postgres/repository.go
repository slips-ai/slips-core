package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slips-ai/slips-core/internal/task/domain"
)

// TaskRepository implements domain.Repository using PostgreSQL
type TaskRepository struct {
	queries *Queries
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		queries: New(pool),
	}
}

// Create creates a new task
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	result, err := r.queries.CreateTask(ctx, CreateTaskParams{
		Title:     task.Title,
		Notes:     task.Notes,
		OwnerID:   task.OwnerID,
		StartDate: timeToPgDate(task.StartDate),
	})
	if err != nil {
		return err
	}

	taskID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return err
	}
	task.ID = taskID
	task.CreatedAt = result.CreatedAt.Time
	task.UpdatedAt = result.UpdatedAt.Time
	if result.ArchivedAt.Valid {
		task.ArchivedAt = &result.ArchivedAt.Time
	} else {
		task.ArchivedAt = nil
	}
	task.StartDate = pgDateToTime(result.StartDate)

	// Create task_tags associations
	for _, tagID := range task.TagIDs {
		pgTaskID := pgtype.UUID{
			Bytes: taskID,
			Valid: true,
		}
		pgTagID := pgtype.UUID{
			Bytes: tagID,
			Valid: true,
		}
		err := r.queries.CreateTaskTag(ctx, CreateTaskTagParams{
			TaskID: pgTaskID,
			TagID:  pgTagID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, id uuid.UUID, ownerID string) (*domain.Task, error) {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	result, err := r.queries.GetTask(ctx, GetTaskParams{
		ID:      pgID,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, err
	}

	taskID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	// Get task tag IDs
	pgTagIDs, err := r.queries.GetTaskTagIDs(ctx, pgID)
	if err != nil {
		return nil, err
	}

	tagIDs := make([]uuid.UUID, len(pgTagIDs))
	for i, pgTagID := range pgTagIDs {
		tagID, err := uuid.FromBytes(pgTagID.Bytes[:])
		if err != nil {
			return nil, err
		}
		tagIDs[i] = tagID
	}

	task := &domain.Task{
		ID:        taskID,
		Title:     result.Title,
		Notes:     result.Notes,
		TagIDs:    tagIDs,
		OwnerID:   result.OwnerID,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
		StartDate: pgDateToTime(result.StartDate),
	}
	checklistItems, err := r.ListChecklistItems(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	task.Checklist = checklistItems
	if result.ArchivedAt.Valid {
		task.ArchivedAt = &result.ArchivedAt.Time
	}
	return task, nil
}

// Update updates a task
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	pgID := pgtype.UUID{
		Bytes: task.ID,
		Valid: true,
	}

	result, err := r.queries.UpdateTask(ctx, UpdateTaskParams{
		ID:        pgID,
		Title:     task.Title,
		Notes:     task.Notes,
		OwnerID:   task.OwnerID,
		StartDate: timeToPgDate(task.StartDate),
	})
	if err != nil {
		return err
	}

	// Delete existing task_tags associations
	err = r.queries.DeleteTaskTags(ctx, pgID)
	if err != nil {
		return err
	}

	// Create new task_tags associations
	for _, tagID := range task.TagIDs {
		pgTagID := pgtype.UUID{
			Bytes: tagID,
			Valid: true,
		}
		err := r.queries.CreateTaskTag(ctx, CreateTaskTagParams{
			TaskID: pgID,
			TagID:  pgTagID,
		})
		if err != nil {
			return err
		}
	}

	task.UpdatedAt = result.UpdatedAt.Time
	return nil
}

// Delete deletes a task
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID, ownerID string) error {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
	return r.queries.DeleteTask(ctx, DeleteTaskParams{
		ID:      pgID,
		OwnerID: ownerID,
	})
}

// List lists tasks with pagination
func (r *TaskRepository) List(ctx context.Context, ownerID string, filterTagIDs []uuid.UUID, limit, offset int, opts domain.ListOptions) ([]*domain.Task, error) {
	// Validate parameters to prevent negative values and potential overflow
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	// Convert filterTagIDs to pgtype.UUID slice
	var pgFilterTagIDs []pgtype.UUID
	if len(filterTagIDs) > 0 {
		pgFilterTagIDs = make([]pgtype.UUID, len(filterTagIDs))
		for i, tagID := range filterTagIDs {
			pgFilterTagIDs[i] = pgtype.UUID{
				Bytes: tagID,
				Valid: true,
			}
		}
	}

	// Convert to int32 (validation is done at gRPC layer)
	results, err := r.queries.ListTasks(ctx, ListTasksParams{
		OwnerID:      ownerID,
		Limit:        int32(limit),
		Offset:       int32(offset),
		FilterTagIds: pgFilterTagIDs,
		IncludeArchived: pgtype.Bool{
			Bool:  opts.IncludeArchived,
			Valid: true,
		},
		ArchivedOnly: pgtype.Bool{
			Bool:  opts.ArchivedOnly,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	tasks := make([]*domain.Task, len(results))
	for i, result := range results {
		taskID, err := uuid.FromBytes(result.ID.Bytes[:])
		if err != nil {
			return nil, err
		}

		// Get task tag IDs
		pgTaskID := pgtype.UUID{
			Bytes: taskID,
			Valid: true,
		}
		pgTagIDs, err := r.queries.GetTaskTagIDs(ctx, pgTaskID)
		if err != nil {
			return nil, err
		}

		tagIDs := make([]uuid.UUID, len(pgTagIDs))
		for j, pgTagID := range pgTagIDs {
			tagID, err := uuid.FromBytes(pgTagID.Bytes[:])
			if err != nil {
				return nil, err
			}
			tagIDs[j] = tagID
		}

		task := &domain.Task{
			ID:        taskID,
			Title:     result.Title,
			Notes:     result.Notes,
			TagIDs:    tagIDs,
			OwnerID:   result.OwnerID,
			CreatedAt: result.CreatedAt.Time,
			UpdatedAt: result.UpdatedAt.Time,
			StartDate: pgDateToTime(result.StartDate),
		}
		if result.ArchivedAt.Valid {
			task.ArchivedAt = &result.ArchivedAt.Time
		}
		tasks[i] = task
	}

	return tasks, nil
}

// Archive archives a task by setting archived_at to current timestamp
func (r *TaskRepository) Archive(ctx context.Context, id uuid.UUID, ownerID string) (*domain.Task, error) {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	result, err := r.queries.ArchiveTask(ctx, ArchiveTaskParams{
		ID:      pgID,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, err
	}

	taskID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	// Get task tag IDs
	pgTagIDs, err := r.queries.GetTaskTagIDs(ctx, pgID)
	if err != nil {
		return nil, err
	}

	tagIDs := make([]uuid.UUID, len(pgTagIDs))
	for i, pgTagID := range pgTagIDs {
		tagID, err := uuid.FromBytes(pgTagID.Bytes[:])
		if err != nil {
			return nil, err
		}
		tagIDs[i] = tagID
	}

	task := &domain.Task{
		ID:        taskID,
		Title:     result.Title,
		Notes:     result.Notes,
		TagIDs:    tagIDs,
		OwnerID:   result.OwnerID,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
		StartDate: pgDateToTime(result.StartDate),
	}
	if result.ArchivedAt.Valid {
		task.ArchivedAt = &result.ArchivedAt.Time
	}
	return task, nil
}

// Unarchive unarchives a task by setting archived_at to NULL
func (r *TaskRepository) Unarchive(ctx context.Context, id uuid.UUID, ownerID string) (*domain.Task, error) {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	result, err := r.queries.UnarchiveTask(ctx, UnarchiveTaskParams{
		ID:      pgID,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, err
	}

	taskID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	// Get task tag IDs
	pgTagIDs, err := r.queries.GetTaskTagIDs(ctx, pgID)
	if err != nil {
		return nil, err
	}

	tagIDs := make([]uuid.UUID, len(pgTagIDs))
	for i, pgTagID := range pgTagIDs {
		tagID, err := uuid.FromBytes(pgTagID.Bytes[:])
		if err != nil {
			return nil, err
		}
		tagIDs[i] = tagID
	}

	task := &domain.Task{
		ID:        taskID,
		Title:     result.Title,
		Notes:     result.Notes,
		TagIDs:    tagIDs,
		OwnerID:   result.OwnerID,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
		StartDate: pgDateToTime(result.StartDate),
	}
	if result.ArchivedAt.Valid {
		task.ArchivedAt = &result.ArchivedAt.Time
	}
	return task, nil
}

// ListChecklistItems lists checklist items for a task.
func (r *TaskRepository) ListChecklistItems(ctx context.Context, taskID uuid.UUID, ownerID string) ([]domain.ChecklistItem, error) {
	pgTaskID := pgtype.UUID{Bytes: taskID, Valid: true}
	rows, err := r.queries.ListChecklistItems(ctx, ListChecklistItemsParams{
		TaskID:  pgTaskID,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.ChecklistItem, len(rows))
	for i := range rows {
		item, err := checklistItemFromDB(rows[i])
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	return items, nil
}

// AddChecklistItem creates a new checklist item for a task.
func (r *TaskRepository) AddChecklistItem(ctx context.Context, taskID uuid.UUID, ownerID, content string) (*domain.ChecklistItem, error) {
	row, err := r.queries.AddChecklistItem(ctx, AddChecklistItemParams{
		TaskID:  pgtype.UUID{Bytes: taskID, Valid: true},
		OwnerID: ownerID,
		Content: content,
	})
	if err != nil {
		return nil, err
	}

	item, err := checklistItemFromDB(row)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// UpdateChecklistItemContent updates checklist item text.
func (r *TaskRepository) UpdateChecklistItemContent(ctx context.Context, itemID uuid.UUID, ownerID, content string) (*domain.ChecklistItem, error) {
	row, err := r.queries.UpdateChecklistItemContent(ctx, UpdateChecklistItemContentParams{
		ItemID:  pgtype.UUID{Bytes: itemID, Valid: true},
		Content: content,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, err
	}

	item, err := checklistItemFromDB(row)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// SetChecklistItemCompleted sets checklist completion state.
func (r *TaskRepository) SetChecklistItemCompleted(ctx context.Context, itemID uuid.UUID, ownerID string, completed bool) (*domain.ChecklistItem, error) {
	row, err := r.queries.SetChecklistItemCompleted(ctx, SetChecklistItemCompletedParams{
		ItemID:    pgtype.UUID{Bytes: itemID, Valid: true},
		Completed: completed,
		OwnerID:   ownerID,
	})
	if err != nil {
		return nil, err
	}

	item, err := checklistItemFromDB(row)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// DeleteChecklistItem deletes a checklist item.
func (r *TaskRepository) DeleteChecklistItem(ctx context.Context, itemID uuid.UUID, ownerID string) error {
	rowsAffected, err := r.queries.DeleteChecklistItem(ctx, DeleteChecklistItemParams{
		ItemID:  pgtype.UUID{Bytes: itemID, Valid: true},
		OwnerID: ownerID,
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ReorderChecklistItems updates checklist item sort order.
func (r *TaskRepository) ReorderChecklistItems(ctx context.Context, taskID uuid.UUID, ownerID string, itemIDs []uuid.UUID) error {
	pgIDs := make([]pgtype.UUID, len(itemIDs))
	for i := range itemIDs {
		pgIDs[i] = pgtype.UUID{Bytes: itemIDs[i], Valid: true}
	}

	return r.queries.ReorderChecklistItems(ctx, ReorderChecklistItemsParams{
		TaskID:  pgtype.UUID{Bytes: taskID, Valid: true},
		ItemIds: pgIDs,
		OwnerID: ownerID,
	})
}

func checklistItemFromDB(row TaskChecklistItem) (domain.ChecklistItem, error) {
	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return domain.ChecklistItem{}, err
	}
	taskID, err := uuid.FromBytes(row.TaskID.Bytes[:])
	if err != nil {
		return domain.ChecklistItem{}, err
	}

	return domain.ChecklistItem{
		ID:        id,
		TaskID:    taskID,
		Content:   row.Content,
		Completed: row.Completed,
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// pgDateToTime converts a pgtype.Date to *time.Time.
// Returns nil if the date is not valid.
func pgDateToTime(d pgtype.Date) *time.Time {
	if d.Valid {
		t := d.Time
		return &t
	}
	return nil
}

// timeToPgDate converts a *time.Time to pgtype.Date.
// Returns an invalid pgtype.Date if the time is nil.
func timeToPgDate(t *time.Time) pgtype.Date {
	if t != nil {
		year, month, day := t.In(time.UTC).Date()
		normalized := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		return pgtype.Date{Time: normalized, Valid: true}
	}
	return pgtype.Date{Valid: false}
}
