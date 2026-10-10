package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
)

type StockMovementRepository struct{ pool *pgxpool.Pool }

func NewStockMovementRepository(pool *pgxpool.Pool) *StockMovementRepository {
	return &StockMovementRepository{pool: pool}
}

func (r *StockMovementRepository) Create(ctx context.Context, m *domain.StockMovement) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO stock_movements (product_id, user_id, type, quantity, stock_before, stock_after,
		 reference_type, reference_id, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`,
		m.ProductID, m.UserID, m.Type, m.Quantity, m.StockBefore, m.StockAfter, m.ReferenceType, m.ReferenceID, m.Note).
		Scan(&m.ID, &m.CreatedAt))
}

var movementSorts = map[string]string{
	"created_at": "m.created_at", "quantity": "m.quantity", "product": "p.name", "type": "m.type",
}

func (r *StockMovementRepository) List(ctx context.Context, f domain.StockMovementFilter) ([]domain.StockMovement, int64, error) {
	w := &where{}
	w.search(f.Search, "p.name", "p.sku", "m.note")
	if f.ProductID != nil {
		w.add("m.product_id = ?", *f.ProductID)
	}
	if f.Type != "" {
		w.add("m.type = ?", f.Type)
	}
	if f.DateFrom != nil {
		w.add("m.created_at >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		w.add("m.created_at < ?", *f.DateTo)
	}
	const from = ` FROM stock_movements m JOIN products p ON p.id = m.product_id JOIN users u ON u.id = m.user_id`
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*)`+from+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(f.Params)
	rows, err := q.Query(ctx,
		`SELECT m.id, m.product_id, p.name, p.sku, m.user_id, u.name, m.type, m.quantity, m.stock_before,
		 m.stock_after, m.reference_type, m.reference_id, m.note, m.created_at`+from+w.sql()+
			orderBy(f.Params, movementSorts, "m.created_at DESC", "m.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.StockMovement, error) {
		var m domain.StockMovement
		err := row.Scan(&m.ID, &m.ProductID, &m.ProductName, &m.ProductSKU, &m.UserID, &m.UserName, &m.Type,
			&m.Quantity, &m.StockBefore, &m.StockAfter, &m.ReferenceType, &m.ReferenceID, &m.Note, &m.CreatedAt)
		return m, err
	})
	return items, total, err
}
