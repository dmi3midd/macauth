package service

import (
	"context"
	"crypto/rsa"
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

type TokenService interface {
	// GenerateTokens generates pair with access and refresh tokens and token id (TokensPair, tokenId, error).
	GenerateTokens(user domain.UserDto, clientId string) (*domain.TokensPair, string, error)
	// ValidateRefreshToken validates refresh token and returns token and user id (tokenId, userId, error).
	// It returns ("", "", error) if validation go wrong.
	// It returns [ErrUnexpectedSigningMethod] if the token uses an unexpected signing method.
	// It returns [ErrInvalidRefreshToken] if the token is invalid.
	// It returns [ErrSubjectAndIDNotFound] if subject or token ID are not found in claims.
	ValidateRefreshToken(refreshToken string) (string, string, error)
	// ValidateAccessToken validates access token and returns userDto and token id (userDto, tokenId, error).
	// It returns (nil, "", error) if validation go wrong.
	// It returns [ErrUnexpectedSigningMethod] if the token uses an unexpected signing method.
	// It returns [ErrInvalidAccessToken] if the token is invalid.
	// It returns [ErrSubjectAndIDNotFound] if subject or token ID are not found in claims.
	ValidateAccessToken(accessToken string) (*domain.UserDto, string, error)
	// SaveToken creates refresh token for the user.
	SaveToken(ctx context.Context, refreshToken, userId, clientId, tokenId string) (string, error)
	// RemoveToken removes refresh token.
	// It returns [ErrTokenNotFound] if no token are found.
	RemoveToken(ctx context.Context, id string) error
	// FindToken finds and returns a Token entity by its refresh token string.
	// It returns [ErrTokenNotFound] if no token are found.
	FindToken(ctx context.Context, id string) (*domain.Token, error)
	// GetPublicKey returns public rsa keys
	GetPublicKey() rsa.PublicKey
}

type tokenService struct {
	tokenStore repository.TokenRepository
	keys       config.KeysPair
}

func NewTokenService(tokenStore repository.TokenRepository, keys *config.KeysPair) TokenService {
	return &tokenService{
		tokenStore: tokenStore,
		keys:       *keys,
	}
}

func (s *tokenService) GenerateTokens(user domain.UserDto, clientId string) (*domain.TokensPair, string, error) {
	op := "tokenService.GenerateTokens"
	accessExpiry, _ := time.ParseDuration("30m")
	refreshExpiry, _ := time.ParseDuration("336h")
	now := time.Now()
	id := xid.New().String()

	// Access token
	accessClaims := domain.AccessClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
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
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Refresh token
	refreshClaims := jwt.RegisteredClaims{
		ID:        id,
		Issuer:    "macauth",
		Subject:   user.UserId,
		Audience:  jwt.ClaimStrings{clientId},
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

func (s *tokenService) ValidateAccessToken(accessToken string) (*domain.UserDto, string, error) {
	op := "tokenService.ValidateAccessToken"
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

func (s *tokenService) SaveToken(ctx context.Context, refreshToken, userId, clientId, tokenId string) (string, error) {
	op := "tokenService.SaveToken"

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
		ClientId:     clientId,
		ExpiresAt:    expiresAt,
	}
	id, err := s.tokenStore.Create(ctx, &token)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *tokenService) RemoveToken(ctx context.Context, id string) error {
	op := "tokenService.RemoveToken"
	if err := s.tokenStore.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *tokenService) FindToken(ctx context.Context, id string) (*domain.Token, error) {
	op := "tokenService.FindToken"
	token, err := s.tokenStore.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (s *tokenService) GetPublicKey() rsa.PublicKey {
	return *s.keys.PublicKey
}
