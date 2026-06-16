package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"macauth/internal/auth/models"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrTokenNotFound error = errors.New("token not found")
	ErrNoRowsDeleted error = errors.New("no rows deleted")
)

type TokenRepository interface {
	// Get retrieves a Token entity by its id.
	// It returns ErrTokenNotFound if no token are found.
	GetById(ctx context.Context, id string) (*models.Token, error)
	// Get retrieves a Token entity by its refresh token.
	// It returns ErrTokenNotFound if no token are found.
	GetByToken(ctx context.Context, refreshToken string) (*models.Token, error)
	// Create creates a Token entity.
	Create(ctx context.Context, token *models.Token) (string, error)
	// Update updates refresh token in the Token entity.
	Update(ctx context.Context, id, refreshToken string) (string, error)
	// DeleteById removes the Token entity by its id.
	DeleteById(ctx context.Context, id string) error
	// DeleteByToken removes the Token entity by its refresh token.
	DeleteByToken(ctx context.Context, refreshToken string) error
	// DeleteExpired removes expired tokens.
	DeleteExpired(ctx context.Context) error
	// DeleteByUserId removes tokens by user id.
	DeleteByUserId(ctx context.Context, userId string) error
}

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepo(db *sqlx.DB) TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

func (r *tokenRepository) GetById(ctx context.Context, id string) (*models.Token, error) {
	op := "tokenRepository.GetById"
	query := `SELECT id, refresh_token, user_id, client_id, expires_at
	FROM tokens WHERE id = $1
	`
	var token models.Token
	err := r.db.GetContext(ctx, &token, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &token, nil
}

func (r *tokenRepository) GetByToken(ctx context.Context, refreshToken string) (*models.Token, error) {
	op := "tokenRepository.GetByToken"
	query := `SELECT id, refresh_token, user_id, client_id, expires_at
	FROM tokens WHERE refresh_token = $1
	`
	var token models.Token
	err := r.db.GetContext(ctx, &token, query, refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &token, nil
}

func (r *tokenRepository) Create(ctx context.Context, token *models.Token) (string, error) {
	op := "tokenRepository.Create"
	query := `INSERT INTO tokens (id, refresh_token, user_id, client_id, expires_at)
			  VALUES (:id, :refresh_token, :user_id, :client_id, :expires_at)`
	if _, err := r.db.NamedExecContext(ctx, query, token); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token.Id, nil
}

func (r *tokenRepository) Update(ctx context.Context, id, refreshToken string) (string, error) {
	op := "tokenRepository.Update"
	query := `UPDATE tokens SET refresh_token = $1
			WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, refreshToken, id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (r *tokenRepository) DeleteById(ctx context.Context, id string) error {
	op := "tokenRepository.DeleteById"
	query := "DELETE FROM tokens WHERE id = $1"
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *tokenRepository) DeleteByToken(ctx context.Context, refreshToken string) error {
	op := "tokenRepository.DeleteByToken"
	query := "DELETE FROM tokens WHERE refresh_token = $1"
	if _, err := r.db.ExecContext(ctx, query, refreshToken); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	op := "TokenRepository.DeleteExpired"
	query := "DELETE FROM tokens WHERE expires_at < $1"
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ErrNoRowsDeleted)
	}
	return nil
}
func (r *tokenRepository) DeleteByUserId(ctx context.Context, userId string) error {
	op := "TokenRepository.DeleteByUserId"
	query := "DELETE FROM tokens WHERE user_id = $1"
	if _, err := r.db.ExecContext(ctx, query, userId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
