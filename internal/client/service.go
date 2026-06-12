package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrClientAlreadyExist = errors.New("client already exist")
	ErrServiceClientNotFound = errors.New("client not found")
)

type ClientService interface {
	// Link registrate a new Client.
	// It returns ErrClientAlreadyExist if the client exist ith the name.
	Link(ctx context.Context, name, secret string) (string, error)
	// Unlink removes the client from the Macauth.
	// Also client's sessions will be removed.
	// It returns ErrServiceClientNotFound if no client are found.
	Unlink(ctx context.Context, clientId string) error
}

type clientService struct {
	clientStore ClientRepository
}

func NewClientService(clientStore ClientRepository) ClientService {
	return &clientService{
		clientStore: clientStore,
	}
}

func (s *clientService) Link(ctx context.Context, name, secret string) (string, error) {
	op := "clientService.Link"

	candidate, err := s.clientStore.GetByName(ctx, name)
	if err != nil {
		if !errors.Is(err, ErrClientNotFound) {
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
	clientId, err := s.clientStore.Create(ctx, &Client{
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
		if errors.Is(err, ErrClientNotFound) {
			return fmt.Errorf("%s: %w", op, ErrServiceClientNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.clientStore.Delete(ctx, clientId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
