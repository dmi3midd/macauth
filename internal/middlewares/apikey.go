package middlewares

import (
	"errors"
	"net/http"

	errs "macauth/internal/errors"
)

var (
	ErrInvalidApiKey = errors.New("invalid api key")
)

type ApiKeyValidationMw struct {
	ApiKey string
}

func NewApiKeyValidationMw(apiKey string) *ApiKeyValidationMw {
	return &ApiKeyValidationMw{
		ApiKey: apiKey,
	}
}

func (m *ApiKeyValidationMw) ApiKeyValidate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("x-api-key")
			if key != m.ApiKey {
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
