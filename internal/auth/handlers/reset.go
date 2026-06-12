package reset

import (
	"encoding/json"
	"errors"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/httputil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ResetHandler struct {
	resetService ResetService
	validate     *validator.Validate
}

func NewResetHandler(resetService ResetService) *ResetHandler {
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
		if errors.Is(err, ErrUserNotFound) {
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
		if errors.Is(err, ErrUserNotFound) {
			return apierror.NewNotFoundError(err, "User does not exist")
		}
		if errors.Is(err, ErrTokenUsed) {
			return apierror.NewBadRequestError(err, "Token already used")
		}
		if errors.Is(err, ErrTokenExpired) {
			return apierror.NewBadRequestError(err, "Token expired")
		}
		if errors.Is(err, ErrInvalidToken) {
			return apierror.NewBadRequestError(err, "Invalid token")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}
