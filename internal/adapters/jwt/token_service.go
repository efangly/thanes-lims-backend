package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Adapter struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func New(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Adapter {
	return &Adapter{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

type claims struct {
	UserID int64     `json:"user_id"`
	Name   string    `json:"name"`
	Role   user.Role `json:"role"`
	// Permissions is only set on access tokens - see
	// ports/user.TokenService.GenerateAccessToken and ADR 0002.
	Permissions []string `json:"permissions,omitempty"`
	jwtlib.RegisteredClaims
}

func (a *Adapter) GenerateAccessToken(u user.User, permissions []string) (string, error) {
	c := claims{
		UserID:      u.ID,
		Name:        u.Name,
		Role:        u.Role,
		Permissions: permissions,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(a.accessTTL)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, c)
	return token.SignedString(a.accessSecret)
}

func (a *Adapter) GenerateRefreshToken(u user.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(a.refreshTTL)
	c := claims{
		UserID: u.ID,
		Name:   u.Name,
		Role:   u.Role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			// ID (jti): without this, two logins for the same user within
			// the same wall-clock second produce byte-identical tokens
			// (IssuedAt has only second resolution), which collide on the
			// refresh_tokens.token_hash unique constraint.
			ID: uuid.NewString(),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, c)
	signed, err := token.SignedString(a.refreshSecret)
	return signed, expiresAt, err
}

func (a *Adapter) ParseAccessToken(tokenStr string) (portuser.Claims, error) {
	return a.parse(tokenStr, a.accessSecret)
}

func (a *Adapter) ParseRefreshToken(tokenStr string) (portuser.Claims, error) {
	return a.parse(tokenStr, a.refreshSecret)
}

func (a *Adapter) parse(tokenStr string, secret []byte) (portuser.Claims, error) {
	var c claims
	token, err := jwtlib.ParseWithClaims(tokenStr, &c, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return portuser.Claims{}, fmt.Errorf("jwt: invalid token: %w", err)
	}
	return portuser.Claims{UserID: c.UserID, Name: c.Name, Role: c.Role, Permissions: c.Permissions}, nil
}

// HashRefreshToken is a plain SHA-256 digest - refresh tokens are already
// high-entropy random-looking JWTs, so a fast deterministic hash (rather
// than bcrypt) is enough to prevent a raw DB dump from being replayable,
// while keeping lookups a simple indexed equality match.
func (a *Adapter) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
