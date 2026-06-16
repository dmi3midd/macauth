package workers

import (
	"context"
	"errors"
	"log/slog"
	"macauth/internal/auth/repositories"
	"time"
)

// TokenCleaner is a worker that cleans expired tokens.
type TokenCleaner struct {
	cleanerInterval time.Duration
	store           repositories.TokenRepository
}

func NewTokenCleaner(cleanerInterval time.Duration, store repositories.TokenRepository) *TokenCleaner {
	return &TokenCleaner{
		cleanerInterval: cleanerInterval,
		store:           store,
	}
}

// Start starts the token cleaner worker.
func (tc *TokenCleaner) Start(ctx context.Context) {
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
func (tc *TokenCleaner) cleanExpiredTokens(ctx context.Context) {
	op := "TokenCleaner.cleanExpiredTokens"
	if err := tc.store.DeleteExpired(ctx); err != nil {
		if errors.Is(err, repositories.ErrNoRowsDeleted) {
			slog.Info("No tokens to delete", slog.String("op", op))
			return
		}
		slog.Warn("Error while deleting expired tokens", slog.String("op", op), slog.String("error", err.Error()))
		return
	}
	slog.Info("Expired tokens deleted", slog.String("op", op))
}
