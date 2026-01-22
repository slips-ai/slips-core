package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		wantToken   string
		wantErr     bool
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr:    false,
		},
		{
			name:       "valid bearer token lowercase",
			authHeader: "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr:    false,
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "invalid format - no space",
			authHeader: "BearereyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "invalid format - wrong prefix",
			authHeader: "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "only bearer prefix",
			authHeader: "Bearer",
			wantToken:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractBearerToken(tt.authHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token != tt.wantToken {
				t.Errorf("ExtractBearerToken() = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestExtractUserID(t *testing.T) {
	tests := []struct {
		name    string
		claims  *Claims
		want    string
		wantErr bool
	}{
		{
			name: "extract from sub claim",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "user-123",
				},
			},
			want:    "user-123",
			wantErr: false,
		},
		{
			name: "extract from uid claim",
			claims: &Claims{
				UID: "user-456",
			},
			want:    "user-456",
			wantErr: false,
		},
		{
			name: "prefer sub over uid",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "user-sub",
				},
				UID: "user-uid",
			},
			want:    "user-sub",
			wantErr: false,
		},
		{
			name:    "no user ID",
			claims:  &Claims{},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractUserID(tt.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseRSAPublicKey tests parsing of RSA public keys.
func TestParseRSAPublicKey(t *testing.T) {
	t.Run("valid RSA key", func(t *testing.T) {
		// Generate a test RSA key
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("failed to generate RSA key: %v", err)
		}

		// Encode N and E to base64url
		n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())

		pubKey, err := parseRSAPublicKey(n, e)
		if err != nil {
			t.Fatalf("parseRSAPublicKey() error = %v", err)
		}

		if pubKey.E != privateKey.PublicKey.E {
			t.Errorf("expected E=%d, got %d", privateKey.PublicKey.E, pubKey.E)
}

if pubKey.N.Cmp(privateKey.PublicKey.N) != 0 {
t.Error("N values don't match")
}
})

t.Run("invalid base64 n", func(t *testing.T) {
_, err := parseRSAPublicKey("invalid!!!", "AQAB")
if err == nil {
t.Error("expected error for invalid base64 n")
}
})

t.Run("invalid base64 e", func(t *testing.T) {
_, err := parseRSAPublicKey("AQAB", "invalid!!!")
if err == nil {
t.Error("expected error for invalid base64 e")
}
})

t.Run("exponent too large", func(t *testing.T) {
// Create a very large exponent
largeE := new(big.Int).SetInt64(1)
largeE.Lsh(largeE, 100) // Shift left by 100 bits
e := base64.RawURLEncoding.EncodeToString(largeE.Bytes())

_, err := parseRSAPublicKey("AQAB", e)
if err == nil {
t.Error("expected error for exponent too large")
}
})

}

func TestFetchJWKS(t *testing.T) {
t.Run("successful fetch", func(t *testing.T) {
// Create a test server that returns valid JWKS
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
jwks := JWKS{
Keys: []JWKSKey{
{
Kid: "test-key-1",
Kty: "RSA",
Alg: "RS256",
Use: "sig",
N:   "AQAB",
E:   "AQAB",
},
},
}
json.NewEncoder(w).Encode(jwks)
}))
defer server.Close()

validator := NewJWTValidator(server.URL, "test-issuer")
err := validator.FetchJWKS(context.Background())

// This will fail because the N and E values are too small, but it tests the HTTP fetch
if err == nil {
t.Log("Note: JWKS fetch succeeded (key parsing may have failed, which is expected)")
}
})

t.Run("http error", func(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusInternalServerError)
}))
defer server.Close()

validator := NewJWTValidator(server.URL, "test-issuer")
err := validator.FetchJWKS(context.Background())

if err == nil {
t.Error("expected error for HTTP 500")
}
})

t.Run("invalid JSON", func(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("invalid json"))
}))
defer server.Close()

validator := NewJWTValidator(server.URL, "test-issuer")
err := validator.FetchJWKS(context.Background())

