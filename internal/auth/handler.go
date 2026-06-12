package auth

import (
	"encoding/json"
	"errors"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/httputil"
	"macauth/internal/token"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService UserService
	validate    *validator.Validate
}

func NewUserHandler(
	userService UserService,
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
		if errors.Is(err, ErrUserAlreadyExist) {
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
		if errors.Is(err, ErrServiceUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist with this email")
		}
		if errors.Is(err, ErrInvalidPassword) {
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
		if errors.Is(err, ErrServiceUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist")
		}
		if errors.Is(err, token.ErrInvalidRefreshToken) {
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
