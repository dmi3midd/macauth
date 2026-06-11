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
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
)

type UserService interface {
	// Registration performs user registration and returns UserData struct.
	// It returns ErrUserAlreadyExist if the user exist.
	Registration(ctx context.Context, username, email, password, clientId string, isAdmin bool) error
	// Login performs user login and returns LoginResult struct.
	// It returns ErrUserNotFound if no user are found.
	// It returns ErrInvalidPassword if the password is invalid.
	Login(ctx context.Context, email, password, clientId string) (*models.AuthDto, error)
	// Logout performs logout user.
	// Look at TokenService.ValidateRefreshToken for other errors.
	Logout(ctx context.Context, refreshToken string) error
	// Refresh performs refreshing access and refresh tokens.
	// It returns ErrUserNotFound if no user are found.
	// Look at TokenService.ValidateRefreshToken for other errors.
	Refresh(ctx context.Context, refreshToken, clientId string) (*models.AuthDto, error)
	// IsAdmin returns true if the user is an admin.
	IsAdmin(ctx context.Context, userId string) (bool, error)
	// GetPermissions returns the user's permissions.
	GetPermissions(ctx context.Context, userId, clientId string) ([]models.Permission, error)
}

type userService struct {
	userStore       repositories.UserRepository
	tokenService    TokenService
	permissionStore repositories.PermissionRepository
}

func NewUserService(
	userStore repositories.UserRepository,
	tokenService TokenService,
	permissionStore repositories.PermissionRepository,
) UserService {
	return &userService{
		userStore:       userStore,
		tokenService:    tokenService,
		permissionStore: permissionStore,
	}
}

func (s *userService) Registration(ctx context.Context, username, email, password, clientId string, isAdmin bool) error {
	op := "UserService.Registration"

	candidate, err := s.userStore.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}

	if candidate != nil {
		return fmt.Errorf("%s: %w", op, ErrUserAlreadyExist)
	}

	id := xid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	user := &models.User{
		Id:             id,
		Username:       username,
		Email:          email,
		IsAdmin:        isAdmin,
		HashedPassword: string(hashedPassword),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if _, err := s.userStore.Create(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, email, password, clientId string) (*models.AuthDto, error) {
	op := "UserService.Login"

	user, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidPassword)
	}

	userDto := user.NewUserDto()
	tokens, tokenId, err := s.tokenService.GenerateTokens(*userDto, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = s.tokenService.SaveToken(ctx, tokens.RefreshToken, userDto.UserId, clientId, tokenId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.AuthDto{
		ClientId: clientId,
		User:     *userDto,
		Tokens:   *tokens,
	}, nil
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	op := "UserService.Logout"
	tokenId, _, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.tokenService.RemoveToken(ctx, tokenId); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *userService) Refresh(ctx context.Context, refreshToken, clientId string) (*models.AuthDto, error) {
	op := "UserService.Refresh"

	tokenId, userId, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	token, err := s.tokenService.FindToken(ctx, tokenId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if token.ClientId != clientId {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	if err := s.tokenService.RemoveToken(ctx, tokenId); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.userStore.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	userDto := user.NewUserDto()
	tokens, newTokenId, err := s.tokenService.GenerateTokens(*userDto, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := s.tokenService.SaveToken(ctx, tokens.RefreshToken, userId, clientId, newTokenId); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.AuthDto{
		ClientId: clientId,
		User:     *userDto,
		Tokens:   *tokens,
	}, nil
}

func (s *userService) IsAdmin(ctx context.Context, userId string) (bool, error) {
	op := "UserService.IsAdmin"

	user, err := s.userStore.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return user.IsAdmin, nil
}

func (s *userService) GetPermissions(ctx context.Context, userId, clientId string) ([]models.Permission, error) {
	op := "UserService.GetPermissions"
	permissions, err := s.permissionStore.GetPermissionsByUser(ctx, userId, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return permissions, nil
}
