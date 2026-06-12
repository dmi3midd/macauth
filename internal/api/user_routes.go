package api

import (
	"macauth/internal/auth"
	"macauth/internal/client"
	"macauth/internal/reset"
	"macauth/internal/shared/apierror"
	"macauth/internal/token"

	"github.com/go-chi/chi/v5"
)

func (s *Server) userRouter(
	userHandler *auth.UserHandler,
	tokenHandler *token.TokenHandler,
	resetHandler *reset.ResetHandler,
	clientValidator *client.ClientValidator,
) chi.Router {
	r := chi.NewRouter()

	r.Use(clientValidator.Validate())

	r.Post("/registration", apierror.ErrorHandler(userHandler.Registration))
	r.Post("/login", apierror.ErrorHandler(userHandler.Login))
	r.Delete("/logout", apierror.ErrorHandler(userHandler.Logout))
	r.Put("/refresh", apierror.ErrorHandler(userHandler.Refresh))
	r.Get("/validate", apierror.ErrorHandler(tokenHandler.Validate))
	r.Post("/reset", apierror.ErrorHandler(resetHandler.InitiateReset))
	r.Post("/confirm-reset", apierror.ErrorHandler(resetHandler.ConfirmReset))

	return r
}
