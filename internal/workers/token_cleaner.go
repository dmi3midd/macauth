package workers

import (
	"context"
	"errors"
	"log/slog"
	"macauth/internal/repository"
	"time"
)

type TokenCleaner interface {
	// Start starts the token cleaner worker.
	Start(ctx context.Context)
	// cleanExpiredTokens cleans expired tokens.
	cleanExpiredTokens(ctx context.Context)
}

// TokenCleaner is a worker that cleans expired tokens.
type tokenCleaner struct {
	cleanerInterval time.Duration
	store           repository.TokenRepository
}

func NewTokenCleaner(cleanerInterval time.Duration, store repository.TokenRepository) TokenCleaner {
	return &tokenCleaner{
		cleanerInterval: cleanerInterval,
		store:           store,
	}
}

// Start starts the token cleaner worker.
func (tc *tokenCleaner) Start(ctx context.Context) {
	slog.Info("Starting TokenCleaner worker", slog.Duration("interval", tc.cleanerInterval))

	tc.cleanExpiredTokens(ctx)

	ticker := time.NewTicker(tc.cleanerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping TokenCleaner worker")
			return
		case <-ticker.C:
			tc.cleanExpiredTokens(ctx)
		}
	}
}

// cleanExpiredTokens cleans expired tokens.
func (tc *tokenCleaner) cleanExpiredTokens(ctx context.Context) {
	op := "TokenCleaner.cleanExpiredTokens"
	if err := tc.store.DeleteExpired(ctx); err != nil {
		if errors.Is(err, repository.ErrNoRowsDeleted) {
			slog.Info("No tokens to delete", slog.String("op", op))
			return
		}
		slog.Warn("Error while deleting expired tokens", slog.String("op", op), slog.String("error", err.Error()))
		return
	}
	slog.Info("Expired tokens deleted", slog.String("op", op))
}
