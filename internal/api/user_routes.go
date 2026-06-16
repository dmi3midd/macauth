package api

import (
	"macauth/internal/auth/handlers"
	"macauth/internal/client"
	"macauth/internal/shared/apierror"

	"github.com/go-chi/chi/v5"
)

func (s *Server) userRouter(
	userHandler *handlers.UserHandler,
	tokenHandler *handlers.TokenHandler,
	resetHandler *handlers.ResetHandler,
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
