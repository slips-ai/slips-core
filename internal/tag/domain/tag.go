package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a tag entity
type Tag struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTag creates a new tag
// Note: CreatedAt and UpdatedAt timestamps are not set here.
// They will be populated by the database on insertion (DEFAULT NOW()).
func NewTag(name string) *Tag {
	now := time.Now()
	return &Tag{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the tag
func (t *Tag) Update(name string) {
	t.Name = name
	t.UpdatedAt = time.Now()
}
