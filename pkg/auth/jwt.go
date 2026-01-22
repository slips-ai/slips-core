package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
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

// JWKSKey represents a key in the JWKS
type JWKSKey struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c,omitempty"`
}

// JWKS represents the JSON Web Key Set
type JWKS struct {
	Keys []JWKSKey `json:"keys"`
}

// Claims represents the JWT claims we care about
type Claims struct {
	jwt.RegisteredClaims
	Type string `json:"typ,omitempty"`
	UID  string `json:"uid,omitempty"`
}

// JWTValidator validates JWTs using Identra JWKS
type JWTValidator struct {
	jwksURL       string
	expectedIssuer string
	keys          map[string]*rsa.PublicKey
	mu            sync.RWMutex
	httpClient    *http.Client
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(jwksURL, expectedIssuer string) *JWTValidator {
	return &JWTValidator{
		jwksURL:       jwksURL,
		expectedIssuer: expectedIssuer,
		keys:          make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchJWKS fetches the JWKS from the Identra endpoint
func (v *JWTValidator) FetchJWKS(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to unmarshal JWKS: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Parse and store the public keys
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			continue
		}

		pubKey, err := parseRSAPublicKey(key.N, key.E)
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
	
	// Convert e to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: eInt,
	}, nil
}

// ValidateToken validates a JWT token
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

	// Validate token type (must be "access")
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

// ExtractUserID extracts user ID from claims
// Prefers "sub" claim, but falls back to "uid" for compatibility
func ExtractUserID(claims *Claims) (string, error) {
	// Prefer sub claim
	if claims.Subject != "" {
		return claims.Subject, nil
	}

	// Fall back to uid
	if claims.UID != "" {
		return claims.UID, nil
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