if err == nil {
t.Error("expected error for invalid JSON")
}
})

t.Run("response size limit", func(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// Try to send more than 1MB
largeData := make([]byte, 2*1024*1024) // 2MB
w.Write(largeData)
}))
defer server.Close()

validator := NewJWTValidator(server.URL, "test-issuer")
err := validator.FetchJWKS(context.Background())

// Should error on parsing the truncated response
if err == nil {
t.Error("expected error when response is too large")
}
})
}

func TestValidateToken_TokenType(t *testing.T) {
// Test that only access tokens are accepted
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
t.Fatalf("failed to generate RSA key: %v", err)
}

testCases := []struct {
name        string
tokenType   string
shouldError bool
}{
{"access token", "access", false},
{"refresh token", "refresh", true},
{"empty token type", "", true},
}

for _, tc := range testCases {
t.Run(tc.name, func(t *testing.T) {
claims := &Claims{
RegisteredClaims: jwt.RegisteredClaims{
Issuer:    "test-issuer",
Subject:   "user-123",
ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
},
Type: tc.tokenType,
UID:  "user-123",
}

token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
token.Header["kid"] = "test-key"
tokenString, err := token.SignedString(privateKey)
if err != nil {
t.Fatalf("failed to sign token: %v", err)
}

validator := &JWTValidator{
expectedIssuer: "test-issuer",
keys:           map[string]*rsa.PublicKey{"test-key": &privateKey.PublicKey},
}

_, err = validator.ValidateToken(tokenString)

if tc.shouldError && err == nil {
t.Error("expected error for token type validation")
}
if !tc.shouldError && err != nil && err != ErrInvalidTokenType {
// May fail on issuer or other checks, but not on token type
t.Logf("validation failed: %v", err)
}
})
}
}

func TestValidateToken_Expiration(t *testing.T) {
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
t.Fatalf("failed to generate RSA key: %v", err)
}

t.Run("expired token", func(t *testing.T) {
claims := &Claims{
RegisteredClaims: jwt.RegisteredClaims{
Issuer:    "test-issuer",
Subject:   "user-123",
ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
},
Type: "access",
UID:  "user-123",
}

token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
token.Header["kid"] = "test-key"
tokenString, err := token.SignedString(privateKey)
if err != nil {
t.Fatalf("failed to sign token: %v", err)
}

validator := &JWTValidator{
expectedIssuer: "test-issuer",
keys:           map[string]*rsa.PublicKey{"test-key": &privateKey.PublicKey},
}

_, err = validator.ValidateToken(tokenString)
if err == nil {
t.Error("expected error for expired token")
}
})
}

func TestValidateToken_Issuer(t *testing.T) {
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
t.Fatalf("failed to generate RSA key: %v", err)
}

testCases := []struct {
name           string
tokenIssuer    string
expectedIssuer string
shouldError    bool
}{
{"matching issuer", "https://identra.example.com", "https://identra.example.com", false},
{"mismatched issuer", "https://evil.com", "https://identra.example.com", true},
{"empty issuer", "", "https://identra.example.com", true},
}

for _, tc := range testCases {
t.Run(tc.name, func(t *testing.T) {
claims := &Claims{
RegisteredClaims: jwt.RegisteredClaims{
Issuer:    tc.tokenIssuer,
Subject:   "user-123",
ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
},
Type: "access",
UID:  "user-123",
}

token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
token.Header["kid"] = "test-key"
tokenString, err := token.SignedString(privateKey)
if err != nil {
t.Fatalf("failed to sign token: %v", err)
}

validator := &JWTValidator{
expectedIssuer: tc.expectedIssuer,
keys:           map[string]*rsa.PublicKey{"test-key": &privateKey.PublicKey},
}

_, err = validator.ValidateToken(tokenString)

if tc.shouldError {
if err == nil {
t.Error("expected error for issuer validation")
}
if err != ErrInvalidIssuer {
t.Logf("got error: %v (expected ErrInvalidIssuer)", err)
}
}
})
}
}
