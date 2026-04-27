package api

import (
	"encoding/json"
	"fmt"
	errs "macauth/internal/errors"
	"macauth/internal/handlers"
	"macauth/internal/middlewares"
	"macauth/internal/repositories"
	"macauth/internal/services"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() chi.Mux {
	// env vars
	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("API_KEY is required!")
	}

	// repositories
	tokenRepo := repositories.NewTokenRepo(s.db.GetDB())
	userRepo := repositories.NewUserRepo(s.db.GetDB())
	clientRepo := repositories.NewClientRepo(s.db.GetDB())

	// services
	tokenService := services.NewTokenService(tokenRepo, s.cfg.Keys)
	userService := services.NewUserService(userRepo, tokenService)
	clientService := services.NewClientService(clientRepo)

	// handlers
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)

	// middlewares
	clientValidator := middlewares.NewClientValidator(clientRepo)
	apikeyValidator := middlewares.NewApiKeyValidator(apiKey)

	// routes
	mainRouter := chi.NewRouter()

	mainRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mainRouter.Use(middleware.RequestID)
	mainRouter.Use(middleware.Recoverer)

	mainRouter.Route("/macauth/api/v1", func(r chi.Router) {
		r.Route("/client", func(r chi.Router) {
			r.Use(apikeyValidator.Validate())
			r.Post("/", errs.ErrorHandler(clientHandler.Link))
			r.Delete("/{clientId}", errs.ErrorHandler(clientHandler.Unlink))
			r.Get("/public-key", errs.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
				key := tokenService.GetPublicKey()
				if err := json.NewEncoder(w).Encode(key); err != nil {
					return errs.InternalServerError(err)
				}
				return nil
			}))
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(clientValidator.Validate())
			r.Post("/registration", errs.ErrorHandler(userHandler.Registration))
			r.Post("/login", errs.ErrorHandler(userHandler.Login))
			r.Delete("/logout", errs.ErrorHandler(userHandler.Logout))
			r.Put("/refresh", errs.ErrorHandler(userHandler.Refresh))
			r.Get("/validate", errs.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
				authHeader := r.Header.Get("Authorization")
				token := ""
				if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
					token = after
				}
				if token == "" {
					return errs.NewUnauthorizedError(
						fmt.Errorf("Invalid or empty Authorization header"),
						"Invalid or empty Authorization header",
					)
				}
				userData, _, err := tokenService.ValidateAccessToken(token)
				if err != nil {
					return errs.NewUnauthorizedError(err, "Invalid access token")
				}
				if err := json.NewEncoder(w).Encode(userData); err != nil {
					return errs.InternalServerError(err)
				}
				return nil
			}))
		})
	})

	return *mainRouter
}
