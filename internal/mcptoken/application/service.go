package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/slips-ai/slips-core/internal/mcptoken/domain"
	"github.com/slips-ai/slips-core/pkg/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("mcptoken-service")

var (
	ErrUnauthorized = errors.New("unauthorized: user mismatch")
)

// Service provides MCP token business logic
type Service struct {
	repo   domain.Repository
	logger *slog.Logger
}

// NewService creates a new MCP token service
func NewService(repo domain.Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateToken creates a new MCP token for the authenticated user
func (s *Service) CreateToken(ctx context.Context, name string, expiresAt *time.Time) (*domain.MCPToken, error) {
	ctx, span := tracer.Start(ctx, "CreateToken", trace.WithAttributes(
		attribute.String("name", name),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	// Create new token
	token := &domain.MCPToken{
		Token:     uuid.New(),
		UserID:    userID,
		Name:      name,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := s.repo.Create(ctx, token); err != nil {
		s.logger.ErrorContext(ctx, "failed to create MCP token", "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "MCP token created", "id", token.ID, "owner_id", userID)
	return token, nil
}

// GetToken retrieves an MCP token by ID (only if owned by the authenticated user)
func (s *Service) GetToken(ctx context.Context, id uuid.UUID) (*domain.MCPToken, error) {
	ctx, span := tracer.Start(ctx, "GetToken", trace.WithAttributes(
		attribute.String("id", id.String()),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	token, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get MCP token", "id", id, "error", err)
		span.RecordError(err)
		return nil, err
	}

	// Verify ownership
	if token.UserID != userID {
		s.logger.WarnContext(ctx, "unauthorized MCP token access attempt", "token_id", id, "token_owner", token.UserID, "requester", userID)
		return nil, ErrUnauthorized
	}

	return token, nil
}

// ListTokens retrieves all MCP tokens for the authenticated user
func (s *Service) ListTokens(ctx context.Context) ([]*domain.MCPToken, error) {
	ctx, span := tracer.Start(ctx, "ListTokens")
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return nil, err
	}

	tokens, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list MCP tokens", "user_id", userID, "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "listed MCP tokens", "user_id", userID, "count", len(tokens))
	return tokens, nil
}

// RevokeToken revokes an MCP token (only if owned by the authenticated user)
func (s *Service) RevokeToken(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "RevokeToken", trace.WithAttributes(
		attribute.String("id", id.String()),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return err
	}

	// Get token to verify ownership
	token, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get MCP token for revocation", "id", id, "error", err)
		span.RecordError(err)
		return err
	}

	// Verify ownership
	if token.UserID != userID {
		s.logger.WarnContext(ctx, "unauthorized MCP token revoke attempt", "token_id", id, "token_owner", token.UserID, "requester", userID)
		return ErrUnauthorized
	}

	if err := s.repo.Revoke(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to revoke MCP token", "id", id, "error", err)
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "MCP token revoked", "id", id, "owner_id", userID)
	return nil
}

// DeleteToken permanently deletes an MCP token (only if owned by the authenticated user)
func (s *Service) DeleteToken(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "DeleteToken", trace.WithAttributes(
		attribute.String("id", id.String()),
	))
	defer span.End()

	// Extract user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user ID from context", "error", err)
		span.RecordError(err)
		return err
	}

	// Get token to verify ownership
	token, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get MCP token for deletion", "id", id, "error", err)
		span.RecordError(err)
		return err
	}

	// Verify ownership
	if token.UserID != userID {
		s.logger.WarnContext(ctx, "unauthorized MCP token delete attempt", "token_id", id, "token_owner", token.UserID, "requester", userID)
		return ErrUnauthorized
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete MCP token", "id", id, "error", err)
		span.RecordError(err)
		return err
	}

	s.logger.InfoContext(ctx, "MCP token deleted", "id", id, "owner_id", userID)
	return nil
}

// ValidateToken validates an MCP token and returns the associated user ID
// This is used by the auth interceptor and does not require authentication
func (s *Service) ValidateToken(ctx context.Context, tokenValue uuid.UUID) (string, error) {
	ctx, span := tracer.Start(ctx, "ValidateToken")
	defer span.End()

	token, err := s.repo.GetByToken(ctx, tokenValue)
	if err != nil {
		s.logger.DebugContext(ctx, "MCP token not found", "error", err)
		span.RecordError(err)
		return "", err
	}

	// Check if token is valid (active and not expired)
	if !token.IsValid() {
		if !token.IsActive {
			s.logger.DebugContext(ctx, "MCP token is inactive", "token_id", token.ID)
			return "", errors.New("token is inactive")
		}
		if token.IsExpired() {
			s.logger.DebugContext(ctx, "MCP token is expired", "token_id", token.ID)
			return "", errors.New("token is expired")
		}
	}

	// Update last used timestamp asynchronously
	go func() {
		// Use background context to avoid cancellation
		updateCtx := context.Background()
		if err := s.repo.UpdateLastUsedAt(updateCtx, token.ID); err != nil {
			s.logger.WarnContext(updateCtx, "failed to update MCP token last used timestamp", "token_id", token.ID, "error", err)
		}
	}()

	s.logger.DebugContext(ctx, "MCP token validated", "token_id", token.ID, "user_id", token.UserID)
	return token.UserID, nil
}
