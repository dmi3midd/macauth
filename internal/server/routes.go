package server

import (
	"context"
	"macauth/internal/repository"
	"macauth/internal/server/auth"
	"macauth/internal/server/permissions"
	"macauth/internal/server/reset"
	"macauth/internal/service"
	"macauth/internal/shared/apierror"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes(ctx context.Context) *chi.Mux {

	// repositories
	tokenRepo := repository.NewTokenRepo(s.db.GetDB())
	userRepo := repository.NewUserRepo(s.db.GetDB())
	resetRepo := repository.NewResetRepo(s.db.GetDB())
	permissionRepo := repository.NewPermissionRepo(s.db.GetDB())

	// workers
	// cleanerInterval := 1 * time.Hour
	// if s.cfg.TokenCleaner.Interval > 0 {
	// 	cleanerInterval = s.cfg.TokenCleaner.Interval
	// }
	// cleaner := workers.NewTokenCleaner(cleanerInterval, tokenRepo)
	// go cleaner.Start(ctx)

	// services
	tokenService := service.NewTokenService(tokenRepo, &s.cfg.Keys)
	userService := service.NewUserService(userRepo, tokenService)
	passwordResetService := service.NewResetService(resetRepo, userRepo, tokenRepo)
	permissionService := service.NewPermissionService(permissionRepo)

	// handlers
	authHandler := auth.NewUserHandler(userService)
	resetHandler := reset.NewResetHandler(passwordResetService)
	permissionHandler := permissions.NewPermissionHandler(permissionService)

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
		r.Route("/auth", func(r chi.Router) {
			r.Post("/registration", apierror.ErrorHandler(authHandler.Registration))
			r.Post("/login", apierror.ErrorHandler(authHandler.Login))
			r.Delete("/logout", apierror.ErrorHandler(authHandler.Logout))
			r.Put("/refresh", apierror.ErrorHandler(authHandler.Refresh))
		})
		r.Route("/reset", func(r chi.Router) {
			r.Post("/initiate-reset", apierror.ErrorHandler(resetHandler.InitiateReset))
			r.Post("/confirm-reset", apierror.ErrorHandler(resetHandler.ConfirmReset))
		})
		r.Route("/permissions", func(r chi.Router) {
			r.Get("/has", apierror.ErrorHandler(permissionHandler.HasPermission))
			r.Get("/", apierror.ErrorHandler(permissionHandler.GetPermissions))
			r.Post("/", apierror.ErrorHandler(permissionHandler.AddPermissions))
			r.Delete("/", apierror.ErrorHandler(permissionHandler.RemovePermissions))
		})
	})

	return mainRouter
}
