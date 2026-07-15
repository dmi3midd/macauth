package service

import (
	"context"
	"errors"
	"fmt"
	"macauth/internal/domain"
	"macauth/internal/repository"
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
	// Returns [ErrUserAlreadyExist] if the user exist.
	Registration(ctx context.Context, username, email, password string) error
	// Login performs user login and returns LoginResult struct.
	// Returns [ErrUserNotFound] if no user are found.
	// Returns [ErrInvalidPassword] if the password is invalid.
	Login(ctx context.Context, email, password string) (*domain.AuthDto, error)
	// Logout performs logout user.
	// Look at TokenService.ValidateRefreshToken for other errors.
	Logout(ctx context.Context, refreshToken string) error
	// Refresh performs refreshing access and refresh tokens.
	// Returns [ErrUserNotFound] if no user are found.
	// Look at TokenService.ValidateRefreshToken for other errors.
	Refresh(ctx context.Context, refreshToken string) (*domain.AuthDto, error)
	// Validate validates access token and returns User data.
	// Look at TokenService.ValidateAccessToken for other errors.
	Validate(ctx context.Context, accessToken string) (*domain.UserDto, error)
}

type userService struct {
	userStore    repository.UserRepository
	tokenManager TokenManager
}

func NewUserService(
	userStore repository.UserRepository,
	tokenManager TokenManager,
) UserService {
	return &userService{
		userStore:    userStore,
		tokenManager: tokenManager,
	}
}

func (s *userService) Registration(ctx context.Context, username, email, password string) error {
	op := "UserService.Registration"

	candidate, err := s.userStore.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
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

	user := &domain.User{
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

	return nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.AuthDto, error) {
	op := "UserService.Login"

	user, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidPassword)
	}

	userDto := user.ToUserDto()
	tokens, tokenId, err := s.tokenManager.GenerateTokens(*userDto)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = s.tokenManager.SaveToken(ctx, tokens.RefreshToken, userDto.UserId, tokenId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.AuthDto{
		User:   *userDto,
		Tokens: domain.TokensPair(*tokens),
	}, nil
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	op := "UserService.Logout"
	tokenId, _, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.tokenManager.RemoveToken(ctx, tokenId); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *userService) Refresh(ctx context.Context, refreshToken string) (*domain.AuthDto, error) {
	op := "UserService.Refresh"

	tokenId, userId, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.tokenManager.RemoveToken(ctx, tokenId); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.userStore.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	userDto := user.ToUserDto()
	tokens, newTokenId, err := s.tokenManager.GenerateTokens(*userDto)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := s.tokenManager.SaveToken(ctx, tokens.RefreshToken, userId, newTokenId); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.AuthDto{
		User:   *userDto,
		Tokens: domain.TokensPair(*tokens),
	}, nil
}

func (s *userService) Validate(ctx context.Context, accessToken string) (*domain.UserDto, error) {
	op := "UserService.Validate"
	userDto, tokenId, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if _, err := s.tokenManager.FindToken(ctx, tokenId); err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidAccessToken)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return userDto, nil
}
