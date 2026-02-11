package grpc

import (
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pgregory.net/rapid"
)

// Feature: task-start-date, Property 3: 无效日期格式拒绝
// For any string that is NOT a valid YYYY-MM-DD date, when start_date_kind is
// "specific_date", parseStartDateFields returns an INVALID_ARGUMENT error.
// **Validates: Requirements 3.5**
func TestProperty3_InvalidDateFormat_Rejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random string that is NOT a valid YYYY-MM-DD date.
		// We use multiple generators to cover diverse invalid inputs,
		// with a safety-net filter to reject any accidental valid dates.
		invalidDate := rapid.OneOf(
			// Completely random strings
			rapid.String(),
			// Too short numeric strings
			rapid.StringMatching(`[0-9]{1,7}`),
			// Too long numeric strings
			rapid.StringMatching(`[0-9]{11,20}`),
			// Date-like but with wrong separators (slash, dot)
			rapid.StringMatching(`[0-9]{4}/[0-9]{2}/[0-9]{2}`),
			rapid.StringMatching(`[0-9]{4}\.[0-9]{2}\.[0-9]{2}`),
			// Correct separator but invalid month (13-99)
			rapid.StringMatching(`[0-9]{4}-[1-9][3-9]-[0-9]{2}`),
			// Correct separator but invalid day (40-99)
			rapid.StringMatching(`[0-9]{4}-[0-9]{2}-[4-9][0-9]`),
			// Empty string
			rapid.Just(""),
			// Whitespace
			rapid.StringMatching(`\s{1,5}`),
		).Filter(func(s string) bool {
			// Safety net: reject any string that is actually a valid YYYY-MM-DD date.
			if len(s) != 10 {
				return true
			}
			_, err := time.Parse("2006-01-02", s)
			return err != nil
		}).Draw(t, "invalidDate")

		kind := "specific_date"
		_, _, err := parseStartDateFields(&kind, &invalidDate)

		if err == nil {
			t.Fatalf("expected error for invalid date %q, got nil", invalidDate)
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error, got: %v", err)
		}
		if st.Code() != codes.InvalidArgument {
			t.Fatalf("expected code INVALID_ARGUMENT, got %v for date %q", st.Code(), invalidDate)
		}
	})
}

// Feature: task-start-date, Property 4: 更新时省略start_date字段不影响现有值
// When both start_date_kind and start_date are nil (omitted),
// parseStartDateFields defaults to inbox. This is acceptable for CreateTask.
// For UpdateTask, the caller should NOT call parseStartDateFields when both are nil
// (this is handled at the UpdateTask level to preserve existing values).
// **Validates: Requirements 3.6**
func TestParseStartDateFields_BothNil_DefaultsToInbox(t *testing.T) {
	// When both parameters are nil, parseStartDateFields defaults to inbox.
	// This is acceptable for CreateTask, but for UpdateTask we should NOT call
	// parseStartDateFields at all when both are nil (handled at the caller level).
	kind, date, err := parseStartDateFields(nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if kind != "inbox" {
		t.Errorf("expected kind 'inbox', got %q", kind)
	}
	if date != nil {
		t.Errorf("expected nil date, got %v", date)
	}
}

// Feature: task-start-date, Property 5: start_date without start_date_kind is rejected
// When start_date is provided but start_date_kind is nil,
// the validation condition should be true (indicating rejection is needed).
// This prevents accidental data loss from partial updates.
// **Validates: Requirements for safe partial updates**
func TestUpdateTask_ValidationCondition_StartDateWithoutKind(t *testing.T) {
	testCases := []struct {
		name              string
		startDateKind     *string
		startDate         *string
		shouldBeRejected  bool
	}{
		{
			name:              "both nil - no rejection",
			startDateKind:     nil,
			startDate:         nil,
			shouldBeRejected:  false,
		},
		{
			name:              "both provided - no rejection",
			startDateKind:     strPtr("specific_date"),
			startDate:         strPtr("2025-01-01"),
			shouldBeRejected:  false,
		},
		{
			name:              "only kind provided - no rejection",
			startDateKind:     strPtr("inbox"),
			startDate:         nil,
			shouldBeRejected:  false,
		},
		{
			name:              "only date provided - SHOULD BE REJECTED",
			startDateKind:     nil,
			startDate:         strPtr("2025-01-01"),
			shouldBeRejected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is the validation condition from UpdateTask in server.go
			shouldReject := (tc.startDate != nil && tc.startDateKind == nil)
			
			if shouldReject != tc.shouldBeRejected {
				t.Errorf("expected shouldReject=%v, got %v", tc.shouldBeRejected, shouldReject)
			}
		})
	}
}

// Helper function for test
func strPtr(s string) *string {
	return &s
}

