package models

import "github.com/golang-jwt/jwt/v5"

type Token struct {
	Id           string `json:"id" db:"id"`
	RefreshToken string `json:"refreshToken" db:"refresh_token"`
	UserId       string `json:"userId" db:"user_id"`
	ClientId     string `json:"clientId" db:"client_id"`
}

type TokensPair struct {
	RefreshToken string
	AccessToken  string
}

type AccessClaims struct {
	Username string
	Email    string
	IsAdmin  bool
	jwt.RegisteredClaims
}
