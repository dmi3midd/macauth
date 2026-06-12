package handlers

import (
	"encoding/json"
	"errors"
	errs "macauth/internal/errors"
	"macauth/internal/models"
	"macauth/internal/services"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService          services.UserService
	passwordResetService services.ResetService
	validate             *validator.Validate
}

func NewUserHandler(
	userService services.UserService,
	passwordResetService services.ResetService,
) *UserHandler {
	return &UserHandler{
		userService:          userService,
		passwordResetService: passwordResetService,
		validate:             validator.New(),
	}
}

type RegistrationRequest struct {
	ClientId    string   `json:"clientId" validate:"required,eq=20"`
	Username    string   `json:"username" validate:"required,min=3,max=64"`
	Email       string   `json:"email" validate:"required,email"`
	Permissions []string `json:"permissions" validate:"required,min=1,max=10,dive,min=1,max=64"`
	Password    string   `json:"password" validate:"required,min=8,max=64"`
}

func BindAndValidate[T any](r *http.Request, val *validator.Validate) (T, error) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, errs.NewBadRequestError(err, "Invalid request body")
	}
	defer r.Body.Close()
	if err := val.Struct(body); err != nil {
		return body, errs.NewBadRequestError(err, "Validation failed: "+err.Error())
	}
	return body, nil
}

func (h *UserHandler) Registration(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[RegistrationRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	if err := h.userService.Registration(
		ctx,
		reqBody.Username,
		reqBody.Email,
		reqBody.Password,
		reqBody.ClientId,
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
	ClientId string `json:"clientId" validate:"required,eq=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}
type LoginResponse struct {
	User         models.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[LoginRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Login(ctx, reqBody.Email, reqBody.Password, reqBody.ClientId)
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
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[LogoutRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	if err := h.userService.Logout(ctx, reqBody.RefreshToken); err != nil {
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

type RefreshRequest struct {
	ClientId     string `json:"clientId" validate:"required,eq=20"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}
type RefreshResponse struct {
	User         models.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[RefreshRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Refresh(ctx, reqBody.RefreshToken, reqBody.ClientId)
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
	Email string `json:"email" validate:"required,email"`
}

type InitiateResetResponse struct {
	ResetToken string `json:"resetToken"`
	Email      string `json:"email"`
}

func (h *UserHandler) InitiateReset(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[InitiateResetRequest](r, h.validate)
	if err != nil {
		return err
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
	ResetToken  string `json:"resetToken" validate:"required,eq=64"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=64"`
}

func (h *UserHandler) ConfirmReset(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[ConfirmResetRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	err = h.passwordResetService.ConfirmReset(ctx, reqBody.ResetToken, reqBody.NewPassword)
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
