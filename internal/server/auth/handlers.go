package auth

import (
	"encoding/json"
	"macauth/internal/domain"
	"macauth/internal/service"
	"macauth/internal/shared/httputil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(
	userService service.UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

type RegistrationRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
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
	); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

type LoginRequest struct {
	ClientId string `json:"clientId" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type LoginResponse struct {
	User         domain.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[LoginRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Login(
		ctx,
		reqBody.Email,
		reqBody.Password,
		reqBody.ClientId,
	)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoginResponse{
		User:         userData.User,
		RefreshToken: userData.Tokens.RefreshToken,
		AccessToken:  userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return err
	}

	return nil
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[LogoutRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	if err := h.userService.Logout(ctx, reqBody.RefreshToken); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

type RefreshRequest struct {
	ClientId     string `json:"clientId" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type RefreshResponse struct {
	User         domain.UserDto `json:"user"`
	RefreshToken string         `json:"refreshToken"`
	AccessToken  string         `json:"accessToken"`
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[RefreshRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Refresh(ctx, reqBody.RefreshToken, reqBody.ClientId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := RefreshResponse{
		User:         userData.User,
		RefreshToken: userData.Tokens.RefreshToken,
		AccessToken:  userData.Tokens.AccessToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return err
	}

	return nil
}

type ValidateRequest struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

type ValidateResponse struct {
	User domain.UserDto `json:"user"`
}

func (h *UserHandler) Validate(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[ValidateRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.userService.Validate(ctx, reqBody.AccessToken)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := ValidateResponse{
		User: *userData,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return err
	}

	return nil
}
