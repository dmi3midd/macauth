package api

import (
	"context"
	"macauth/internal/auth/handlers"
	"macauth/internal/auth/repositories"
	"macauth/internal/auth/services"
	"macauth/internal/auth/workers"
	"macauth/internal/client"
	"macauth/internal/shared/middleware"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
	clientRepo := client.NewClientRepo(s.db.GetDB())
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
	clientService := client.NewClientService(clientRepo)
	passwordResetService := services.NewResetService(resetRepo, userRepo, tokenRepo)

	// handlers
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := client.NewClientHandler(clientService)
	tokenHandler := handlers.NewTokenHandler(tokenService)
	resetHandler := handlers.NewResetHandler(passwordResetService)

	// middlewares
	clientValidator := client.NewClientValidator(clientRepo)
	apikeyValidator := middleware.NewApiKeyValidator(apiKey)

	// routes
	mainRouter := chi.NewRouter()

	mainRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mainRouter.Use(chimiddleware.RequestID)
	mainRouter.Use(chimiddleware.Recoverer)

	mainRouter.Route("/macauth/api/v1", func(r chi.Router) {
		r.Mount("/client", s.clientRouter(clientHandler, tokenHandler, apikeyValidator))
		r.Mount("/user", s.userRouter(userHandler, tokenHandler, resetHandler, clientValidator))
	})

	return mainRouter
}
