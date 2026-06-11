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
	userService          services.UserService
	passwordResetService services.ResetService
}

func NewUserHandler(
	userService services.UserService,
	passwordResetService services.ResetService,
) *UserHandler {
	return &UserHandler{
		userService:          userService,
		passwordResetService: passwordResetService,
	}
}

type RegistrationRequest struct {
	ClientId    string   `json:"clientId"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	Password    string   `json:"password"`
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
		reqBody.Permissions,
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

type InitiateResetRequest struct {
	Email string `json:"email"`
}

type InitiateResetResponse struct {
	ResetToken string `json:"resetToken"`
	Email      string `json:"email"`
}

func (h *UserHandler) InitiateReset(w http.ResponseWriter, r *http.Request) error {
	var reqBody InitiateResetRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.NewBadRequestError(err, "Invalid request body")
	}

	ctx := r.Context()
	userData, err := h.passwordResetService.InitiateReset(ctx, reqBody.Email)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return errs.NewNotFoundError(err, "User does not exist")
		}
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := InitiateResetResponse{
		ResetToken: userData.ResetToken,
		Email:      userData.Email,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return errs.InternalServerError(err)
	}

	return nil
}

type ConfirmResetRequest struct {
	ResetToken  string `json:"resetToken"`
	NewPassword string `json:"newPassword"`
}

func (h *UserHandler) ConfirmReset(w http.ResponseWriter, r *http.Request) error {
	var reqBody ConfirmResetRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.NewBadRequestError(err, "Invalid request body")
	}

	ctx := r.Context()
	err := h.passwordResetService.ConfirmReset(ctx, reqBody.ResetToken, reqBody.NewPassword)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return errs.NewNotFoundError(err, "User does not exist")
		}
		if errors.Is(err, services.ErrTokenUsed) {
			return errs.NewBadRequestError(err, "Token already used")
		}
		if errors.Is(err, services.ErrTokenExpired) {
			return errs.NewBadRequestError(err, "Token expired")
		}
		if errors.Is(err, services.ErrInvalidToken) {
			return errs.NewBadRequestError(err, "Invalid token")
		}
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}
