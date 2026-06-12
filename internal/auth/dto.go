package auth

import (
	"macauth/internal/token"
)

type RegistrationRequest struct {
	ClientId    string   `json:"clientId" validate:"required,eq=20"`
	Username    string   `json:"username" validate:"required,min=3,max=64"`
	Email       string   `json:"email" validate:"required,email"`
	Permissions []string `json:"permissions" validate:"required,min=1,max=10,dive,min=1,max=64"`
	Password    string   `json:"password" validate:"required,min=8,max=64"`
}

type LoginRequest struct {
	ClientId string `json:"clientId" validate:"required,eq=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type LoginResponse struct {
	User         token.UserDto `json:"user"`
	RefreshToken string        `json:"refreshToken"`
	AccessToken  string        `json:"accessToken"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type RefreshRequest struct {
	ClientId     string `json:"clientId" validate:"required,eq=20"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type RefreshResponse struct {
	User         token.UserDto `json:"user"`
	RefreshToken string        `json:"refreshToken"`
	AccessToken  string        `json:"accessToken"`
}
