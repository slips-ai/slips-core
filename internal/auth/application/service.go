package application

import (
	"context"
	"log/slog"

	"github.com/slips-ai/slips-core/internal/auth/domain"
	"github.com/slips-ai/slips-core/pkg/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("auth-service")

// Service provides authentication business logic including OAuth
type Service struct {
	repo          domain.Repository
	identraClient *auth.IdentraClient
	logger        *slog.Logger
	provider      string
	redirectURL   string
}

// NewService creates a new OAuth service
func NewService(repo domain.Repository, identraClient *auth.IdentraClient, provider, redirectURL string, logger *slog.Logger) *Service {
	return &Service{
		repo:          repo,
		identraClient: identraClient,
		logger:        logger,
		provider:      provider,
		redirectURL:   redirectURL,
	}
}

// GetAuthorizationURL generates OAuth authorization URL
func (s *Service) GetAuthorizationURL(ctx context.Context, provider string) (string, string, error) {
	ctx, span := tracer.Start(ctx, "GetAuthorizationURL", trace.WithAttributes(
		attribute.String("provider", provider),
	))
	defer span.End()

	resp, err := s.identraClient.GetOAuthAuthorizationURL(ctx, provider, s.redirectURL)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get OAuth authorization URL", "error", err, "provider", provider)
		span.RecordError(err)
		return "", "", err
	}

	s.logger.InfoContext(ctx, "OAuth authorization URL generated", "provider", provider)
	return resp.Url, resp.State, nil
}

// HandleCallback processes OAuth callback and returns tokens and user info
func (s *Service) HandleCallback(ctx context.Context, code, state string) (*CallbackResult, error) {
	ctx, span := tracer.Start(ctx, "HandleCallback")
	defer span.End()

	// Exchange code for tokens via identra
	resp, err := s.identraClient.LoginByOAuth(ctx, code, state)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to login by OAuth", "error", err)
		span.RecordError(err)
		return nil, err
	}

	// Store user info in database only if username, avatar, or email are provided
	if resp.Username != "" || resp.AvatarUrl != "" || resp.Email != "" {
		// Extract user ID from the access token
		userID, err := auth.ExtractUserIDFromToken(resp.Token.AccessToken.Token)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to extract user ID from token", "error", err)
			span.RecordError(err)
			return nil, err
		}

		// Upsert user (only updates if fields are NULL)
		user := domain.NewUser(userID, resp.Username, resp.AvatarUrl, resp.Email)
		_, err = s.repo.UpsertUser(ctx, user)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to upsert user", "error", err, "user_id", userID)
			span.RecordError(err)
			// Don't fail the entire login if user storage fails
			// Log the error and continue
		} else {
			s.logger.InfoContext(ctx, "user info stored", "user_id", userID, "username", resp.Username, "email", resp.Email)
		}
	}

	result := &CallbackResult{
		AccessToken:           resp.Token.AccessToken.Token,
		AccessTokenExpiresAt:  resp.Token.AccessToken.ExpiresAt,
		RefreshToken:          resp.Token.RefreshToken.Token,
		RefreshTokenExpiresAt: resp.Token.RefreshToken.ExpiresAt,
		TokenType:             resp.Token.TokenType,
		Username:              resp.Username,
		AvatarURL:             resp.AvatarUrl,
		Email:                 resp.Email,
	}

	return result, nil
}

// RefreshToken refreshes the access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	ctx, span := tracer.Start(ctx, "RefreshToken")
	defer span.End()

	resp, err := s.identraClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to refresh token", "error", err)
		span.RecordError(err)
		return nil, err
	}

	result := &TokenResult{
		AccessToken:           resp.Token.AccessToken.Token,
		AccessTokenExpiresAt:  resp.Token.AccessToken.ExpiresAt,
		RefreshToken:          resp.Token.RefreshToken.Token,
		RefreshTokenExpiresAt: resp.Token.RefreshToken.ExpiresAt,
		TokenType:             resp.Token.TokenType,
	}

	s.logger.InfoContext(ctx, "token refreshed successfully")
	return result, nil
}

// GetUserProfile retrieves user profile from database
func (s *Service) GetUserProfile(ctx context.Context) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfile")
	defer span.End()

	// Extract user ID from context (JWT token)
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	user, err := s.repo.GetUserByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user profile", "error", err, "user_id", userID)
		span.RecordError(err)
		return nil, err
	}

	return user, nil
}

// UpdateUserProfile updates current user's profile settings
func (s *Service) UpdateUserProfile(ctx context.Context, tavilyMCPToken string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserProfile")
	defer span.End()

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	updatedUser, err := s.repo.UpdateUserTavilyMCPToken(ctx, userID, tavilyMCPToken)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update user profile", "error", err, "user_id", userID)
		span.RecordError(err)
		return nil, err
	}

	return updatedUser, nil
}

// CallbackResult contains the result of OAuth callback processing
type CallbackResult struct {
	AccessToken           string
	AccessTokenExpiresAt  int64
	RefreshToken          string
	RefreshTokenExpiresAt int64
	TokenType             string
	Username              string
	Email                 string
	AvatarURL             string
}

// TokenResult contains the result of token refresh
type TokenResult struct {
	AccessToken           string
	AccessTokenExpiresAt  int64
	RefreshToken          string
	RefreshTokenExpiresAt int64
	TokenType             string
}
