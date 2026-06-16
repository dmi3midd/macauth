package handlers

import (
	"encoding/json"
	"fmt"
	"macauth/internal/auth/services"
	"macauth/internal/shared/apierror"
	"net/http"
	"strings"
)

type TokenHandler struct {
	tokenService services.TokenService
}

func NewTokenHandler(tokenService services.TokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: tokenService,
	}
}

func (h *TokenHandler) GetPublicKey(w http.ResponseWriter, r *http.Request) error {
	key := h.tokenService.GetPublicKey()
	if err := json.NewEncoder(w).Encode(key); err != nil {
		return apierror.InternalServerError(err)
	}
	return nil
}

func (h *TokenHandler) Validate(w http.ResponseWriter, r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	token := ""
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		token = after
	}
	if token == "" {
		return apierror.NewUnauthorizedError(
			fmt.Errorf("Invalid or empty Authorization header"),
			"Invalid or empty Authorization header",
		)
	}
	userData, _, err := h.tokenService.ValidateAccessToken(token)
	if err != nil {
		return apierror.NewUnauthorizedError(err, "Invalid access token")
	}
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		return apierror.InternalServerError(err)
	}
	return nil
}
