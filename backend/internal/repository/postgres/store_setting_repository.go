package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
)

type StoreSettingRepository struct{ pool *pgxpool.Pool }

func NewStoreSettingRepository(pool *pgxpool.Pool) *StoreSettingRepository {
	return &StoreSettingRepository{pool: pool}
}

func (r *StoreSettingRepository) Get(ctx context.Context) (*domain.StoreSetting, error) {
	var s domain.StoreSetting
	err := conn(ctx, r.pool).QueryRow(ctx,
		`SELECT id, name, address, phone, logo_url, tax_rate_bp, receipt_footer, updated_at
		 FROM store_settings WHERE id = 1`).
		Scan(&s.ID, &s.Name, &s.Address, &s.Phone, &s.LogoURL, &s.TaxRateBP, &s.ReceiptFooter, &s.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &s, nil
}

func (r *StoreSettingRepository) Update(ctx context.Context, s *domain.StoreSetting) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`UPDATE store_settings SET name = $1, address = $2, phone = $3, tax_rate_bp = $4, receipt_footer = $5
		 WHERE id = 1 RETURNING updated_at`,
		s.Name, s.Address, s.Phone, s.TaxRateBP, s.ReceiptFooter).Scan(&s.UpdatedAt))
}

func (r *StoreSettingRepository) UpdateLogo(ctx context.Context, url *string) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx, `UPDATE store_settings SET logo_url = $1 WHERE id = 1`, url))
}
