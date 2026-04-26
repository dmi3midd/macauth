package api

import (
	errs "macauth/internal/errors"
	"macauth/internal/handlers"
	"macauth/internal/middlewares"
	"macauth/internal/repositories"
	"macauth/internal/services"
	"os"

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

	mainRouter.Group(func(r chi.Router) {
		r.Use(apikeyValidator.Validate())
		r.Post("/", errs.ErrorHandler(clientHandler.Link))
		r.Delete("/{clientId}", errs.ErrorHandler(clientHandler.Unlink))
	})

	mainRouter.Group(func(r chi.Router) {
		r.Use(clientValidator.Validate())
		r.Post("/registration", errs.ErrorHandler(userHandler.Registration))
		r.Post("/login", errs.ErrorHandler(userHandler.Login))
		r.Delete("/logout", errs.ErrorHandler(userHandler.Logout))
		r.Put("/refresh", errs.ErrorHandler(userHandler.Refresh))
	})

	return *mainRouter
}
