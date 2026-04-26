package errors

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		HandleError(w, r, err)
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr APIError
	reqID := middleware.GetReqID(r.Context())

	if errors.As(err, &apiErr) {
		slog.Error(
			"failed to response",
			slog.String("request id", reqID),
			slog.String("error", apiErr.Error()),
		)
		http.Error(w, apiErr.UserMessage, apiErr.Code)
		return
	}

	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
