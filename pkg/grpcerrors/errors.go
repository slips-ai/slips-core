package grpcerrors

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError converts an error to an appropriate gRPC status error
func ToGRPCError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}

	// Check for not found errors
	if errors.Is(err, pgx.ErrNoRows) {
		return status.Errorf(codes.NotFound, "%s: %v", defaultMsg, err)
	}

	// Check for unique constraint violations
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 is the PostgreSQL error code for unique_violation
		if pgErr.Code == "23505" {
			return status.Errorf(codes.AlreadyExists, "%s: %v", defaultMsg, err)
		}
	}

	// Default to internal error
	return status.Errorf(codes.Internal, "%s: %v", defaultMsg, err)
}

// ValidateNotEmpty validates that a string is not empty
func ValidateNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return status.Errorf(codes.InvalidArgument, "%s cannot be empty", fieldName)
	}
	return nil
}
