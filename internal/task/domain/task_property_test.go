package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// Feature: task-start-date, Property 1: 设置具体日期后应被持有
func TestProperty1_SetStartDate_AssignsDate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		task := NewTask("test", "", "owner", nil)
		d := rapid.Map(rapid.Int64Range(0, 4102444800), func(sec int64) time.Time {
			return time.Unix(sec, 0).UTC().Truncate(24 * time.Hour)
		}).Draw(t, "date")

		task.SetStartDate(&d)

		if task.StartDate == nil || !task.StartDate.Equal(d) {
			t.Fatalf("expected start_date=%v, got %v", d, task.StartDate)
		}
	})
}

// Feature: task-start-date, Property 2: 新建 Task 默认 start_date 为 nil（Inbox）
func TestProperty2_NewTask_Defaults(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		title := rapid.String().Draw(t, "title")
		notes := rapid.String().Draw(t, "notes")
		ownerID := rapid.String().Draw(t, "ownerID")

		numTags := rapid.IntRange(0, 5).Draw(t, "numTags")
		tagIDs := make([]uuid.UUID, numTags)
		for i := range tagIDs {
			tagIDs[i] = uuid.New()
		}

		task := NewTask(title, notes, ownerID, tagIDs)

		if task.StartDate != nil {
			t.Fatalf("expected StartDate=nil, got %v", task.StartDate)
		}
	})
}
