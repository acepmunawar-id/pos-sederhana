package dto

import (
	"time"

	"github.com/posumkm/backend/internal/domain"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	RoleID   int64  `json:"role_id" validate:"required,gt=0"`
	IsActive *bool  `json:"is_active"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=150"`
	RoleID   int64  `json:"role_id" validate:"required,gt=0"`
	IsActive bool   `json:"is_active"`
	// Password is optional; when set, the user's password is reset.
	Password string `json:"password" validate:"omitempty,min=8,max=72"`
}

type RoleResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code" example:"cashier"`
	Name string `json:"name" example:"Kasir"`
}

type UserResponse struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Role        RoleResponse `json:"role"`
	IsActive    bool         `json:"is_active"`
	LastLoginAt *time.Time   `json:"last_login_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func ToRoleResponse(r domain.Role) RoleResponse {
	return RoleResponse{ID: r.ID, Code: r.Code, Name: r.Name}
}

func ToUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Role:        RoleResponse{ID: u.RoleID, Code: u.RoleCode, Name: u.RoleName},
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func ToUserResponses(users []domain.User) []UserResponse {
	out := make([]UserResponse, len(users))
	for i := range users {
		out[i] = ToUserResponse(&users[i])
	}
	return out
}
