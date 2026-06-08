package workers

import (
	"context"
	"errors"
	"log/slog"
	"macauth/internal/repositories"
	"sync"
	"time"
)

type TokenCleaner struct {
	cleanerInterval time.Duration
	store           repositories.TokenRepository
	stopOnce        sync.Once
}

func NewTokenCleaner(cleanerInterval time.Duration, store repositories.TokenRepository) *TokenCleaner {
	return &TokenCleaner{
		cleanerInterval: cleanerInterval,
		store:           store,
	}
}

func (tc *TokenCleaner) Start(ctx context.Context) {
	slog.Info("Starting TokenCleaner worker", slog.Duration("interval", tc.cleanerInterval))

	// Clean immediately on startup
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

func (tc *TokenCleaner) Stop() {
	tc.stopOnce.Do(func() {
		// Perform any cleanup here
	})
}
