package middlewares

import (
	"errors"
	errs "macauth/internal/errors"
	"macauth/internal/repositories"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyClientSecret = errors.New("x-client-secret is empty")
	ErrEmptyClientId     = errors.New("x-client-id is empty")
	ErrClientNotFound    = errors.New("client not found")
	ErrInvalidSecret     = errors.New("invalid client secret")
)

type ClientValidator struct {
	clientStore repositories.ClientRepository
}

func NewClientValidator(clientStore repositories.ClientRepository) *ClientValidator {
	return &ClientValidator{
		clientStore: clientStore,
	}
}

func (m *ClientValidator) Validate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientSecret := r.Header.Get("x-client-secret")
			if clientSecret == "" {
				apiErr := errs.NewBadRequestError(ErrEmptyClientSecret, "Client secret is empty")
				errs.HandleError(w, r, apiErr)
				return
			}
			clientId := r.Header.Get("x-client-id")
			if clientId == "" {
				apiErr := errs.NewBadRequestError(ErrEmptyClientId, "Client id is empty")
				errs.HandleError(w, r, apiErr)
				return
			}

			ctx := r.Context()
			candidate, err := m.clientStore.GetById(ctx, clientId)
			if err != nil {
				if errors.Is(err, repositories.ErrClientNotFound) {
					apiErr := errs.NewNotFoundError(ErrClientNotFound, "Client does not exist")
					errs.HandleError(w, r, apiErr)
					return
				}
				apiErr := errs.InternalServerError(err)
				errs.HandleError(w, r, apiErr)
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(candidate.HashedSecret), []byte(clientSecret))
			if err != nil {
				apiErr := errs.NewBadRequestError(ErrInvalidSecret, "Invalid client secret")
				errs.HandleError(w, r, apiErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
