package repositories

import (
	"context"
	"fmt"
	"macauth/internal/auth/models"

	"github.com/jmoiron/sqlx"
)

type PermissionRepository interface {
	GetByUserIdAndClientId(ctx context.Context, userId, clientId string) ([]string, error)
	CreateMany(ctx context.Context, permissions []models.Permission) error
	DeleteAllForUserAndClient(ctx context.Context, userId, clientId string) error
}

type permissionRepository struct {
	db *sqlx.DB
}

func NewPermissionRepo(db *sqlx.DB) PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

func (r *permissionRepository) GetByUserIdAndClientId(ctx context.Context, userId, clientId string) ([]string, error) {
	op := "PermissionRepository.GetByUserIdAndClientId"
	query := `
	SELECT permission 
	FROM permissions 
	WHERE user_id = $1 AND client_id = $2`

	var permissions []string
	err := r.db.SelectContext(ctx, &permissions, query, userId, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if permissions == nil {
		permissions = []string{}
	}

	return permissions, nil
}

func (r *permissionRepository) CreateMany(ctx context.Context, permissions []models.Permission) error {
	op := "PermissionRepository.CreateMany"
	if len(permissions) == 0 {
		return nil
	}

	query := `
	INSERT INTO permissions (id, user_id, client_id, permission, created_at, updated_at)
	VALUES (:id, :user_id, :client_id, :permission, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, permissions)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *permissionRepository) DeleteAllForUserAndClient(ctx context.Context, userId, clientId string) error {
	op := "PermissionRepository.DeleteAllForUserAndClient"
	query := `
	DELETE FROM permissions 
	WHERE user_id = $1 AND client_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, userId, clientId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
