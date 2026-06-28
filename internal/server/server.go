package server

import (
	"context"
	"net/http"

	_ "github.com/joho/godotenv/autoload"

	"macauth/internal/config"
	"macauth/internal/database"
)

type Server struct {
	db  database.DBService
	cfg *config.Config
}

func NewServer(ctx context.Context, cfg *config.Config, db database.DBService) *http.Server {
	s := &Server{
		db:  db,
		cfg: cfg,
	}

	router := s.RegisterRoutes(ctx)
	return &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
	}
}
