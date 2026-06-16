package handlers

import (
	"encoding/json"
	"errors"
	"macauth/internal/auth/models"
	"macauth/internal/auth/services"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/httputil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService services.UserService
	validate    *validator.Validate
}

func NewUserHandler(
	userService services.UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

func (h *UserHandler) Registration(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[RegistrationRequest](r, h.validate)
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
			return apierror.NewConflictError(err, "User already exist with this email")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[LoginRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Login(ctx, reqBody.Email, reqBody.Password, reqBody.ClientId)
	if err != nil {
		if errors.Is(err, services.ErrServiceUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist with this email")
		}
		if errors.Is(err, services.ErrInvalidPassword) {
			return apierror.NewBadRequestError(err, "Invalid password")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoginResponse{
		User:         userData.User,
		RefreshToken: userData.Tokens.RefreshToken,
		AccessToken:  userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[LogoutRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	if err := h.userService.Logout(ctx, reqBody.RefreshToken); err != nil {
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[RefreshRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Refresh(ctx, reqBody.RefreshToken, reqBody.ClientId)
	if err != nil {
		if errors.Is(err, services.ErrServiceUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist")
		}
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return apierror.NewUnauthorizedError(err, "Invalid refresh token")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := RefreshResponse{
		User:         userData.User,
		RefreshToken: userData.Tokens.RefreshToken,
		AccessToken:  userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

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
	User         models.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
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
