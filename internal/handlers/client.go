package handlers

import (
	"encoding/json"
	"errors"
	errs "macauth/internal/errors"
	"macauth/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var (
	ErrEmptyClientId = errors.New("client id is empty")
)

type ClientHandler struct {
	clientService services.ClientService
}

func NewClientHandler(clientService services.ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
	}
}

type LinkRequest struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

func (h *ClientHandler) Link(w http.ResponseWriter, r *http.Request) error {
	var reqBody LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return errs.InternalServerError(err)
	}
	defer r.Body.Close()

	ctx := r.Context()
	clientId, err := h.clientService.Link(ctx, reqBody.Name, reqBody.Secret)
	if err != nil {
		if errors.Is(err, services.ErrClientAlreadyExist) {
			return errs.NewConflictError(err, "Client already exist with this name")
		}
		return errs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-client-id", clientId)
	w.WriteHeader(http.StatusOK)

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
	w.Header().Del("x-client-id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}
