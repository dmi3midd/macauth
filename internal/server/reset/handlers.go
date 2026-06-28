package reset

import (
	"encoding/json"
	"errors"
	"macauth/internal/service"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/httputil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ResetHandler struct {
	resetService service.ResetService
	validate     *validator.Validate
}

func NewResetHandler(resetService service.ResetService) *ResetHandler {
	return &ResetHandler{
		resetService: resetService,
		validate:     validator.New(),
	}
}

func (h *ResetHandler) InitiateReset(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[InitiateResetRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	userData, err := h.resetService.InitiateReset(ctx, reqBody.Email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := InitiateResetResponse{
		ResetToken: userData.ResetToken,
		Email:      userData.Email,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

func (h *ResetHandler) ConfirmReset(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[ConfirmResetRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	err = h.resetService.ConfirmReset(ctx, reqBody.ResetToken, reqBody.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist")
		}
		if errors.Is(err, service.ErrTokenUsed) {
			return apierror.NewBadRequestError(err, "Token already used")
		}
		if errors.Is(err, service.ErrTokenExpired) {
			return apierror.NewBadRequestError(err, "Token expired")
		}
		if errors.Is(err, service.ErrInvalidToken) {
			return apierror.NewBadRequestError(err, "Invalid token")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

type InitiateResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type InitiateResetResponse struct {
	ResetToken string `json:"resetToken"`
	Email      string `json:"email"`
}

type ConfirmResetRequest struct {
	ResetToken  string `json:"resetToken" validate:"required,eq=64"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=64"`
}
