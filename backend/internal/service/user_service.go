package service

import (
	"context"
	"errors"
	"strings"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/hash"
)

type UserService struct {
	users  domain.UserRepository
	roles  domain.RoleRepository
	tokens domain.RefreshTokenRepository
	tx     domain.TxManager
	hasher hash.Hasher
	now    Clock
}

func NewUserService(users domain.UserRepository, roles domain.RoleRepository, tokens domain.RefreshTokenRepository,
	tx domain.TxManager, hasher hash.Hasher, now Clock) *UserService {
	return &UserService{users: users, roles: roles, tokens: tokens, tx: tx, hasher: hasher, now: now}
}

func (s *UserService) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.roles.List(ctx)
}

func (s *UserService) List(ctx context.Context, f domain.UserFilter) ([]domain.User, int64, error) {
	f.Normalize()
	return s.users.List(ctx, f)
}

func (s *UserService) Get(ctx context.Context, id int64) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, id)
	return u, notFound(err, "Pengguna tidak ditemukan")
}

func (s *UserService) Create(ctx context.Context, actor Actor, req dto.CreateUserRequest) (*domain.User, error) {
	role, err := s.resolveRole(ctx, actor, req.RoleID)
	if err != nil {
		return nil, err
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := s.ensureEmailFree(ctx, email, 0); err != nil {
		return nil, err
	}
	pw, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		RoleID: role.ID, Name: strings.TrimSpace(req.Name), Email: email, PasswordHash: pw,
		IsActive: req.IsActive == nil || *req.IsActive,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, duplicate(err, "Email sudah digunakan")
	}
	return s.Get(ctx, u.ID)
}

func (s *UserService) Update(ctx context.Context, actor Actor, id int64, req dto.UpdateUserRequest) (*domain.User, error) {
	existing, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCanManage(actor, existing); err != nil {
		return nil, err
	}
	role, err := s.resolveRole(ctx, actor, req.RoleID)
	if err != nil {
		return nil, err
	}
	if actor.UserID == id && (role.Code != existing.RoleCode || !req.IsActive) {
		return nil, apperror.BadRequest("Anda tidak dapat mengubah role atau menonaktifkan akun sendiri")
	}
	if existing.RoleCode == domain.RoleOwner && (role.Code != domain.RoleOwner || !req.IsActive) {
		if err := s.ensureNotLastOwner(ctx); err != nil {
			return nil, err
		}
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := s.ensureEmailFree(ctx, email, id); err != nil {
		return nil, err
	}

	existing.RoleID, existing.Name, existing.Email, existing.IsActive = role.ID, strings.TrimSpace(req.Name), email, req.IsActive
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.users.Update(ctx, existing); err != nil {
			return duplicate(err, "Email sudah digunakan")
		}
		if req.Password != "" {
			pw, err := s.hasher.Hash(req.Password)
			if err != nil {
				return err
			}
			if err := s.users.UpdatePassword(ctx, id, pw); err != nil {
				return err
			}
		}
		// Changing credentials or deactivating a user ends their existing sessions.
		if req.Password != "" || !req.IsActive {
			return s.tokens.RevokeAllByUser(ctx, id, s.now())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *UserService) Delete(ctx context.Context, actor Actor, id int64) error {
	if actor.UserID == id {
		return apperror.BadRequest("Anda tidak dapat menghapus akun sendiri")
	}
	existing, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ensureCanManage(actor, existing); err != nil {
		return err
	}
	if existing.RoleCode == domain.RoleOwner {
		if err := s.ensureNotLastOwner(ctx); err != nil {
			return err
		}
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.users.SoftDelete(ctx, id); err != nil {
			return notFound(err, "Pengguna tidak ditemukan")
		}
		return s.tokens.RevokeAllByUser(ctx, id, s.now())
	})
}

// resolveRole loads the role and enforces that only an Owner may grant the Owner role.
func (s *UserService) resolveRole(ctx context.Context, actor Actor, roleID int64) (*domain.Role, error) {
	role, err := s.roles.FindByID(ctx, roleID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, apperror.Validation([]apperror.FieldError{{Field: "role_id", Message: "role tidak ditemukan"}})
	}
	if err != nil {
		return nil, err
	}
	if role.Code == domain.RoleOwner && actor.Role != domain.RoleOwner {
		return nil, apperror.Forbidden("Hanya Owner yang dapat memberikan role Owner")
	}
	return role, nil
}

// ensureCanManage stops Admins from editing or deleting Owner accounts.
func (s *UserService) ensureCanManage(actor Actor, target *domain.User) error {
	if target.RoleCode == domain.RoleOwner && actor.Role != domain.RoleOwner {
		return apperror.Forbidden("Admin tidak dapat mengubah akun Owner")
	}
	return nil
}

func (s *UserService) ensureNotLastOwner(ctx context.Context) error {
	n, err := s.users.CountActiveByRole(ctx, domain.RoleOwner)
	if err != nil {
		return err
	}
	if n <= 1 {
		return apperror.BadRequest("Minimal harus ada satu Owner yang aktif")
	}
	return nil
}

func (s *UserService) ensureEmailFree(ctx context.Context, email string, excludeID int64) error {
	taken, err := s.users.EmailExists(ctx, email, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return apperror.Validation([]apperror.FieldError{{Field: "email", Message: "email sudah digunakan"}})
	}
	return nil
}
