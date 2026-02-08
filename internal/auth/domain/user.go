package domain

import (
	"time"
)

// User represents a user entity in the OAuth context
type User struct {
	ID        int64
	UserID    string
	Username  string
	AvatarURL string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new user instance
func NewUser(userID, username, avatarURL, email string) *User {
	return &User{
		UserID:    userID,
		Username:  username,
		AvatarURL: avatarURL,
		Email:     email,
	}
}
