<p align="center">
  <img src="assets/logo.png" alt="macauth logo" width="300">
</p>

# macauth

### A lightweight SSO microservice designed for authentication and secure session management for your pet-project

## 🚀 Quick Start (Docker)

The easiest way to run `macauth` is using Docker. An automated initialization script is provided to set everything up in seconds.

### 1. Clone the repository

```bash
git clone https://github.com/dmi3midd/macauth.git
cd macauth
```

### 2. Initialize the environment

Run the setup script. It will automatically:

- Prepare `config.yaml`.
- Generate an RSA key pair (`private.pem` & `public.pem`) required for JWT signing.

```bash
make setup
```

### 3. Start the service

**Option A — All-in-one** (app + bundled PostgreSQL):

No extra configuration needed. This starts a PostgreSQL container alongside the app:

```bash
# Set host: postgres in config.yaml
make docker-run-all
```

**Option B — App only** (connect to an external PostgreSQL):

If you already have a running PostgreSQL instance, configure the connection in `config.yaml`:

```yaml
database:
  dbname: macauth
  host: your-postgres-host  # e.g., localhost, 192.168.1.100, host.docker.internal
  port: 5432
  user: macauth_user
  password: macauth_9876
  sslmode: disable
```

Then start only the app container:

```bash
make docker-run
```

> **Tip:** When connecting to a PostgreSQL running on the host machine from Docker, use `host.docker.internal` as the host.

Your SSO service is now up and running at `http://localhost:2800` (default port).

---

## 🔑 Integration with other microservices

You can make requests to `http://localhost:2800/macauth/api/v1/auth/validate` to validate an access token. Send a `POST` request with the `accessToken` in the JSON request body:

```json
{
  "accessToken": "your-access-token"
}
```

If you want to validate access tokens locally (without network requests), you can use `macauth`'s public key. To get the public key, you can make a `GET` request to `http://localhost:2800/macauth/api/v1/public-key`.

To integrate token validation into your projects:

1. Make a `GET` request to `http://localhost:2800/macauth/api/v1/public-key` during the initialization of your external service.
2. Cache the returned **Public Key** in memory.
3. Validate all incoming JWT access tokens locally using this public key—ensuring zero network latency and maximum performance.

### Example in Go

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
 User UserDto `json:"user"`
 jwt.RegisteredClaims
}
type UserDto struct {
 UserId      string   `json:"userId"`
 Username    string   `json:"username"`
 Email       string   `json:"email"`
}

func ValidateAccessToken(accessToken string, publicKey *rsa.PublicKey) (*UserDto, string, error) {
 op := "TokenService.ValidateAccessToken"
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
  UserId:      userId,
  Username:    claims.User.Username,
  Email:       claims.User.Email,
 }, tokenId, nil
}
```

---

## 🔄 Password Reset Flow (Service-to-Service)

Since `macauth` does not communicate with end-users directly, the password reset is a two-step flow managed by your consuming service:

### 1. Initiate Password Reset

When a user requests a password reset:

1. Make a `POST` request to `/macauth/api/v1/reset/initiate-reset` with the user's email:

   ```json
   {
     "email": "user@example.com"
   }
   ```

2. `macauth` generates a secure reset token and returns it:

   ```json
   {
     "resetToken": "generated-raw-token-here",
     "email": "user@example.com"
   }
   ```

3. Your consuming service constructs the reset link (e.g., `https://myapp.com/reset-password?token=generated-raw-token-here`) and sends it to the user via Email/SMS.

### 2. Confirm Password Reset

When the user submits their new password:

1. Make a `POST` request to `/macauth/api/v1/reset/confirm-reset` with the token and new password:

   ```json
   {
     "resetToken": "generated-raw-token-here",
     "newPassword": "NewSecurePassword123!"
   }
   ```

2. `macauth` validates the token, hashes the new password, updates the user record, marks the token as used, and **invalidates all active sessions (Refresh Tokens)** for this user.

## ❗️ WARNINGS

Since 4.0.0 version macauth philosophy changed. Now macauth doesn't support permissions.
Macauth now its just a simple auth service for your another backend service.
For full SSO service you can see [grpcsso](https://github.com/dmi3midd/grpcsso), but now it's still in development.
