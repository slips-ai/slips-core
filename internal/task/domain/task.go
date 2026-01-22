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
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTask creates a new task
// Note: CreatedAt and UpdatedAt timestamps are not set here.
// They will be populated by the database on insertion (DEFAULT NOW()).
func NewTask(title, notes, ownerID string) *Task {
	return &Task{
		ID:      uuid.New(),
		Title:   title,
		Notes:   notes,
		OwnerID: ownerID,
	}
}

// Update updates the task
func (t *Task) Update(title, notes string) {
	t.Title = title
	t.Notes = notes
}
