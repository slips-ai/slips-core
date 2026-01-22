package auth

import (
	"context"

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
