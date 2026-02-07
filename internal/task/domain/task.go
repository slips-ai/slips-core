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
	TagIDs    []uuid.UUID
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTask creates a new task
// Note: CreatedAt and UpdatedAt timestamps are not set here.
// They will be populated by the database on insertion (DEFAULT NOW()).
func NewTask(title, notes, ownerID string, tagIDs []uuid.UUID) *Task {
	return &Task{
		ID:      uuid.New(),
		Title:   title,
		Notes:   notes,
		TagIDs:  tagIDs,
		OwnerID: ownerID,
	}
}

// Update updates the task
func (t *Task) Update(title, notes string, tagIDs []uuid.UUID) {
	t.Title = title
	t.Notes = notes
	t.TagIDs = tagIDs
}
