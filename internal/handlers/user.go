package handlers

import (
	"encoding/json"
	"errors"
	errs "macauth/internal/errors"
	"macauth/internal/models"
	"macauth/internal/services"
	"net/http"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type RegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Registration(w http.ResponseWriter, r *http.Request) error {
	var reqBody RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.InternalServerError(err)
	}
	defer r.Body.Close()

	clientId := r.Header.Get("x-client-id")

	ctx := r.Context()
	if err := h.userService.Registration(
		ctx,
		reqBody.Username,
		reqBody.Email,
		reqBody.Password,
		clientId,
	); err != nil {
		if errors.Is(err, services.ErrUserAlreadyExist) {
			return errs.NewConflictError(err, "User already exist with this email")
		}
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var reqBody LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.InternalServerError(err)
	}
	defer r.Body.Close()

	clientId := r.Header.Get("x-client-id")

	ctx := r.Context()
	userData, err := h.userService.Login(ctx, reqBody.Email, reqBody.Password, clientId)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return errs.NewNotFoundError(err, "User does not exist with this email")
		}
		if errors.Is(err, services.ErrInvalidPassword) {
			return errs.NewBadRequestError(err, "Invalid password")
		}
		return errs.InternalServerError(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    userData.Tokens.RefreshToken,
		MaxAge:   14 * 24 * 60 * 60,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := struct {
		User        models.UserDto `json:"user"`
		AccessToken string         `json:"accessToken"`
	}{
		User:        userData.User,
		AccessToken: userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return errs.InternalServerError(err)
	}

	return nil
}
