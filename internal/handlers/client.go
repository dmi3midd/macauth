package handlers

import (
	"encoding/json"
	"errors"
	customerrs "macauth/internal/errors"
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
		return customerrs.InternalServerError(err)
	}
	defer r.Body.Close()

	ctx := r.Context()
	clientId, err := h.clientService.Link(ctx, reqBody.Name, reqBody.Secret)
	if err != nil {
		if errors.Is(err, services.ErrClientAlreadyExist) {
			return customerrs.ClientAlreadyExist(err)
		}
		return customerrs.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-client-id", clientId)
	w.WriteHeader(http.StatusOK)

	return nil
}

func (h *ClientHandler) Unlink(w http.ResponseWriter, r *http.Request) error {
	clientId := chi.URLParam(r, "clientId")
	if clientId == "" {
		return customerrs.ClientIdIsRequired(ErrEmptyClientId)
	}

	ctx := r.Context()
	if err := h.clientService.Unlink(ctx, clientId); err != nil {
		if errors.Is(err, services.ErrClientNotFound) {
			return customerrs.ClientNotFound(err)
		}
		return customerrs.InternalServerError(err)
	}
	w.Header().Del("x-client-id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}
