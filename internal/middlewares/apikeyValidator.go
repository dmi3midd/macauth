package middlewares

import (
	"errors"
	"net/http"

	errs "macauth/internal/errors"
)

var (
	ErrInvalidApiKey = errors.New("invalid api key")
)

type ApiKeyValidator struct {
	apiKey string
}

func NewApiKeyValidator(apiKey string) *ApiKeyValidator {
	return &ApiKeyValidator{
		apiKey: apiKey,
	}
}

func (m *ApiKeyValidator) Validate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("x-api-key")
			if key != m.apiKey {
				apiErr := errs.NewForbiddenError(
					ErrInvalidApiKey,
					"Invalid API key",
				)
				errs.HandleError(w, r, apiErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
