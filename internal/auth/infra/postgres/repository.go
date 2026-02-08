package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slips-ai/slips-core/internal/auth/domain"
)

// Repository implements domain.Repository using PostgreSQL
type Repository struct {
	queries *Queries
}

// NewRepository creates a new Auth repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: New(pool),
	}
}

// UpsertUser creates or updates a user
func (r *Repository) UpsertUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	result, err := r.queries.UpsertUser(ctx, UpsertUserParams{
		UserID:    user.UserID,
		Username:  textFromString(user.Username),
		AvatarUrl: textFromString(user.AvatarURL),
		Email:     textFromString(user.Email),
	})
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:        int64(result.ID),
		UserID:    result.UserID,
		Username:  stringFromText(result.Username),
		AvatarURL: stringFromText(result.AvatarUrl),
		Email:     stringFromText(result.Email),
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

// GetUserByUserID retrieves a user by their user ID
func (r *Repository) GetUserByUserID(ctx context.Context, userID string) (*domain.User, error) {
	result, err := r.queries.GetUserByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:        int64(result.ID),
		UserID:    result.UserID,
		Email:     stringFromText(result.Email),
		Username:  stringFromText(result.Username),
		AvatarURL: stringFromText(result.AvatarUrl),
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

// GetUserByID retrieves a user by their database ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	result, err := r.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:        int64(result.ID),
		UserID:    result.UserID,
		Username:  stringFromText(result.Username),
		Email:     stringFromText(result.Email),
		AvatarURL: stringFromText(result.AvatarUrl),
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

// textFromString converts a string to pgtype.Text
func textFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// stringFromText converts pgtype.Text to string, returning empty string if null
func stringFromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}
