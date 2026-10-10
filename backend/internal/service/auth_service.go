package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/hash"
)

// TokenIssuer issues signed access tokens (implemented by pkg/jwt.Manager).
type TokenIssuer interface {
	Generate(userID int64, role, name string, now time.Time) (string, time.Time, error)
}

type AuthService struct {
	users      domain.UserRepository
	tokens     domain.RefreshTokenRepository
	tx         domain.TxManager
	hasher     hash.Hasher
	issuer     TokenIssuer
	refreshTTL time.Duration
	now        Clock
}

func NewAuthService(users domain.UserRepository, tokens domain.RefreshTokenRepository, tx domain.TxManager,
	hasher hash.Hasher, issuer TokenIssuer, refreshTTL time.Duration, now Clock) *AuthService {
	return &AuthService{users: users, tokens: tokens, tx: tx, hasher: hasher, issuer: issuer,
		refreshTTL: refreshTTL, now: now}
}

// ClientInfo is recorded alongside refresh tokens for auditing.
type ClientInfo struct {
	UserAgent string
	IP        string
}

var errInvalidCredentials = apperror.Unauthorized("Email atau password salah")

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, client ClientInfo) (*dto.AuthResult, error) {
	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if errors.Is(err, domain.ErrNotFound) {
		return nil, errInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !s.hasher.Compare(user.PasswordHash, req.Password) {
		return nil, errInvalidCredentials
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("Akun Anda dinonaktifkan. Hubungi pemilik toko")
	}

	now := s.now()
	var result *dto.AuthResult
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.users.UpdateLastLogin(ctx, user.ID, now); err != nil {
			return err
		}
		user.LastLoginAt = &now
		result, err = s.issue(ctx, user, client, now)
		return err
	})
	return result, err
}

// Refresh rotates the refresh token. Presenting an already-revoked token is treated as
// token theft: every session of that user is revoked.
func (s *AuthService) Refresh(ctx context.Context, raw string, client ClientInfo) (*dto.AuthResult, error) {
	invalid := apperror.Unauthorized("Sesi berakhir, silakan login kembali")
	stored, err := s.tokens.FindByHash(ctx, hash.SHA256(raw))
	if errors.Is(err, domain.ErrNotFound) {
		return nil, invalid
	}
	if err != nil {
		return nil, err
	}
	now := s.now()
	if stored.RevokedAt != nil {
		if err := s.tokens.RevokeAllByUser(ctx, stored.UserID, now); err != nil {
			return nil, err
		}
		return nil, invalid
	}
	if !now.Before(stored.ExpiresAt) {
		return nil, invalid
	}
	user, err := s.users.FindByID(ctx, stored.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, invalid
	}
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("Akun Anda dinonaktifkan. Hubungi pemilik toko")
	}

	var result *dto.AuthResult
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.tokens.Revoke(ctx, stored.ID, now); err != nil {
			return err
		}
		result, err = s.issue(ctx, user, client, now)
		return err
	})
	return result, err
}

// Logout revokes the given refresh token if it belongs to the user. Unknown tokens are ignored.
func (s *AuthService) Logout(ctx context.Context, userID int64, raw string) error {
	stored, err := s.tokens.FindByHash(ctx, hash.SHA256(raw))
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if stored.UserID != userID {
		return nil
	}
	return s.tokens.Revoke(ctx, stored.ID, s.now())
}

func (s *AuthService) Me(ctx context.Context, userID int64) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	return u, notFound(err, "Pengguna tidak ditemukan")
}

// ChangePassword verifies the old password, stores the new one and signs out every session.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return notFound(err, "Pengguna tidak ditemukan")
	}
	if !s.hasher.Compare(user.PasswordHash, req.OldPassword) {
		return apperror.Validation([]apperror.FieldError{{Field: "old_password", Message: "password lama salah"}})
	}
	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.users.UpdatePassword(ctx, userID, newHash); err != nil {
			return err
		}
		return s.tokens.RevokeAllByUser(ctx, userID, s.now())
	})
}

func (s *AuthService) issue(ctx context.Context, user *domain.User, client ClientInfo, now time.Time) (*dto.AuthResult, error) {
	access, accessExp, err := s.issuer.Generate(user.ID, user.RoleCode, user.Name, now)
	if err != nil {
		return nil, err
	}
	raw, err := hash.RandomToken(32)
	if err != nil {
		return nil, err
	}
	rt := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash.SHA256(raw),
		UserAgent: truncate(client.UserAgent, 255),
		IPAddress: truncate(client.IP, 64),
		ExpiresAt: now.Add(s.refreshTTL),
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return nil, err
	}
	return &dto.AuthResult{
		AccessToken: access, AccessTokenExpiresAt: accessExp,
		RefreshToken: raw, RefreshTokenExpiresAt: rt.ExpiresAt,
		User: user,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
