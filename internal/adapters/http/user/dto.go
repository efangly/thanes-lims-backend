package user

import (
	"time"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest carries the refresh token for non-browser clients that
// can't rely on the Refresh Cookie - the cookie is always tried first (see
// ADR 0004), so this field is optional at the JSON level even though the
// struct tag says required; handlers only bind it when no cookie was found.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AccessTokenResponse is returned by Login: the Refresh Token is never
// echoed in the body, only set as an httpOnly cookie (see ADR 0004).
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// TokenPairResponse is returned by Refresh. RefreshToken is populated only
// when the request didn't use the Refresh Cookie (i.e. a non-browser client
// on the JSON-body fallback path) - cookie-based callers get the new token
// via Set-Cookie only.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required"`
	Role string `json:"role" validate:"required"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toUserResponse(u domainuser.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
