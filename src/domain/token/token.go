package token

import "time"

type TokenType string

const (
	Refresh TokenType = "REFRESH"
	Reset   TokenType = "RESET"
)

type Token struct {
	ID        int
	UserID    int
	Token     string
	Type      TokenType
	ExpiresAt time.Time
	CreatedAt time.Time
}
