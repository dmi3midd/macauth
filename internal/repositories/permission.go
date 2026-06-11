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
	ErrPermissionNotFound = fmt.Errorf("permission not found")
)

type PermissionRepository interface {
	// GetPermissionById returns a permission by its ID
	// Returns [ErrPermissionNotFound] if no permission is found
	GetPermissionById(ctx context.Context, id string) (*models.Permission, error)
	// GetPermissionsByUser returns all permissions for a given user and client
	GetPermissionsByUser(ctx context.Context, userId, clientId string) ([]models.Permission, error)
}

type permissionRepository struct {
	db *sqlx.DB
}

func NewPermissionRepo(db *sqlx.DB) PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

func (r *permissionRepository) GetPermissionById(ctx context.Context, id string) (*models.Permission, error) {
	op := "GetPermissionById"
	query := `
	SELECT id, user_id, client_id, permission, created_at, updated_at
	FROM permissions
	WHERE id = $1
	`
	var permission models.Permission
	err := r.db.GetContext(ctx, &permission, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ErrPermissionNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &permission, nil
}

func (r *permissionRepository) GetPermissionsByUser(ctx context.Context, userId, clientId string) ([]models.Permission, error) {
	op := "GetPermissionsByUser"
	query := `
	SELECT id, user_id, client_id, permission, created_at, updated_at
	FROM permissions
	WHERE user_id = $1 AND client_id = $2
	`
	var permissions []models.Permission
	err := r.db.SelectContext(ctx, &permissions, query, userId, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return permissions, nil
}
