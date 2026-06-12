package handlers

import (
	"encoding/json"
	"errors"
	errs "macauth/internal/errors"
	"macauth/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var (
	ErrEmptyClientId = errors.New("client id is empty")
)

type ClientHandler struct {
	clientService services.ClientService
	validate      *validator.Validate
}

func NewClientHandler(clientService services.ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
		validate:      validator.New(),
	}
}

type LinkRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=64"`
	Secret string `json:"secret" validate:"required,min=8,max=64"`
}
type LinkResponse struct {
	ClientId string `json:"clientId"`
}

func (h *ClientHandler) Link(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := BindAndValidate[LinkRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	clientId, err := h.clientService.Link(ctx, reqBody.Name, reqBody.Secret)
	if err != nil {
		if errors.Is(err, services.ErrClientAlreadyExist) {
			return errs.NewConflictError(err, "Client already exist with this name")
		}
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LinkResponse{
		ClientId: clientId,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return errs.InternalServerError(err)
	}

	return nil
}

func (h *ClientHandler) Unlink(w http.ResponseWriter, r *http.Request) error {
	clientId := chi.URLParam(r, "clientId")
	if clientId == "" {
		return errs.NewBadRequestError(ErrEmptyClientId, "Client id is not provided")
	}

	ctx := r.Context()
	if err := h.clientService.Unlink(ctx, clientId); err != nil {
		if errors.Is(err, services.ErrClientNotFound) {
			return errs.NewNotFoundError(err, "Client does not exist")
		}
		return errs.InternalServerError(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}
