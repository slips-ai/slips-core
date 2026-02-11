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
