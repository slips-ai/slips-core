package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidMCPToken  = errors.New("invalid MCP token format")
	ErrMCPTokenNotFound = errors.New("MCP token not found")
)

// MCPTokenValidator validates MCP tokens
type MCPTokenValidator interface {
	// ValidateToken validates an MCP token and returns the associated user ID
	ValidateToken(ctx context.Context, token uuid.UUID) (string, error)
}

// ExtractMCPToken extracts MCP token from authorization header
// Expects format: "MCP-Token <uuid>"
func ExtractMCPToken(authHeader string) (uuid.UUID, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return uuid.Nil, ErrInvalidMCPToken
	}

	if parts[0] != "MCP-Token" {
		return uuid.Nil, fmt.Errorf("expected MCP-Token scheme, got %s", parts[0])
	}

	token, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	return token, nil
}
