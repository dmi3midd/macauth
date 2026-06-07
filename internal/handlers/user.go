package handlers

import (
	"encoding/json"
	"errors"
	errs "macauth/internal/errors"
	"macauth/internal/models"
	"macauth/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	ClientId string `json:"clientId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
	Password string `json:"password"`
}

func (h *UserHandler) Registration(w http.ResponseWriter, r *http.Request) error {
	var reqBody RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.InternalServerError(err)
	}
	defer r.Body.Close()

	clientId := reqBody.ClientId

	ctx := r.Context()
	if err := h.userService.Registration(
		ctx,
		reqBody.Username,
		reqBody.Email,
		reqBody.Password,
		clientId,
		reqBody.IsAdmin,
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
	ClientId string `json:"clientId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	User         models.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var reqBody LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.InternalServerError(err)
	}
	defer r.Body.Close()

	clientId := reqBody.ClientId

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoginResponse{
		User:         userData.User,
		RefreshToken: userData.Tokens.RefreshToken,
		AccessToken:  userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return errs.InternalServerError(err)
	}

	return nil
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	var reqBody LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.NewBadRequestError(err, "Invalid request body")
	}
	refreshToken := reqBody.RefreshToken
	ctx := r.Context()
	if err := h.userService.Logout(ctx, refreshToken); err != nil {
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

type RefreshRequest struct {
	ClientId     string `json:"clientId"`
	RefreshToken string `json:"refreshToken"`
}
type RefreshResponse struct {
	User         models.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) error {
	var reqBody RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.NewBadRequestError(err, "Invalid request body")
	}
	refreshToken := reqBody.RefreshToken
	clientId := reqBody.ClientId

	ctx := r.Context()
	userData, err := h.userService.Refresh(ctx, refreshToken, clientId)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return errs.NewNotFoundError(err, "User does not exist")
		}
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return errs.NewUnauthorizedError(err, "Invalid refresh token")
		}
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := RefreshResponse{
		User:         userData.User,
		RefreshToken: userData.Tokens.RefreshToken,
		AccessToken:  userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return errs.InternalServerError(err)
	}

	return nil
}

func (h *UserHandler) IsAdmin(w http.ResponseWriter, r *http.Request) error {
	userId := chi.URLParam(r, "userId")
	if userId == "" {
		return errs.NewBadRequestError(errors.New("userId is required"), "userId is required")
	}

	ctx := r.Context()
	isAdmin, err := h.userService.IsAdmin(ctx, userId)
	if err != nil {
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(isAdmin); err != nil {
		return errs.InternalServerError(err)
	}

	return nil
}
