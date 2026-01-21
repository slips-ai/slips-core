package grpcerrors

import (
"errors"

"github.com/jackc/pgx/v5"
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

// Default to internal error
return status.Errorf(codes.Internal, "%s: %v", defaultMsg, err)
}
