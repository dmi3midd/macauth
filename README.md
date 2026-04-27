<p align="center">
  <img src="assets/logo.png" alt="macauth logo" width="300">
</p>

# macauth

### A lightweight SSO microservice designed for centralized authentication and secure session management for your pet-projects.

## 🚀 Quick Start (Docker)

The easiest way to run `macauth` is using Docker. An automated initialization script is provided to set everything up in seconds.

### 1. Clone the repository
```bash
git clone https://github.com/dmi3midd/macauth.git
cd macauth
```

### 2. Initialize the environment
Run the setup script. It will automatically:
- Generate a highly secure random `API_KEY` (saved in the `.env` file).
- Prepare `config.yaml`.
- Create a SQLite database file.
- Generate an RSA key pair (`private.pem` & `public.pem`) required for JWT signing.

```bash
./setup.sh
```
*Note: Make sure to securely store your `API_KEY` found in the `.env` file.*

### 3. Start the service
Start the API in the background using Docker Compose:
```bash
docker compose up -d --build
```
Your SSO service is now up and running at `http://localhost:2800` (default port).

---

## 🔑 Integration with other microservices
You can make requests to `http://localhost:2800/macauth/api/v1/user/validate` to validate access token. To validate token you need to set `Authorization` header with value `Bearer {access_token}`. Also you need to set `x-client-id` header with value `{client_id}` and `x-client-secret` header with value `{client_secret}`.

But if you want to validate access token locally (without network requests), you can use `macauth`'s public key. To get the public key, you can make a `GET` request to `http://localhost:2800/macauth/api/v1/client/public-key`.
*(Note: since this endpoint is protected, you must include your setup `API_KEY` in the request headers as `x-api-key`)*.

To integrate token validation into your projects:
1. Make a `GET` request to `http://localhost:2800/macauth/api/v1/client/public-key` during the initialization of your external service.
2. Cache the returned **Public Key** in memory.
3. Validate all incoming JWT access tokens locally using this public key—ensuring zero network latency and maximum performance.

### Example in Go:
```go
import (
    "errors"
    "fmt"
    "crypto/rsa"
    "github.com/golang-jwt/jwt/v5"
)

var (
    ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
    ErrInvalidAccessToken      = errors.New("invalid access token")
    ErrSubjectAndIDNotFound    = errors.New("subject and id not found")
)

type AccessClaims struct {
	Username string
	Email    string
	jwt.RegisteredClaims
}

type UserDto struct {
	UserId   string
	Username string
	Email    string
}

func ValidateAccessToken(accessToken string, publicKey *rsa.PublicKey) (*UserDto, string, error) {
	op := "tokenService.ValidateAccessToken"
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%s: %w %v", op, ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return publicKey, nil
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

	return &UserDto{
		UserId:   userId,
		Username: claims.Username,
		Email:    claims.Email,
	}, tokenId, nil
}
```