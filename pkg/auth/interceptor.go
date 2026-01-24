package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary interceptor for JWT authentication
func UnaryServerInterceptor(validator *JWTValidator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Extract bearer token
		tokenString, err := ExtractBearerToken(authHeaders[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// Validate token
		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Extract user ID
		userID, err := ExtractUserID(claims)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token claims: %v", err)
		}

		// Add user ID to context
		ctx = WithUserID(ctx, userID)

		// Call the handler
		return handler(ctx, req)
	}
}

// UnaryServerInterceptorWithMCP returns a gRPC unary interceptor that supports both JWT and MCP token authentication
func UnaryServerInterceptorWithMCP(jwtValidator *JWTValidator, mcpValidator MCPTokenValidator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		authHeader := authHeaders[0]
		var userID string

		// Try to determine the token type based on the prefix
		if strings.HasPrefix(authHeader, "Bearer ") {
			// JWT token
			tokenString, err := ExtractBearerToken(authHeader)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, err.Error())
			}

			claims, err := jwtValidator.ValidateToken(tokenString)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid JWT token: %v", err)
			}

			userID, err = ExtractUserID(claims)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token claims: %v", err)
			}
		} else if strings.HasPrefix(authHeader, "MCP-Token ") {
			// MCP token
			token, err := ExtractMCPToken(authHeader)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid MCP token format: %v", err)
			}

			userID, err = mcpValidator.ValidateToken(ctx, token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid MCP token: %v", err)
			}
		} else {
			return nil, status.Error(codes.Unauthenticated, "unsupported authentication scheme (expected 'Bearer' or 'MCP-Token')")
		}

		// Add user ID to context
		ctx = WithUserID(ctx, userID)

		// Call the handler
		return handler(ctx, req)
	}
}
