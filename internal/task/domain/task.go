package domain

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a task entity
type Task struct {
	ID        uuid.UUID
	Title     string
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTask creates a new task
func NewTask(title, notes string) *Task {
	now := time.Now()
	return &Task{
		ID:        uuid.New(),
		Title:     title,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the task
func (t *Task) Update(title, notes string) {
	t.Title = title
	t.Notes = notes
	t.UpdatedAt = time.Now()
}
