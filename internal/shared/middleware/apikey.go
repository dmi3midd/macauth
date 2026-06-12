package middleware

import (
	"errors"
	"macauth/internal/shared/apierror"
	"net/http"
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
				apiErr := apierror.NewForbiddenError(
					ErrInvalidApiKey,
					"Invalid API key",
				)
				apierror.HandleError(w, r, apiErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
