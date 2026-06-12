package client

import (
	"errors"
	"macauth/internal/shared/apierror"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyClientSecret = errors.New("x-client-secret is empty")
	ErrEmptyClientHeader = errors.New("x-client-id is empty")
	ErrMiddlewareClientNotFound = errors.New("client not found")
	ErrInvalidSecret     = errors.New("invalid client secret")
)

type ClientValidator struct {
	clientStore ClientRepository
}

func NewClientValidator(clientStore ClientRepository) *ClientValidator {
	return &ClientValidator{
		clientStore: clientStore,
	}
}

func (m *ClientValidator) Validate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientSecret := r.Header.Get("x-client-secret")
			if clientSecret == "" {
				apiErr := apierror.NewBadRequestError(ErrEmptyClientSecret, "Client secret is empty")
				apierror.HandleError(w, r, apiErr)
				return
			}
			clientId := r.Header.Get("x-client-id")
			if clientId == "" {
				apiErr := apierror.NewBadRequestError(ErrEmptyClientHeader, "Client id is empty")
				apierror.HandleError(w, r, apiErr)
				return
			}

			ctx := r.Context()
			candidate, err := m.clientStore.GetById(ctx, clientId)
			if err != nil {
				if errors.Is(err, ErrClientNotFound) {
					apiErr := apierror.NewNotFoundError(ErrMiddlewareClientNotFound, "Client does not exist")
					apierror.HandleError(w, r, apiErr)
					return
				}
				apiErr := apierror.InternalServerError(err)
				apierror.HandleError(w, r, apiErr)
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(candidate.HashedSecret), []byte(clientSecret))
			if err != nil {
				apiErr := apierror.NewBadRequestError(ErrInvalidSecret, "Invalid client secret")
				apierror.HandleError(w, r, apiErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
