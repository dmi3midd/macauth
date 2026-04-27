package errors

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type AppHandler func(w http.ResponseWriter, r *http.Request) error

func ErrorHandler(fn AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			HandleError(w, r, err)
		}
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

		userErr := UserError{
			Code:      apiErr.Code,
			Message:   apiErr.UserMessage,
			Timestamp: apiErr.Timestamp,
		}

		bytesErr, err := json.Marshal(userErr)
		if err != nil {
			bytesErr = []byte("Internal server error")
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w,
			string(bytesErr),
			apiErr.Code)
		return
	}

	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
