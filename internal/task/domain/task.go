package domain

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a task entity
type Task struct {
	ID         uuid.UUID
	Title      string
	Notes      string
	TagIDs     []uuid.UUID
	Checklist  []ChecklistItem
	OwnerID    string
	ArchivedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	StartDate  *time.Time
}

// ChecklistItem represents a single checklist row for a task.
type ChecklistItem struct {
	ID        uuid.UUID
	TaskID    uuid.UUID
	Content   string
	Completed bool
	SortOrder int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTask creates a new task
// Note: CreatedAt and UpdatedAt timestamps are not set here.
// They will be populated by the database on insertion (DEFAULT NOW()).
func NewTask(title, notes, ownerID string, tagIDs []uuid.UUID) *Task {
	return &Task{
		ID:         uuid.New(),
		Title:      title,
		Notes:      notes,
		TagIDs:     tagIDs,
		OwnerID:    ownerID,
		ArchivedAt: nil,
		StartDate:  nil,
	}
}

// Update updates the task
func (t *Task) Update(title, notes string, tagIDs []uuid.UUID) {
	t.Title = title
	t.Notes = notes
	t.TagIDs = tagIDs
}

// Archive marks the task as archived with the current timestamp
func (t *Task) Archive() {
	now := time.Now()
	t.ArchivedAt = &now
}

// Unarchive marks the task as active by clearing the archived timestamp
func (t *Task) Unarchive() {
	t.ArchivedAt = nil
}

// IsArchived returns true if the task is archived
func (t *Task) IsArchived() bool {
	return t.ArchivedAt != nil
}

// SetStartDate sets or clears the start date for the task.
// A nil date means the task belongs to inbox.
func (t *Task) SetStartDate(date *time.Time) {
	t.StartDate = date
}
