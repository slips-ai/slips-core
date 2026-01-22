package auth

import (
	"testing"

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
