package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
)

type RoleRepository struct{ pool *pgxpool.Pool }

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository { return &RoleRepository{pool: pool} }

func (r *RoleRepository) List(ctx context.Context) ([]domain.Role, error) {
	rows, err := conn(ctx, r.pool).Query(ctx,
		`SELECT id, code, name, created_at, updated_at FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Role, error) {
		var x domain.Role
		err := row.Scan(&x.ID, &x.Code, &x.Name, &x.CreatedAt, &x.UpdatedAt)
		return x, err
	})
}

func (r *RoleRepository) FindByID(ctx context.Context, id int64) (*domain.Role, error) {
	var x domain.Role
	err := conn(ctx, r.pool).QueryRow(ctx,
		`SELECT id, code, name, created_at, updated_at FROM roles WHERE id = $1`, id).
		Scan(&x.ID, &x.Code, &x.Name, &x.CreatedAt, &x.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &x, nil
}

type UserRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) *UserRepository { return &UserRepository{pool: pool} }

const userSelect = `SELECT u.id, u.role_id, r.code, r.name, u.name, u.email, u.password_hash, u.is_active,
	u.last_login_at, u.created_at, u.updated_at, u.deleted_at
	FROM users u JOIN roles r ON r.id = u.role_id`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.RoleID, &u.RoleCode, &u.RoleName, &u.Name, &u.Email, &u.PasswordHash,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	err := conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO users (role_id, name, email, password_hash, is_active)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		u.RoleID, u.Name, u.Email, u.PasswordHash, u.IsActive).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	return mapErr(err)
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE users SET role_id = $2, name = $3, email = $4, is_active = $5
		 WHERE id = $1 AND deleted_at IS NULL`,
		u.ID, u.RoleID, u.Name, u.Email, u.IsActive))
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1 AND deleted_at IS NULL`, id, passwordHash))
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64, at time.Time) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE users SET last_login_at = $2 WHERE id = $1`, id, at))
}

func (r *UserRepository) SoftDelete(ctx context.Context, id int64) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE users SET deleted_at = NOW(), is_active = FALSE WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return scanUser(conn(ctx, r.pool).QueryRow(ctx, userSelect+` WHERE u.id = $1 AND u.deleted_at IS NULL`, id))
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(conn(ctx, r.pool).QueryRow(ctx,
		userSelect+` WHERE LOWER(u.email) = LOWER($1) AND u.deleted_at IS NULL`, email))
}

func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	return exists(ctx, conn(ctx, r.pool),
		`SELECT 1 FROM users WHERE LOWER(email) = LOWER($1) AND id <> $2 AND deleted_at IS NULL`, email, excludeID)
}

func (r *UserRepository) CountActiveByRole(ctx context.Context, roleCode string) (int64, error) {
	return count(ctx, conn(ctx, r.pool),
		`SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id
		 WHERE r.code = $1 AND u.is_active AND u.deleted_at IS NULL`, []any{roleCode})
}

var userSorts = map[string]string{
	"name": "u.name", "email": "u.email", "created_at": "u.created_at", "last_login_at": "u.last_login_at",
}

func (r *UserRepository) List(ctx context.Context, f domain.UserFilter) ([]domain.User, int64, error) {
	w := &where{}
	w.add("u.deleted_at IS NULL")
	w.search(f.Search, "u.name", "u.email")
	if f.RoleCode != "" {
		w.add("r.code = ?", f.RoleCode)
	}
	if f.IsActive != nil {
		w.add("u.is_active = ?", *f.IsActive)
	}
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id`+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(f.Params)
	rows, err := q.Query(ctx, userSelect+w.sql()+orderBy(f.Params, userSorts, "u.created_at DESC", "u.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.User, error) {
		u, err := scanUser(row)
		if err != nil {
			return domain.User{}, err
		}
		return *u, nil
	})
	return users, total, err
}

type RefreshTokenRepository struct{ pool *pgxpool.Pool }

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
		 VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5) RETURNING id, created_at`,
		t.UserID, t.TokenHash, t.UserAgent, t.IPAddress, t.ExpiresAt).Scan(&t.ID, &t.CreatedAt))
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	var ua, ip *string
	err := conn(ctx, r.pool).QueryRow(ctx,
		`SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, revoked_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`, hash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &ua, &ip, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	if ua != nil {
		t.UserAgent = *ua
	}
	if ip != nil {
		t.IPAddress = *ip
	}
	return &t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int64, at time.Time) error {
	_, err := conn(ctx, r.pool).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL`, id, at)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUser(ctx context.Context, userID int64, at time.Time) error {
	_, err := conn(ctx, r.pool).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`, userID, at)
	return err
}
