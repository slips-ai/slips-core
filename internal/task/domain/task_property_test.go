package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// Feature: task-start-date, Property 1: StartDateKind 与 StartDate 一致性不变量
// For any Task set via SetStartDate(), ValidateStartDate() returns nil.
// **Validates: Requirements 1.4, 1.5**
func TestProperty1_SetStartDate_Consistency(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		task := NewTask("test", "", "owner", nil)

		kind := rapid.SampledFrom([]StartDateKind{
			StartDateKindInbox,
			StartDateKindSpecificDate,
		}).Draw(t, "kind")

		var date *time.Time
		if kind == StartDateKindSpecificDate {
			d := rapid.Map(rapid.Int64Range(0, 4102444800), func(sec int64) time.Time {
				return time.Unix(sec, 0).UTC().Truncate(24 * time.Hour)
			}).Draw(t, "date")
			date = &d
		}

		task.SetStartDate(kind, date)

		if err := task.ValidateStartDate(); err != nil {
			t.Fatalf("ValidateStartDate() returned error after SetStartDate(%q, %v): %v", kind, date, err)
		}
	})
}

// Feature: task-start-date, Property 2: 新建 Task 默认值
// For any Task created via NewTask(), StartDateKind is inbox and StartDate is nil.
// **Validates: Requirements 1.3**
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

		if task.StartDateKind != StartDateKindInbox {
			t.Fatalf("expected StartDateKind=%q, got %q", StartDateKindInbox, task.StartDateKind)
		}
		if task.StartDate != nil {
			t.Fatalf("expected StartDate=nil, got %v", task.StartDate)
		}
	})
}
