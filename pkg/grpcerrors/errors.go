package grpcerrors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// MaxTitleLength is the maximum allowed length for task titles
	MaxTitleLength = 500
	// MaxNotesLength is the maximum allowed length for task notes
	MaxNotesLength = 50000
	// MaxTagNameLength is the maximum allowed length for tag names
	MaxTagNameLength = 100
)

// ToGRPCError converts an error to an appropriate gRPC status error
// Note: This includes the original error which may contain sensitive info.
// Use with caution in production and ensure detailed errors are logged server-side.
func ToGRPCError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}

	// Check for not found errors
	if errors.Is(err, pgx.ErrNoRows) {
		return status.Errorf(codes.NotFound, "%s", defaultMsg)
	}

	// Check for unique constraint violations
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 is the PostgreSQL error code for unique_violation
		if pgErr.Code == "23505" {
			return status.Errorf(codes.AlreadyExists, "%s: duplicate entry", defaultMsg)
		}
	}

	// Default to internal error - don't leak internal details
	return status.Errorf(codes.Internal, "%s", defaultMsg)
}

// ValidateNotEmpty validates that a string is not empty
func ValidateNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return status.Errorf(codes.InvalidArgument, "%s cannot be empty", fieldName)
	}
	return nil
}

// ValidateLength validates that a string does not exceed the maximum length
func ValidateLength(value, fieldName string, maxLength int) error {
	if len(value) > maxLength {
		return status.Errorf(codes.InvalidArgument, "%s exceeds maximum length of %d characters", fieldName, maxLength)
	}
	return nil
}

// ValidateTagName validates tag name requirements
func ValidateTagName(name string) error {
	if err := ValidateNotEmpty(name, "name"); err != nil {
		return err
	}
	if err := ValidateLength(name, "name", MaxTagNameLength); err != nil {
		return err
	}
	// Check for control characters and other invalid characters
	for i, r := range name {
		if r < 32 || r == 127 {
			return status.Errorf(codes.InvalidArgument, "name contains invalid character at position %d", i)
		}
	}
	return nil
}

// ValidateInt32Range validates that an int value is within int32 bounds
func ValidateInt32Range(value int, fieldName string) error {
	if value < 0 {
		return status.Errorf(codes.InvalidArgument, "%s cannot be negative", fieldName)
	}
	if value > 2147483647 {
		return status.Errorf(codes.InvalidArgument, "%s exceeds maximum value of 2147483647", fieldName)
	}
	return nil
}
