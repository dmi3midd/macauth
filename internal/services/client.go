package services

import (
	"context"
	"errors"
	"fmt"
	"macauth/internal/models"
	"macauth/internal/repositories"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrClientAlreadyExist = errors.New("client already exist")
	ErrClientNotFound     = errors.New("client not found")
)

type ClientService interface {
	// Link registrate a new Client.
	// It returns ErrClientAlreadyExist if the client exist ith the name.
	Link(ctx context.Context, name, secret string) (string, error)
	// Unlink removes the client from the Macauth.
	// Also client's sessions will be removed.
	// It returns ErrClientNotFound if no client are found.
	Unlink(ctx context.Context, clientId string) error
}

type clientService struct {
	clientStore repositories.ClientRepository
}

func NewClientService(clientStore repositories.ClientRepository) ClientService {
	return &clientService{
		clientStore: clientStore,
	}
}

func (s *clientService) Link(ctx context.Context, name, secret string) (string, error) {
	op := "clientService.Link"

	candidate, err := s.clientStore.GetByName(ctx, name)
	if err != nil {
		if !errors.Is(err, repositories.ErrClientNotFound) {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}
	if candidate != nil {
		return "", fmt.Errorf("%s: %w", op, ErrClientAlreadyExist)
	}

	id := xid.New().String()
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	clientId, err := s.clientStore.Create(ctx, &models.Client{
		Id:           id,
		Name:         name,
		HashedSecret: string(hashedSecret),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return clientId, nil
}

func (s *clientService) Unlink(ctx context.Context, clientId string) error {
	op := "clientService.Unlink"
	if _, err := s.clientStore.GetById(ctx, clientId); err != nil {
		if errors.Is(err, repositories.ErrClientNotFound) {
			return fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.clientStore.Delete(ctx, clientId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
