package client

import (
	"encoding/json"
	"errors"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var (
	ErrEmptyClientId = errors.New("client id is empty")
)

type ClientHandler struct {
	clientService ClientService
	validate      *validator.Validate
}

func NewClientHandler(clientService ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
		validate:      validator.New(),
	}
}

func (h *ClientHandler) Link(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := httputil.BindAndValidate[LinkRequest](r, h.validate)
	if err != nil {
		return err
	}

	ctx := r.Context()
	clientId, err := h.clientService.Link(ctx, reqBody.Name, reqBody.Secret)
	if err != nil {
		if errors.Is(err, ErrClientAlreadyExist) {
			return apierror.NewConflictError(err, "Client already exist with this name")
		}
		return apierror.InternalServerError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LinkResponse{
		ClientId: clientId,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return apierror.InternalServerError(err)
	}

	return nil
}

func (h *ClientHandler) Unlink(w http.ResponseWriter, r *http.Request) error {
	clientId := chi.URLParam(r, "clientId")
	if clientId == "" {
		return apierror.NewBadRequestError(ErrEmptyClientId, "Client id is not provided")
	}

	ctx := r.Context()
	if err := h.clientService.Unlink(ctx, clientId); err != nil {
		if errors.Is(err, ErrServiceClientNotFound) {
			return apierror.NewNotFoundError(err, "Client does not exist")
		}
		return apierror.InternalServerError(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}
