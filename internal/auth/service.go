package auth

import (
	"context"
	"errors"
	"fmt"
	"macauth/internal/token"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrServiceUserNotFound = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
)

type UserService interface {
	// Registration performs user registration and returns UserData struct.
	// It returns ErrUserAlreadyExist if the user exist.
	Registration(ctx context.Context, username, email, password, clientId string, permissions []string) error
	// Login performs user login and returns LoginResult struct.
	// It returns ErrServiceUserNotFound if no user are found.
	// It returns ErrInvalidPassword if the password is invalid.
	Login(ctx context.Context, email, password, clientId string) (*token.AuthDto, error)
	// Logout performs logout user.
	// Look at TokenService.ValidateRefreshToken for other errors.
	Logout(ctx context.Context, refreshToken string) error
	// Refresh performs refreshing access and refresh tokens.
	// It returns ErrServiceUserNotFound if no user are found.
	// Look at TokenService.ValidateRefreshToken for other errors.
	Refresh(ctx context.Context, refreshToken, clientId string) (*token.AuthDto, error)
}

type userService struct {
	userStore       UserRepository
	tokenService    token.TokenService
	permissionStore PermissionRepository
}

func NewUserService(
	userStore UserRepository,
	tokenService token.TokenService,
	permissionStore PermissionRepository,
) UserService {
	return &userService{
		userStore:       userStore,
		tokenService:    tokenService,
		permissionStore: permissionStore,
	}
}

func (u *User) NewUserDto() *token.UserDto {
	return &token.UserDto{
		UserId:      u.Id,
		Username:    u.Username,
		Email:       u.Email,
		Permissions: []string{},
	}
}

func (s *userService) Registration(ctx context.Context, username, email, password, clientId string, permissions []string) error {
	op := "UserService.Registration"

	candidate, err := s.userStore.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
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

	user := &User{
		Id:             id,
		Username:       username,
		Email:          email,
		HashedPassword: string(hashedPassword),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if _, err := s.userStore.Create(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	permModels := make([]Permission, len(permissions))
	for i, perm := range permissions {
		permModels[i] = Permission{
			Id:         xid.New().String(),
			UserId:     id,
			ClientId:   clientId,
			Permission: perm,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}
	if err := s.permissionStore.CreateMany(ctx, permModels); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, email, password, clientId string) (*token.AuthDto, error) {
	op := "UserService.Login"

	user, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrServiceUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidPassword)
	}

	userDto := user.NewUserDto()
	permissions, err := s.permissionStore.GetByUserIdAndClientId(ctx, user.Id, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	userDto.Permissions = permissions

	tokens, tokenId, err := s.tokenService.GenerateTokens(*userDto, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = s.tokenService.SaveToken(ctx, tokens.RefreshToken, userDto.UserId, clientId, tokenId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &token.AuthDto{
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
		if errors.Is(err, token.ErrServiceTokenNotFound) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *userService) Refresh(ctx context.Context, refreshToken, clientId string) (*token.AuthDto, error) {
	op := "UserService.Refresh"

	tokenId, userId, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	foundToken, err := s.tokenService.FindToken(ctx, tokenId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if foundToken.ClientId != clientId {
		return nil, fmt.Errorf("%s: %w", op, token.ErrInvalidRefreshToken)
	}

	if err := s.tokenService.RemoveToken(ctx, tokenId); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.userStore.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrServiceUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	userDto := user.NewUserDto()
	permissions, err := s.permissionStore.GetByUserIdAndClientId(ctx, userId, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	userDto.Permissions = permissions

	tokens, newTokenId, err := s.tokenService.GenerateTokens(*userDto, clientId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := s.tokenService.SaveToken(ctx, tokens.RefreshToken, userId, clientId, newTokenId); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &token.AuthDto{
		ClientId: clientId,
		User:     *userDto,
		Tokens:   *tokens,
	}, nil
}
