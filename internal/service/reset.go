package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"macauth/internal/domain"
	"macauth/internal/repository"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTokenExpired error = errors.New("token expired")
	ErrTokenUsed    error = errors.New("token used")
	ErrInvalidToken error = errors.New("invalid token")
)

type ResetService interface {
	// InitiateReset creates a password reset entity for the given email and returns a reset token.
	// Returns [ErrUserNotFound] if no user is found with the given email.
	InitiateReset(ctx context.Context, email string) (*ResetToken, error)
	// ConfirmReset validates the reset token and updates the user's password.
	// Returns [ErrTokenUsed] if the token has already been used.
	// Returns [ErrTokenExpired] if the token has expired.
	// Returns [ErrInvalidToken] if the token is invalid.
	// Returns [ErrUserNotFound] if no user is found with the given email.
	ConfirmReset(ctx context.Context, token string, newPassword string) error
}

type resetService struct {
	resetStore repository.ResetRepository
	userStore  repository.UserRepository
	tokenStore repository.TokenRepository
}

func NewResetService(resetStore repository.ResetRepository, userStore repository.UserRepository, tokenStore repository.TokenRepository) ResetService {
	return &resetService{
		resetStore: resetStore,
		userStore:  userStore,
		tokenStore: tokenStore,
	}
}

type ResetToken struct {
	ResetToken string
	Email      string
}

func (s *resetService) InitiateReset(ctx context.Context, email string) (*ResetToken, error) {
	op := "ResetService.InitiateReset"
	candidate, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resetToken := rand.Text()
	hashedToken := sha256.Sum256([]byte(resetToken))

	resetEntity := &domain.PasswordReset{
		Id:        xid.New().String(),
		UserId:    candidate.Id,
		TokenHash: hex.EncodeToString(hashedToken[:]),
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}

	_, err = s.resetStore.Create(ctx, resetEntity)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ResetToken{
		ResetToken: resetToken,
		Email:      email,
	}, nil
}

func (s *resetService) ConfirmReset(ctx context.Context, tokenStr string, newPassword string) error {
	op := "ResetService.ConfirmReset"

	hashedToken := sha256.Sum256([]byte(tokenStr))
	candidateToken, err := s.resetStore.FindValidByTokenHash(ctx, hex.EncodeToString(hashedToken[:]))
	if err != nil {
		if errors.Is(err, repository.ErrResetNotFound) {
			return fmt.Errorf("%s: %w", op, ErrInvalidToken)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if candidateToken.UsedAt != nil {
		return fmt.Errorf("%s: %w", op, ErrTokenUsed)
	}

	if candidateToken.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("%s: %w", op, ErrTokenExpired)
	}

	candidateUser, err := s.userStore.GetById(ctx, candidateToken.UserId)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	candidateUser.HashedPassword = string(newHashedPassword)
	candidateUser.UpdatedAt = time.Now()

	if _, err = s.userStore.Update(ctx, candidateUser); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	newUsedAt := time.Now()
	candidateToken.UsedAt = &newUsedAt
	if err = s.resetStore.Update(ctx, candidateToken); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.tokenStore.DeleteByUserId(ctx, candidateUser.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
