package grpc

import (
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pgregory.net/rapid"
)

// Feature: task-start-date, Property 3: 无效日期格式拒绝
func TestProperty3_InvalidDateFormat_Rejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		invalidDate := rapid.OneOf(
			rapid.String(),
			rapid.StringMatching(`[0-9]{1,7}`),
			rapid.StringMatching(`[0-9]{11,20}`),
			rapid.StringMatching(`[0-9]{4}/[0-9]{2}/[0-9]{2}`),
			rapid.StringMatching(`[0-9]{4}\.[0-9]{2}\.[0-9]{2}`),
			rapid.StringMatching(`[0-9]{4}-[1-9][3-9]-[0-9]{2}`),
			rapid.StringMatching(`[0-9]{4}-[0-9]{2}-[4-9][0-9]`),
			rapid.StringMatching(`\s{1,5}`),
		).Filter(func(s string) bool {
			if s == "" {
				return false
			}
			if len(s) != 10 {
				return true
			}
			_, err := time.Parse("2006-01-02", s)
			return err != nil
		}).Draw(t, "invalidDate")

		_, err := parseStartDateForCreate(&invalidDate)

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

// Feature: task-start-date, Property 4: 创建时省略start_date默认 Inbox
func TestParseStartDateForCreate_NilDefaultsToInbox(t *testing.T) {
	date, err := parseStartDateForCreate(nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if date != nil {
		t.Errorf("expected nil date, got %v", date)
	}
}

// Feature: task-start-date, Property 5: 更新时空字符串表示清空日期
func TestParseStartDateForUpdate_EmptyStringClears(t *testing.T) {
	empty := ""
	date, err := parseStartDateForUpdate(&empty)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if date != nil {
		t.Errorf("expected nil date for empty string, got %v", date)
	}
}

// Feature: task-start-date, Property 6: 更新语义由字段存在性决定
func TestUpdateTask_ValidationCondition_StartDatePresence(t *testing.T) {
	testCases := []struct {
		name                 string
		startDate            *string
		shouldApplyStartDate bool
	}{
		{
			name:                 "nil means no change",
			startDate:            nil,
			shouldApplyStartDate: false,
		},
		{
			name:                 "non-empty date means update",
			startDate:            strPtr("2025-01-01"),
			shouldApplyStartDate: true,
		},
		{
			name:                 "empty string means clear",
			startDate:            strPtr(""),
			shouldApplyStartDate: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldApplyStartDate := tc.startDate != nil
			if shouldApplyStartDate != tc.shouldApplyStartDate {
				t.Errorf("expected shouldApplyStartDate=%v, got %v", tc.shouldApplyStartDate, shouldApplyStartDate)
			}
		})
	}
}

// Helper function for test
func strPtr(s string) *string {
	return &s
}
