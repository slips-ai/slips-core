package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
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

func TestIsAuthServicePublicMethod(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		want       bool
	}{
		{
			name:       "GetAuthorizationURL is public",
			fullMethod: "/auth.v1.AuthService/GetAuthorizationURL",
			want:       true,
		},
		{
			name:       "HandleCallback is public",
			fullMethod: "/auth.v1.AuthService/HandleCallback",
			want:       true,
		},
		{
			name:       "RefreshToken is public",
			fullMethod: "/auth.v1.AuthService/RefreshToken",
			want:       true,
		},
		{
			name:       "GetUserProfile is not public",
			fullMethod: "/auth.v1.AuthService/GetUserProfile",
			want:       false,
		},
		{
			name:       "Task service method is not public",
			fullMethod: "/task.v1.TaskService/CreateTask",
			want:       false,
		},
		{
			name:       "MCP token service method is not public",
			fullMethod: "/mcptoken.v1.MCPTokenService/CreateMCPToken",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAuthServicePublicMethod(tt.fullMethod)
			if got != tt.want {
				t.Errorf("isAuthServicePublicMethod(%q) = %v, want %v", tt.fullMethod, got, tt.want)
			}
		})
	}
}

func TestUnaryServerInterceptorWithMCP_PublicMethod(t *testing.T) {
	// Create mock validator (not actually used for public methods)
	jwtValidator := &JWTValidator{}

	// Create mock MCP validator
	mockMCPValidator := &mockMCPTokenValidator{}

	interceptor := UnaryServerInterceptorWithMCP(jwtValidator, mockMCPValidator)

	// Create context without any authorization header (should still succeed for public methods)
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.v1.AuthService/GetAuthorizationURL",
	}

	resp, err := interceptor(ctx, nil, info, mockHandler)

	if err != nil {
		t.Fatalf("expected no error for public method, got: %v", err)
	}

	if resp != "success" {
		t.Errorf("expected 'success' response, got: %v", resp)
	}
}

func TestUnaryServerInterceptorWithMCP_NonPublicMethod_MissingAuth(t *testing.T) {
	jwtValidator := &JWTValidator{}
	mockMCPValidator := &mockMCPTokenValidator{}

	interceptor := UnaryServerInterceptorWithMCP(jwtValidator, mockMCPValidator)

	// Create context without authorization header
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.v1.AuthService/GetUserProfile",
	}

	_, err := interceptor(ctx, nil, info, mockHandler)

	if err == nil {
		t.Fatal("expected error for non-public method without auth, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated code, got %v", st.Code())
	}
}

// mockMCPTokenValidator is a simple mock for testing
type mockMCPTokenValidator struct{}

func (m *mockMCPTokenValidator) ValidateToken(ctx context.Context, token uuid.UUID) (string, error) {
	return "test-user-id", nil
}
