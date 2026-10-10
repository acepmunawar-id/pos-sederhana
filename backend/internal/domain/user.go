package domain

import (
	"context"
	"time"

	"github.com/posumkm/backend/pkg/pagination"
)

type Role struct {
	ID        int64
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID           int64
	RoleID       int64
	RoleCode     string
	RoleName     string
	Name         string
	Email        string
	PasswordHash string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type UserFilter struct {
	pagination.Params
	RoleCode string
	IsActive *bool
}

type RoleRepository interface {
	List(ctx context.Context) ([]Role, error)
	FindByID(ctx context.Context, id int64) (*Role, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id int64, at time.Time) error
	SoftDelete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	EmailExists(ctx context.Context, email string, excludeID int64) (bool, error)
	CountActiveByRole(ctx context.Context, roleCode string) (int64, error)
	List(ctx context.Context, f UserFilter) ([]User, int64, error)
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	UserAgent string
	IPAddress string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id int64, at time.Time) error
	RevokeAllByUser(ctx context.Context, userID int64, at time.Time) error
}
