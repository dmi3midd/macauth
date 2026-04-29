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
	ErrUserNotFound error = errors.New("user not found")
)

type UserRepository interface {
	// GetById retrieves a User entity by its id.
	// It returns ErrUserNotFound if no user are found.
	GetById(ctx context.Context, userId string) (*models.User, error)
	// GetByEmail retrieves a User entity by its email.
	// It returns ErrUserNotFound if no user are found.
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// Create creates a User entity and returns it.
	Create(ctx context.Context, user *models.User) (string, error)
	// Update updates the User entity.
	Update(ctx context.Context, user *models.User) (string, error)
	// Delete removes the User entity.
	Delete(ctx context.Context, userId string) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetById(ctx context.Context, userId string) (*models.User, error) {
	op := "userRepository.GetById"
	query := `SELECT id, username, email, is_admin, hashed_password, created_at, updated_at
	FROM users WHERE id = $1
	`
	var user models.User
	err := r.db.GetContext(ctx, &user, query, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	op := "userRepository.GetByEmail"
	query := `SELECT id, username, email, is_admin, hashed_password, created_at, updated_at 
	FROM users WHERE email = $1
	`
	var user models.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (string, error) {
	op := "userRepository.Create"
	query := `INSERT INTO users 
		   (id, username, email, is_admin, hashed_password, created_at, updated_at)
	VALUES (:id, :username, :email, :is_admin, :hashed_password, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return user.Id, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) (string, error) {
	op := "userRepository.Update"
	query := `UPDATE users 
	SET username = :username, email = :email, is_admin = :is_admin, hashed_password = :hashed_password, updated_at = :updated_at 
	WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return user.Id, nil
}

func (r *userRepository) Delete(ctx context.Context, userId string) error {
	op := "userRepository.Delete"
	query := "DELETE FROM users WHERE id = $1"
	if _, err := r.db.ExecContext(ctx, query, userId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
