package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"macauth/internal/domain"
	"macauth/internal/repository"

	"github.com/rs/xid"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
	ErrEmptyPermissions   = errors.New("empty permissions")
	ErrTooManyPermissions = errors.New("too many permissions")
)

type PermissionService interface {
	// HasPermissions checks if a user has ALL the specified permissions.
	// Returns [ErrEmptyPermissions] if the permissions slice is empty.
	// Returns [ErrTooManyPermissions] if the permissions slice is too large (more than 16).
	HasPermissions(ctx context.Context, userId, clientId string, permissions []string) (bool, error)
	// GetPermissions returns permissions for a user and client.
	// Returns an empty slice if the user has no permissions.
	GetPermissions(ctx context.Context, userId, clientId string) ([]string, error)
	// AddPermissions adds permissions for a user and client.
	// Returns [ErrEmptyPermissions] if the permissions slice is empty.
	// Returns [ErrTooManyPermissions] if the permissions slice is too large (more than 16).
	AddPermissions(ctx context.Context, userId, clientId string, permissions []string) error
	// RemovePermissions removes permissions for a user and client.
	// Returns [ErrEmptyPermissions] if the permissions slice is empty.
	// Returns [ErrTooManyPermissions] if the permissions slice is too large (more than 16).
	RemovePermissions(ctx context.Context, userId, clientId string, permissions []string) error
}

type permissionService struct {
	permissionStore repository.PermissionRepository
}

func NewPermissionService(permissionStore repository.PermissionRepository) PermissionService {
	return &permissionService{
		permissionStore: permissionStore,
	}
}

func (s *permissionService) HasPermissions(ctx context.Context, userId string, clientId string, permissions []string) (bool, error) {
	op := "PermissionService.HasPermissions"

	if len(permissions) == 0 {
		return false, fmt.Errorf("%s: %w", op, ErrEmptyPermissions)
	}
	if len(permissions) > 16 {
		return false, fmt.Errorf("%s: %w", op, ErrTooManyPermissions)
	}

	permissionsFromDb, err := s.permissionStore.GetByUserIdAndClientId(ctx, userId, clientId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if len(permissionsFromDb) == 0 {
		return false, nil
	}

	permSet := make(map[string]bool, len(permissions))
	for _, p := range permissions {
		permSet[p] = true
	}

	for _, p := range permissionsFromDb {
		if _, ok := permSet[p.Permission]; !ok {
			return false, nil
		}
	}

	return true, nil
}

func (s *permissionService) GetPermissions(ctx context.Context, userId string, clientId string) ([]string, error) {
	op := "PermissionService.GetPermissions"

	permissions, err := s.permissionStore.GetByUserIdAndClientId(ctx, userId, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(permissions) == 0 {
		return []string{}, nil
	}
	result := make([]string, 0, len(permissions))
	for _, p := range permissions {
		result = append(result, p.Permission)
	}

	return result, nil
}

func (s *permissionService) AddPermissions(ctx context.Context, userId string, clientId string, permissions []string) error {
	op := "PermissionService.AddPermissions"

	if len(permissions) == 0 {
		return fmt.Errorf("%s: %w", op, ErrEmptyPermissions)
	}
	if len(permissions) > 16 {
		return fmt.Errorf("%s: %w", op, ErrTooManyPermissions)
	}
	newPermissions := make([]domain.Permission, len(permissions))
	for i, permission := range permissions {
		now := time.Now()
		newPermissions[i] = domain.Permission{
			Id:         xid.New().String(),
			UserId:     userId,
			ClientId:   clientId,
			Permission: permission,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}
	if err := s.permissionStore.CreateMany(ctx, newPermissions); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *permissionService) RemovePermissions(ctx context.Context, userId string, clientId string, permissions []string) error {
	op := "PermissionService.RemovePermissions"

	if len(permissions) == 0 {
		return fmt.Errorf("%s: %w", op, ErrEmptyPermissions)
	}
	if len(permissions) > 16 {
		return fmt.Errorf("%s: %w", op, ErrTooManyPermissions)
	}

	if err := s.permissionStore.DeleteMany(ctx, userId, clientId, permissions); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
