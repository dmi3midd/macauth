package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	Id           string    `json:"id" db:"id"`
	RefreshToken string    `json:"refreshToken" db:"refresh_token"`
	UserId       string    `json:"userId" db:"user_id"`
	ClientId     string    `json:"clientId" db:"client_id"`
	ExpiresAt    time.Time `json:"expiresAt" db:"expires_at"`
}

type TokensPair struct {
	RefreshToken string
	AccessToken  string
}

type AccessClaims struct {
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}
