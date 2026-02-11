package domain

import (
	"testing"
	"time"
)

// --- StartDateKind.IsValid() ---

func TestStartDateKind_IsValid_Inbox(t *testing.T) {
	if !StartDateKindInbox.IsValid() {
		t.Fatal("expected inbox to be valid")
	}
}

func TestStartDateKind_IsValid_SpecificDate(t *testing.T) {
	if !StartDateKindSpecificDate.IsValid() {
		t.Fatal("expected specific_date to be valid")
	}
}

func TestStartDateKind_IsValid_Empty(t *testing.T) {
	if StartDateKind("").IsValid() {
		t.Fatal("expected empty string to be invalid")
	}
}

func TestStartDateKind_IsValid_Unknown(t *testing.T) {
	invalid := []StartDateKind{"unknown", "INBOX", "Specific_Date", "scheduled", " inbox"}
	for _, k := range invalid {
		if k.IsValid() {
			t.Fatalf("expected %q to be invalid", k)
		}
	}
}

// --- SetStartDate ---

func TestSetStartDate_SpecificDate_SetsDateAndKind(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	task.SetStartDate(StartDateKindSpecificDate, &d)

	if task.StartDateKind != StartDateKindSpecificDate {
		t.Fatalf("expected kind=specific_date, got %q", task.StartDateKind)
	}
	if task.StartDate == nil || !task.StartDate.Equal(d) {
		t.Fatalf("expected date=%v, got %v", d, task.StartDate)
	}
}

func TestSetStartDate_Inbox_NilDate(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	task.SetStartDate(StartDateKindInbox, nil)

	if task.StartDateKind != StartDateKindInbox {
		t.Fatalf("expected kind=inbox, got %q", task.StartDateKind)
	}
	if task.StartDate != nil {
		t.Fatalf("expected date=nil, got %v", task.StartDate)
	}
}

func TestSetStartDate_Inbox_ForcesDateToNil(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Even if a date is passed with inbox, it should be forced to nil
	task.SetStartDate(StartDateKindInbox, &d)

	if task.StartDate != nil {
		t.Fatalf("expected date=nil when kind=inbox, got %v", task.StartDate)
	}
}

func TestSetStartDate_SpecificDate_NilDate_SetsNil(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	// Setting specific_date with nil date — SetStartDate stores it, but ValidateStartDate will catch it
	task.SetStartDate(StartDateKindSpecificDate, nil)

	if task.StartDateKind != StartDateKindSpecificDate {
		t.Fatalf("expected kind=specific_date, got %q", task.StartDateKind)
	}
	if task.StartDate != nil {
		t.Fatalf("expected date=nil (SetStartDate stores what's given), got %v", task.StartDate)
	}
}

func TestSetStartDate_SwitchFromSpecificToInbox(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC)
	task.SetStartDate(StartDateKindSpecificDate, &d)
	task.SetStartDate(StartDateKindInbox, nil)

	if task.StartDateKind != StartDateKindInbox {
		t.Fatalf("expected kind=inbox after switch, got %q", task.StartDateKind)
	}
	if task.StartDate != nil {
		t.Fatalf("expected date=nil after switch to inbox, got %v", task.StartDate)
	}
}

func TestSetStartDate_SwitchFromInboxToSpecific(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	task.SetStartDate(StartDateKindInbox, nil)
	task.SetStartDate(StartDateKindSpecificDate, &d)

	if task.StartDateKind != StartDateKindSpecificDate {
		t.Fatalf("expected kind=specific_date after switch, got %q", task.StartDateKind)
	}
	if task.StartDate == nil || !task.StartDate.Equal(d) {
		t.Fatalf("expected date=%v after switch, got %v", d, task.StartDate)
	}
}

// --- ValidateStartDate ---

func TestValidateStartDate_InboxNilDate_OK(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	// Default is inbox + nil, should be valid
	if err := task.ValidateStartDate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateStartDate_SpecificDateWithDate_OK(t *testing.T) {
	task := NewTask("t", "", "owner", nil)
	d := time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC)
	task.SetStartDate(StartDateKindSpecificDate, &d)

	if err := task.ValidateStartDate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateStartDate_SpecificDateNoDate_Error(t *testing.T) {
	task := &Task{
		StartDateKind: StartDateKindSpecificDate,
		StartDate:     nil,
	}
	err := task.ValidateStartDate()
	if err == nil {
		t.Fatal("expected error for specific_date with nil date")
	}
}

func TestValidateStartDate_InboxWithDate_Error(t *testing.T) {
	d := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	task := &Task{
		StartDateKind: StartDateKindInbox,
		StartDate:     &d,
	}
	err := task.ValidateStartDate()
	if err == nil {
		t.Fatal("expected error for inbox with non-nil date")
	}
}

func TestValidateStartDate_UnknownKind_Error(t *testing.T) {
	task := &Task{
		StartDateKind: StartDateKind("unknown"),
		StartDate:     nil,
	}
	err := task.ValidateStartDate()
	if err == nil {
		t.Fatal("expected error for unknown start_date_kind")
	}
}

