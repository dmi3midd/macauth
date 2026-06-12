package api

import (
	"macauth/internal/client"
	"macauth/internal/shared/apierror"
	"macauth/internal/shared/middleware"
	"macauth/internal/token"

	"github.com/go-chi/chi/v5"
)

func (s *Server) clientRouter(
	clientHandler *client.ClientHandler,
	tokenHandler *token.TokenHandler,
	apikeyValidator *middleware.ApiKeyValidator,
) chi.Router {
	r := chi.NewRouter()

	r.Use(apikeyValidator.Validate())

	r.Post("/", apierror.ErrorHandler(clientHandler.Link))
	r.Delete("/{clientId}", apierror.ErrorHandler(clientHandler.Unlink))
	r.Get("/public-key", apierror.ErrorHandler(tokenHandler.GetPublicKey))

	return r
}
