package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ClientRepository interface {
	// GetById retrieves a Client entity by its id.
	// It returns ErrClientNotFound if no client are found.
	GetById(ctx context.Context, clientId string) (*Client, error)
	// GetByName retrieves a Client entity by its name.
	// It returns ErrClientNotFound if no client are found.
	GetByName(ctx context.Context, name string) (*Client, error)
	// Create creates a Client entity and returns its id.
	Create(ctx context.Context, client *Client) (string, error)
	// Delete removes the Client entity from db.
	Delete(ctx context.Context, clientId string) error
}

type clientRepository struct {
	db *sqlx.DB
}

func NewClientRepo(db *sqlx.DB) ClientRepository {
	return &clientRepository{
		db: db,
	}
}

func (r *clientRepository) GetById(ctx context.Context, clientId string) (*Client, error) {
	op := "clientRepository.GetById"
	query := `SELECT id, name, hashed_secret, created_at, updated_at 
	FROM clients WHERE id = $1
	`
	var client Client
	err := r.db.GetContext(ctx, &client, query, clientId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &client, nil
}

func (r *clientRepository) GetByName(ctx context.Context, name string) (*Client, error) {
	op := "clientRepository.GetByName"
	query := `SELECT id, name, hashed_secret, created_at, updated_at 
	FROM clients WHERE name = $1
	`
	var client Client
	err := r.db.GetContext(ctx, &client, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &client, nil
}

func (r *clientRepository) Create(ctx context.Context, client *Client) (string, error) {
	op := "clientRepository.Create"
	query := `INSERT INTO clients
		   (id, name, hashed_secret, created_at, updated_at)
	VALUES (:id, :name, :hashed_secret, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, client); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return client.Id, nil
}

func (r *clientRepository) Delete(ctx context.Context, clientId string) error {
	op := "clientRepository.Delete"
	query := "DELETE FROM clients WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, clientId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
