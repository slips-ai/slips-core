package domain

import (
	"testing"
	"time"
)

func TestSetStartDate_SetsSpecificDate(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	task.SetStartDate(&d)

	if task.StartDate == nil || !task.StartDate.Equal(d) {
		t.Fatalf("expected date=%v, got %v", d, task.StartDate)
	}
}

func TestSetStartDate_ClearsToInboxWhenNil(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	task.SetStartDate(nil)

	if task.StartDate != nil {
		t.Fatalf("expected date=nil, got %v", task.StartDate)
	}
}

func TestSetStartDate_SwitchFromDateToInbox(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	task.SetStartDate(&d)
	task.SetStartDate(nil)

	if task.StartDate != nil {
		t.Fatalf("expected date=nil after clearing, got %v", task.StartDate)
	}
}
