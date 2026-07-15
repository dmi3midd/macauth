package service

import (
	"context"
	"errors"
	"fmt"
	"macauth/internal/config"
	"macauth/internal/domain"
	"macauth/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
)

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrInvalidAccessToken      = errors.New("invalid access token")
	ErrSubjectAndIDNotFound    = errors.New("subject and id not found")
	ErrTokenNotFound           = errors.New("token not found")
)

type TokenManager interface {
	// GenerateTokens generates pair with access and refresh tokens and token id (TokensPair, tokenId, error).
	GenerateTokens(user domain.UserDto) (*domain.TokensPair, string, error)
	// ValidateRefreshToken validates refresh token and returns token and user id (tokenId, userId, error).
	// Returns ("", "", error) if validation go wrong.
	// Returns [ErrUnexpectedSigningMethod] if the token uses an unexpected signing method.
	// Returns [ErrInvalidRefreshToken] if the token is invalid.
	// Returns [ErrSubjectAndIDNotFound] if subject or token ID are not found in claims.
	ValidateRefreshToken(refreshToken string) (string, string, error)
	// ValidateAccessToken validates access token and returns userDto and token id (userDto, tokenId, error).
	// Returns (nil, "", error) if validation go wrong.
	// Returns [ErrUnexpectedSigningMethod] if the token uses an unexpected signing method.
	// Returns [ErrInvalidAccessToken] if the token is invalid.
	// Returns [ErrSubjectAndIDNotFound] if subject or token ID are not found in claims.
	ValidateAccessToken(accessToken string) (*domain.UserDto, string, error)
	// SaveToken creates refresh token for the user.
	SaveToken(ctx context.Context, refreshToken, userId, tokenId string) (string, error)
	// RemoveToken removes refresh token.
	// Returns [ErrTokenNotFound] if no token are found.
	RemoveToken(ctx context.Context, id string) error
	// FindToken finds and returns a Token entity by its refresh token string.
	// Returns [ErrTokenNotFound] if no token are found.
	FindToken(ctx context.Context, id string) (*domain.Token, error)
}

type tokenManager struct {
	tokenStore repository.TokenRepository
	jwt        *config.JWT
	keys       config.KeysPair
}

func NewTokenManager(tokenStore repository.TokenRepository, cfg *config.JWT, keys *config.KeysPair) TokenManager {
	return &tokenManager{
		tokenStore: tokenStore,
		jwt:        cfg,
		keys:       *keys,
	}
}

func (s *tokenManager) GenerateTokens(user domain.UserDto) (*domain.TokensPair, string, error) {
	op := "TokenManager.GenerateTokens"
	accessExpiry := s.jwt.AccessTokenTTL
	refreshExpiry := s.jwt.RefreshTokenTTL
	now := time.Now()
	id := xid.New().String()

	// Access token
	accessClaims := domain.AccessClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Issuer:    "macauth",
			Subject:   user.UserId,
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(s.keys.PrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Refresh token
	refreshClaims := jwt.RegisteredClaims{
		ID:        id,
		Issuer:    "macauth",
		Subject:   user.UserId,
		ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(s.keys.PrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	return &domain.TokensPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, id, nil
}

func (s *tokenManager) ValidateRefreshToken(refreshToken string) (string, string, error) {
	op := "TokenManager.ValidateRefreshToken"
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

func (s *tokenManager) ValidateAccessToken(accessToken string) (*domain.UserDto, string, error) {
	op := "TokenManager.ValidateAccessToken"
	claims := &domain.AccessClaims{}
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

	return &domain.UserDto{
		UserId:   userId,
		Username: claims.User.Username,
		Email:    claims.User.Email,
	}, tokenId, nil
}

func (s *tokenManager) SaveToken(ctx context.Context, refreshToken, userId, tokenId string) (string, error) {
	op := "TokenManager.SaveToken"

	claims := &jwt.RegisteredClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(refreshToken, claims)
	if err != nil {
		return "", fmt.Errorf("%s: failed to parse refresh token: %w", op, err)
	}

	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	} else {
		// fallback to 14 days if not present
		expiresAt = time.Now().Add(336 * time.Hour)
	}

	token := domain.Token{
		Id:           tokenId,
		RefreshToken: refreshToken,
		UserId:       userId,
		ExpiresAt:    expiresAt,
	}
	id, err := s.tokenStore.Create(ctx, &token)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *tokenManager) RemoveToken(ctx context.Context, id string) error {
	op := "TokenManager.RemoveToken"
	if err := s.tokenStore.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *tokenManager) FindToken(ctx context.Context, id string) (*domain.Token, error) {
	op := "TokenManager.FindToken"
	token, err := s.tokenStore.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}
