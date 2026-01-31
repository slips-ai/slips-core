package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrInvalidTokenType is returned when token type is not 'access'
	ErrInvalidTokenType = errors.New("token type must be 'access'")
	// ErrInvalidIssuer is returned when token issuer doesn't match
	ErrInvalidIssuer = errors.New("invalid token issuer")
)

// Claims represents Identra JWT claims
// This matches Identra's StandardClaims structure with:
// - typ: token type ("access" or "refresh")
// - user_id: user ID (primary identifier from Identra)
type Claims struct {
	jwt.RegisteredClaims
	Type   string `json:"typ,omitempty"`     // Token type: "access" or "refresh"
	UserID string `json:"user_id,omitempty"` // User ID (Identra user_id)
}

// JWTValidator validates Identra JWTs using JWKS
type JWTValidator struct {
	identraClient  *IdentraClient
	expectedIssuer string
	keys           map[string]*rsa.PublicKey
	mu             sync.RWMutex
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(identraClient *IdentraClient, expectedIssuer string) *JWTValidator {
	return &JWTValidator{
		identraClient:  identraClient,
		expectedIssuer: expectedIssuer,
		keys:           make(map[string]*rsa.PublicKey),
	}
}

// FetchJWKS fetches the JWKS from the Identra gRPC endpoint
func (v *JWTValidator) FetchJWKS(ctx context.Context) error {
	resp, err := v.identraClient.GetJWKS(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	if len(resp.Keys) == 0 {
		return errors.New("empty JWKS response")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Parse and store the public keys
	for _, key := range resp.Keys {
		if key.Kty != "RSA" {
			continue
		}

		pubKey, err := parseRSAPublicKey(*key.N, *key.E)
		if err != nil {
			return fmt.Errorf("failed to parse RSA public key: %w", err)
		}

		v.keys[key.Kid] = pubKey
	}

	return nil
}

// parseRSAPublicKey parses RSA public key from n and e
func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	// Decode base64url encoded n
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %w", err)
	}

	// Decode base64url encoded e
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %w", err)
	}

	// Convert to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Verify e fits in an int (typically RSA uses 65537)
	if !e.IsInt64() {
		return nil, fmt.Errorf("RSA exponent too large")
	}
	eInt64 := e.Int64()

	// Common RSA exponents are small (e.g., 65537), verify it's reasonable
	const maxInt32 = int64(1<<31 - 1)
	if eInt64 > maxInt32 || eInt64 <= 0 {
		return nil, fmt.Errorf("invalid RSA exponent: %d", eInt64)
	}
	eInt := int(eInt64)

	return &rsa.PublicKey{
		N: n,
		E: eInt,
	}, nil
}

// ValidateToken validates an Identra JWT token
// The token must:
// - Be signed with RS256 using a key from the JWKS
// - Have typ="access" (refresh tokens are rejected)
// - Have iss matching expectedIssuer
// - Not be expired
func (v *JWTValidator) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		// Get the public key
		v.mu.RLock()
		pubKey, exists := v.keys[kid]
		v.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}

		return pubKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate token type (must be "access", per Identra spec)
	if claims.Type != "access" {
		return nil, ErrInvalidTokenType
	}

	// Validate issuer
	if claims.Issuer != v.expectedIssuer {
		return nil, ErrInvalidIssuer
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// ExtractUserID extracts user ID from Identra claims
// Priority order: user_id (primary), sub (standard JWT)
func ExtractUserID(claims *Claims) (string, error) {
	// Prefer user_id claim (Identra primary identifier)
	if claims.UserID != "" {
		return claims.UserID, nil
	}

	// Fall back to sub claim (standard JWT)
	if claims.Subject != "" {
		return claims.Subject, nil
	}

	return "", errors.New("no user ID found in token claims")
}

// ExtractBearerToken extracts the bearer token from the Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// ExtractUserIDFromToken parses a JWT token and extracts the user ID without full validation
// This is used when we just need the user ID from a token that was already validated by Identra
func ExtractUserIDFromToken(tokenString string) (string, error) {
	// Parse token without verification (we trust Identra's token)
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return ExtractUserID(claims)
}
