package postgres

import (
	"context"

	"github.com/google/uuid"
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
		Title:   task.Title,
		Notes:   task.Notes,
		OwnerID: task.OwnerID,
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

	return &domain.Task{
		ID:        taskID,
		Title:     result.Title,
		Notes:     result.Notes,
		OwnerID:   result.OwnerID,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

// Update updates a task
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	pgID := pgtype.UUID{
		Bytes: task.ID,
		Valid: true,
	}

	result, err := r.queries.UpdateTask(ctx, UpdateTaskParams{
		ID:      pgID,
		Title:   task.Title,
		Notes:   task.Notes,
		OwnerID: task.OwnerID,
	})
	if err != nil {
		return err
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
func (r *TaskRepository) List(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Task, error) {
	// Validate parameters to prevent negative values and potential overflow
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	// Convert to int32 (validation is done at gRPC layer)
	results, err := r.queries.ListTasks(ctx, ListTasksParams{
		OwnerID: ownerID,
		Limit:   int32(limit),
		Offset:  int32(offset),
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
		tasks[i] = &domain.Task{
			ID:        taskID,
			Title:     result.Title,
			Notes:     result.Notes,
			OwnerID:   result.OwnerID,
			CreatedAt: result.CreatedAt.Time,
			UpdatedAt: result.UpdatedAt.Time,
		}
	}

	return tasks, nil
}
