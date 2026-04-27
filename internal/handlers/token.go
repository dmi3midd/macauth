package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	errs "macauth/internal/errors"
	"macauth/internal/services"
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
		return errs.InternalServerError(err)
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
		return errs.NewUnauthorizedError(
			fmt.Errorf("Invalid or empty Authorization header"),
			"Invalid or empty Authorization header",
		)
	}
	userData, _, err := h.tokenService.ValidateAccessToken(token)
	if err != nil {
		return errs.NewUnauthorizedError(err, "Invalid access token")
	}
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		return errs.InternalServerError(err)
	}
	return nil
}
