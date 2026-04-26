package services

import (
	"context"
	"errors"
	"fmt"
	"macauth/internal/config"
	"macauth/internal/models"
	"macauth/internal/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrInvalidAccessToken      = errors.New("invalid access token")
	ErrSubjectAndIDNotFound    = errors.New("subject and id not found")
	ErrTokenNotFound           = errors.New("token not found")
)

type TokenService interface {
	// GenerateTokens generates pair with access and refresh tokens.
	GenerateTokens(user models.UserDto, clientId, tokenId string) (*models.TokensPair, error)
	// ValidateRefreshToken validates refresh token and returns token and user id (tokenId, userId, error).
	// It returns ("", "", error) if validation go wrong.
	// It returns ErrUnexpectedSigningMethod if the token uses an unexpected signing method.
	// It returns ErrInvalidRefreshToken if the token is invalid.
	// It returns ErrSubjectAndIDNotFound if subject or token ID are not found in claims.
	ValidateRefreshToken(refreshToken string) (string, string, error)
	// ValidateAccessToken validates access token and returns userDto and token id.
	// It returns (nil, "", error) if validation go wrong.
	// It returns ErrUnexpectedSigningMethod if the token uses an unexpected signing method.
	// It returns ErrInvalidAccessToken if the token is invalid.
	// It returns ErrSubjectAndIDNotFound if subject or token ID are not found in claims.
	ValidateAccessToken(refreshToken string) (*models.UserDto, string, error)
	// SaveToken creates refresh token for the user.
	SaveToken(ctx context.Context, refreshToken, userId, clientId, tokenId string) (string, error)
	// RemoveToken removes refresh token.
	// It returns ErrTokenNotFound if no token are found.
	RemoveToken(ctx context.Context, id string) error
	// FindToken finds and returns a Token entity by its refresh token string.
	// It returns ErrTokenNotFound if no token are found.
	FindToken(ctx context.Context, id string) (*models.Token, error)
}

type tokenService struct {
	tokenStore repositories.TokenRepository
	keys       config.KeysPair
}

func NewTokenService(tokenStore repositories.TokenRepository, keys *config.KeysPair) TokenService {
	return &tokenService{
		tokenStore: tokenStore,
		keys:       *keys,
	}
}

func (s *tokenService) GenerateTokens(user models.UserDto, clientId, tokenId string) (*models.TokensPair, error) {
	op := "tokenService.GenerateTokens"
	accessExpiry, _ := time.ParseDuration("30m")
	refreshExpiry, _ := time.ParseDuration("336h")
	now := time.Now()

	// Access token
	accessClaims := models.AccessClaims{
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "macauth",
			Subject:   user.UserId,
			Audience:  jwt.ClaimStrings{clientId},
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(s.keys.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Refresh token
	refreshClaims := jwt.RegisteredClaims{
		ID:        tokenId,
		Issuer:    "macauth",
		Subject:   user.UserId,
		Audience:  jwt.ClaimStrings{clientId},
		ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(s.keys.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.TokensPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *tokenService) ValidateRefreshToken(refreshToken string) (string, string, error) {
	op := "tokenService.ValidateRefreshToken"
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%s: %w %v", op, ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return s.keys.PublicKey, nil
	})

	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if !token.Valid {
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	userId := claims.Subject
	tokenId := claims.ID

	if userId == "" || tokenId == "" {
		return "", "", fmt.Errorf("%s: %w", op, ErrSubjectAndIDNotFound)
	}

	return tokenId, userId, nil
}

func (s *tokenService) ValidateAccessToken(accessToken string) (*models.UserDto, string, error) {
	op := "tokenService.ValidateAccessToken"
	claims := &models.AccessClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%s: %w %v", op, ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return s.keys.PublicKey, nil
	})

	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if !token.Valid {
		return nil, "", fmt.Errorf("%s: %w", op, ErrInvalidAccessToken)
	}

	userId := claims.Subject
	tokenId := claims.ID

	if userId == "" || tokenId == "" {
		return nil, "", fmt.Errorf("%s: %w", op, ErrSubjectAndIDNotFound)
	}

	return &models.UserDto{
		UserId:   userId,
		Username: claims.Username,
		Email:    claims.Email,
	}, tokenId, nil
}

// TODO: review 3 methods below

func (s *tokenService) SaveToken(ctx context.Context, refreshToken, userId, clientId, tokenId string) (string, error) {
	op := "tokenService.SaveToken"
	token := models.Token{
		Id:           tokenId,
		RefreshToken: refreshToken,
		UserId:       userId,
		ClientId:     clientId,
	}
	id, err := s.tokenStore.Create(ctx, &token)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *tokenService) RemoveToken(ctx context.Context, id string) error {
	op := "tokenService.RemoveToken"
	if _, err := s.tokenStore.GetById(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrTokenNotFound) {
			return fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.tokenStore.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *tokenService) FindToken(ctx context.Context, id string) (*models.Token, error) {
	op := "tokenService.FindToken"
	token, err := s.tokenStore.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrTokenNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}
