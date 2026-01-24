package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slips-ai/slips-core/internal/mcptoken/domain"
)

// MCPTokenRepository implements domain.Repository using PostgreSQL
type MCPTokenRepository struct {
	queries *Queries
}

// NewMCPTokenRepository creates a new MCP token repository
func NewMCPTokenRepository(pool *pgxpool.Pool) *MCPTokenRepository {
	return &MCPTokenRepository{
		queries: New(pool),
	}
}

// Create creates a new MCP token
func (r *MCPTokenRepository) Create(ctx context.Context, token *domain.MCPToken) error {
	pgToken := pgtype.UUID{
		Bytes: token.Token,
		Valid: true,
	}

	var pgExpiresAt pgtype.Timestamp
	if token.ExpiresAt != nil {
		pgExpiresAt = pgtype.Timestamp{
			Time:  *token.ExpiresAt,
			Valid: true,
		}
	}

	result, err := r.queries.CreateMCPToken(ctx, CreateMCPTokenParams{
		Token:     pgToken,
		UserID:    token.UserID,
		Name:      token.Name,
		ExpiresAt: pgExpiresAt,
	})
	if err != nil {
		return err
	}

	tokenID, err := uuid.FromBytes(result.ID.Bytes[:])
	if err != nil {
		return err
	}

	token.ID = tokenID
	token.CreatedAt = result.CreatedAt.Time
	token.IsActive = result.IsActive

	if result.ExpiresAt.Valid {
		token.ExpiresAt = &result.ExpiresAt.Time
	}
	if result.LastUsedAt.Valid {
		token.LastUsedAt = &result.LastUsedAt.Time
	}

	return nil
}

// GetByToken retrieves an MCP token by its token value
func (r *MCPTokenRepository) GetByToken(ctx context.Context, token uuid.UUID) (*domain.MCPToken, error) {
	pgToken := pgtype.UUID{
		Bytes: token,
		Valid: true,
	}

	result, err := r.queries.GetMCPTokenByToken(ctx, pgToken)
	if err != nil {
		return nil, err
	}

	return r.toDomain(&result)
}

// GetByID retrieves an MCP token by its ID
func (r *MCPTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MCPToken, error) {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	result, err := r.queries.GetMCPTokenByID(ctx, pgID)
	if err != nil {
		return nil, err
	}

	return r.toDomain(&result)
}

// ListByUserID retrieves all MCP tokens for a user
func (r *MCPTokenRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.MCPToken, error) {
	results, err := r.queries.ListMCPTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokens := make([]*domain.MCPToken, len(results))
	for i, result := range results {
		token, err := r.toDomain(&result)
		if err != nil {
			return nil, err
		}
		tokens[i] = token
	}

	return tokens, nil
}

// UpdateLastUsedAt updates the last used timestamp
func (r *MCPTokenRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	return r.queries.UpdateMCPTokenLastUsedAt(ctx, pgID)
}

// Revoke revokes (deactivates) an MCP token
func (r *MCPTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	return r.queries.RevokeMCPToken(ctx, pgID)
}

// Delete permanently deletes an MCP token
func (r *MCPTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	return r.queries.DeleteMCPToken(ctx, pgID)
}

// Helper function to convert database model to domain model
func (r *MCPTokenRepository) toDomain(row *McpToken) (*domain.MCPToken, error) {
	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	token, err := uuid.FromBytes(row.Token.Bytes[:])
	if err != nil {
		return nil, err
	}

	mcpToken := &domain.MCPToken{
		ID:        id,
		Token:     token,
		UserID:    row.UserID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Time,
		IsActive:  row.IsActive,
	}

	if row.ExpiresAt.Valid {
		mcpToken.ExpiresAt = &row.ExpiresAt.Time
	}

	if row.LastUsedAt.Valid {
		mcpToken.LastUsedAt = &row.LastUsedAt.Time
	}

	return mcpToken, nil
}
