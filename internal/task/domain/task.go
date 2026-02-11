package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// StartDateKind represents the type of start date for a task
type StartDateKind string

const (
	StartDateKindInbox        StartDateKind = "inbox"
	StartDateKindSpecificDate StartDateKind = "specific_date"
)

// IsValid returns true if the StartDateKind is a valid value
func (k StartDateKind) IsValid() bool {
	return k == StartDateKindInbox || k == StartDateKindSpecificDate
}

// Task represents a task entity
type Task struct {
	ID            uuid.UUID
	Title         string
	Notes         string
	TagIDs        []uuid.UUID
	OwnerID       string
	ArchivedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartDateKind StartDateKind
	StartDate     *time.Time

}

// NewTask creates a new task
// Note: CreatedAt and UpdatedAt timestamps are not set here.
// They will be populated by the database on insertion (DEFAULT NOW()).
func NewTask(title, notes, ownerID string, tagIDs []uuid.UUID) *Task {
	return &Task{
		ID:            uuid.New(),
		Title:         title,
		Notes:         notes,
		TagIDs:        tagIDs,
		OwnerID:       ownerID,
		ArchivedAt:    nil,
		StartDateKind: StartDateKindInbox,
		StartDate:     nil,
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
// SetStartDate sets the start date kind and date for the task.
// When kind is specific_date, the provided date is used.
// When kind is inbox, the date is forced to nil.
func (t *Task) SetStartDate(kind StartDateKind, date *time.Time) {
	t.StartDateKind = kind
	if kind == StartDateKindSpecificDate {
		t.StartDate = date
	} else {
		t.StartDate = nil
	}
}

// ValidateStartDate validates the consistency of StartDateKind and StartDate.
func (t *Task) ValidateStartDate() error {
	if !t.StartDateKind.IsValid() {
		return errors.New("invalid start_date_kind: must be 'inbox' or 'specific_date'")
	}
	if t.StartDateKind == StartDateKindSpecificDate && t.StartDate == nil {
		return errors.New("start_date is required when start_date_kind is specific_date")
	}
	if t.StartDateKind == StartDateKindInbox && t.StartDate != nil {
		return errors.New("start_date must be nil when start_date_kind is inbox")
	}
	return nil
}
