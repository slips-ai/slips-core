package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockHandler is a simple handler for testing
func mockHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "success", nil
}

func TestUnaryServerInterceptor_MissingMetadata(t *testing.T) {
	validator := &JWTValidator{}
	interceptor := UnaryServerInterceptor(validator)
	
	ctx := context.Background()
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)
	
	if err == nil {
		t.Fatal("expected error for missing metadata, got nil")
	}
	
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated code, got %v", st.Code())
	}
	
	if st.Message() != "missing metadata" {
		t.Errorf("unexpected error message: %s", st.Message())
	}
}

func TestUnaryServerInterceptor_MissingAuthorizationHeader(t *testing.T) {
	validator := &JWTValidator{}
	interceptor := UnaryServerInterceptor(validator)
	
	// Create context with metadata but no authorization header
	md := metadata.New(map[string]string{"other-header": "value"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)
	
	if err == nil {
		t.Fatal("expected error for missing authorization header, got nil")
	}
	
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated code, got %v", st.Code())
	}
	
	if st.Message() != "missing authorization header" {
		t.Errorf("unexpected error message: %s", st.Message())
	}
}

func TestUnaryServerInterceptor_InvalidAuthorizationFormat(t *testing.T) {
	validator := &JWTValidator{}
	interceptor := UnaryServerInterceptor(validator)
	
	// Create context with invalid authorization header
	md := metadata.New(map[string]string{"authorization": "InvalidFormat"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)
	
	if err == nil {
		t.Fatal("expected error for invalid authorization format, got nil")
	}
	
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated code, got %v", st.Code())
	}
}

func TestUnaryServerInterceptor_ValidToken(t *testing.T) {
	// This test would require a full JWT validator setup with JWKS
	// Skipping for now as it requires significant infrastructure
	t.Skip("Full integration test requires JWKS server setup")
}
