// Package dto defines request and response payloads, separate from domain entities.
package dto

import (
	"time"

	"github.com/posumkm/backend/internal/domain"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"owner@toko.id"`
	Password string `json:"password" validate:"required" example:"owner123"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72,nefield=OldPassword"`
}

type TokenResponse struct {
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	TokenType             string       `json:"token_type" example:"Bearer"`
	User                  UserResponse `json:"user"`
}

// AuthResult is what AuthService returns; the handler maps it to TokenResponse.
type AuthResult struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	User                  *domain.User
}

func ToTokenResponse(r *AuthResult) TokenResponse {
	return TokenResponse{
		AccessToken:           r.AccessToken,
		AccessTokenExpiresAt:  r.AccessTokenExpiresAt,
		RefreshToken:          r.RefreshToken,
		RefreshTokenExpiresAt: r.RefreshTokenExpiresAt,
		TokenType:             "Bearer",
		User:                  ToUserResponse(r.User),
	}
}
