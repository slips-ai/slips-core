package grpc

import (
	"context"

	authv1 "github.com/slips-ai/slips-core/gen/go/auth/v1"
	"github.com/slips-ai/slips-core/internal/auth/application"
	"github.com/slips-ai/slips-core/pkg/auth"
	"github.com/slips-ai/slips-core/pkg/grpcerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the AuthService gRPC server
type Server struct {
	authv1.UnimplementedAuthServiceServer
	service *application.Service
}

// NewServer creates a new Auth gRPC server
func NewServer(service *application.Service) *Server {
	return &Server{
		service: service,
	}
}

// GetAuthorizationURL generates OAuth authorization URL
func (s *Server) GetAuthorizationURL(ctx context.Context, req *authv1.GetAuthorizationURLRequest) (*authv1.GetAuthorizationURLResponse, error) {
	// Validate provider
	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider is required")
	}

	// Currently only support GitHub
	if req.Provider != "github" {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported provider: %s (only 'github' is supported)", req.Provider)
	}

	url, state, err := s.service.GetAuthorizationURL(ctx, req.Provider)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to get authorization URL")
	}

	return &authv1.GetAuthorizationURLResponse{
		Url:   url,
		State: state,
	}, nil
}

// HandleCallback processes OAuth callback
func (s *Server) HandleCallback(ctx context.Context, req *authv1.HandleCallbackRequest) (*authv1.HandleCallbackResponse, error) {
	// Validate input
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}
	if req.State == "" {
		return nil, status.Error(codes.InvalidArgument, "state is required")
	}

	result, err := s.service.HandleCallback(ctx, req.Code, req.State)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to handle OAuth callback")
	}

	// Extract user ID from token for the response
	userID := ""
	if result.AccessToken != "" {
		extractedUserID, err := auth.ExtractUserIDFromToken(result.AccessToken)
		if err == nil {
			userID = extractedUserID
		}
		// If extraction fails, we continue with empty userID
	}

	return &authv1.HandleCallbackResponse{
		Token: &authv1.Token{
			AccessToken:           result.AccessToken,
			AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
			RefreshToken:          result.RefreshToken,
			RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
			TokenType:             result.TokenType,
		},
		UserInfo: &authv1.UserInfo{
			UserId:    userID,
			Username:  result.Username,
			AvatarUrl: result.AvatarURL,
		},
	}, nil
}

// RefreshToken refreshes an access token
func (s *Server) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	// Validate input
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	result, err := s.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to refresh token")
	}

	return &authv1.RefreshTokenResponse{
		Token: &authv1.Token{
			AccessToken:           result.AccessToken,
			AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
			RefreshToken:          result.RefreshToken,
			RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
			TokenType:             result.TokenType,
		},
	}, nil
}

// GetUserProfile retrieves the current user's profile
func (s *Server) GetUserProfile(ctx context.Context, req *authv1.GetUserProfileRequest) (*authv1.GetUserProfileResponse, error) {
	user, err := s.service.GetUserProfile(ctx)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to get user profile")
	}

	return &authv1.GetUserProfileResponse{
		UserInfo: &authv1.UserInfo{
			UserId:    user.UserID,
			Username:  user.Username,
			AvatarUrl: user.AvatarURL,
		},
	}, nil
}
