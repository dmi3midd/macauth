package api

import (
	"context"
	errs "macauth/internal/errors"
	"macauth/internal/handlers"
	"macauth/internal/middlewares"
	"macauth/internal/repositories"
	"macauth/internal/services"
	"macauth/internal/workers"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes(ctx context.Context) *chi.Mux {
	// env vars
	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("API_KEY is required!")
	}

	// repositories
	tokenRepo := repositories.NewTokenRepo(s.db.GetDB())
	userRepo := repositories.NewUserRepo(s.db.GetDB())
	clientRepo := repositories.NewClientRepo(s.db.GetDB())
	resetRepo := repositories.NewResetRepo(s.db.GetDB())
	permissionRepo := repositories.NewPermissionRepo(s.db.GetDB())

	// workers
	cleanerInterval := 1 * time.Hour
	if s.cfg.TokenCleaner.Interval > 0 {
		cleanerInterval = s.cfg.TokenCleaner.Interval
	}
	cleaner := workers.NewTokenCleaner(cleanerInterval, tokenRepo)
	go cleaner.Start(ctx)

	// services
	tokenService := services.NewTokenService(tokenRepo, &s.cfg.Keys)
	userService := services.NewUserService(userRepo, tokenService, permissionRepo)
	clientService := services.NewClientService(clientRepo)
	passwordResetService := services.NewResetService(resetRepo, userRepo, tokenRepo)

	// handlers
	userHandler := handlers.NewUserHandler(userService, passwordResetService)
	clientHandler := handlers.NewClientHandler(clientService)
	tokenHandler := handlers.NewTokenHandler(tokenService)

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
			r.Get("/public-key", errs.ErrorHandler(tokenHandler.GetPublicKey))
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(clientValidator.Validate())
			r.Post("/registration", errs.ErrorHandler(userHandler.Registration))
			r.Post("/login", errs.ErrorHandler(userHandler.Login))
			r.Delete("/logout", errs.ErrorHandler(userHandler.Logout))
			r.Put("/refresh", errs.ErrorHandler(userHandler.Refresh))
			r.Get("/validate", errs.ErrorHandler(tokenHandler.Validate))
			r.Get("/is-admin/{userId}", errs.ErrorHandler(userHandler.IsAdmin))
			r.Post("/reset", errs.ErrorHandler(userHandler.InitiateReset))
			r.Post("/confirm-reset", errs.ErrorHandler(userHandler.ConfirmReset))
		})
	})

	return mainRouter
}
