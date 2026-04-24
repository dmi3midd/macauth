package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"macauth/internal/models"

	"github.com/jmoiron/sqlx"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ClientRepository interface {
	// GetById retrieves a Client entity by its id.
	// It returns ErrClientNotFound if no client are found.
	GetById(ctx context.Context, clientId string) (*models.Client, error)
	// GetByName retrieves a Client entity by its name.
	// It returns ErrClientNotFound if no client are found.
	GetByName(ctx context.Context, name string) (*models.Client, error)
	// Create creates a Client entity and returns its id.
	Create(ctx context.Context, client *models.Client) (*string, error)
	// Delete removes the Client entity from db.
	Delete(ctx context.Context, clientId string) error
}

type clientReposiroty struct {
	db *sqlx.DB
}

func NewClientRepo(db *sqlx.DB) ClientRepository {
	return &clientReposiroty{
		db: db,
	}
}

func (r *clientReposiroty) GetById(ctx context.Context, clientId string) (*models.Client, error) {
	op := "clientRepository.GetById"
	query := `SELECT id, name, hashed_secret, created_at, updated_at 
	FROM clients WHERE id = $1
	`
	var client models.Client
	err := r.db.GetContext(ctx, &client, query, clientId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &client, nil
}

func (r *clientReposiroty) GetByName(ctx context.Context, name string) (*models.Client, error) {
	op := "clientRepository.GetByName"
	query := `SELECT id, name, hashed_secret, created_at, updated_at 
	FROM clients WHERE name = $1
	`
	var client models.Client
	err := r.db.GetContext(ctx, &client, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &client, nil
}

func (r *clientReposiroty) Create(ctx context.Context, client *models.Client) (*string, error) {
	op := "clientRepository.Create"
	query := `INSERT INTO clients
		   (id, name, hashed_secret, created_at, updated_at)
	VALUES (:id, :name, :hashed_secret, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, client); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &client.Id, nil
}

func (r *clientReposiroty) Delete(ctx context.Context, clientId string) error {
	op := "clientRepository.Delete"
	query := "DELETE FROM clients WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, clientId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
