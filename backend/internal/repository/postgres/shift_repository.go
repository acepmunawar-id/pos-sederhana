package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
)

type ShiftRepository struct{ pool *pgxpool.Pool }

func NewShiftRepository(pool *pgxpool.Pool) *ShiftRepository { return &ShiftRepository{pool: pool} }

const shiftSelect = `SELECT s.id, s.user_id, u.name, s.status, s.opening_cash, s.expected_cash, s.actual_cash,
	s.cash_difference, s.total_sales, s.total_transactions, s.opening_note, s.closing_note, s.opened_at, s.closed_at,
	s.created_at, s.updated_at
	FROM shifts s JOIN users u ON u.id = s.user_id`

func scanShift(row pgx.Row) (domain.Shift, error) {
	var s domain.Shift
	err := row.Scan(&s.ID, &s.UserID, &s.UserName, &s.Status, &s.OpeningCash, &s.ExpectedCash, &s.ActualCash,
		&s.CashDifference, &s.TotalSales, &s.TotalTransactions, &s.OpeningNote, &s.ClosingNote, &s.OpenedAt,
		&s.ClosedAt, &s.CreatedAt, &s.UpdatedAt)
	return s, mapErr(err)
}

func (r *ShiftRepository) Create(ctx context.Context, s *domain.Shift) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO shifts (user_id, status, opening_cash, opening_note, opened_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		s.UserID, s.Status, s.OpeningCash, s.OpeningNote, s.OpenedAt).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt))
}

func (r *ShiftRepository) Close(ctx context.Context, s *domain.Shift) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE shifts SET status = $2, expected_cash = $3, actual_cash = $4, cash_difference = $5,
		 total_sales = $6, total_transactions = $7, closing_note = $8, closed_at = $9
		 WHERE id = $1 AND status = 'OPEN'`,
		s.ID, s.Status, s.ExpectedCash, s.ActualCash, s.CashDifference, s.TotalSales, s.TotalTransactions,
		s.ClosingNote, s.ClosedAt))
}

func (r *ShiftRepository) FindByID(ctx context.Context, id int64) (*domain.Shift, error) {
	s, err := scanShift(conn(ctx, r.pool).QueryRow(ctx, shiftSelect+` WHERE s.id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShiftRepository) FindOpenByUser(ctx context.Context, userID int64) (*domain.Shift, error) {
	s, err := scanShift(conn(ctx, r.pool).QueryRow(ctx,
		shiftSelect+` WHERE s.user_id = $1 AND s.status = 'OPEN'`, userID))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShiftRepository) Summary(ctx context.Context, shiftID int64) (*domain.ShiftSummary, error) {
	q := conn(ctx, r.pool)
	sum := &domain.ShiftSummary{PaymentTotals: map[string]int64{}}
	err := q.QueryRow(ctx,
		`SELECT COALESCE(SUM(total) FILTER (WHERE status = 'COMPLETED'), 0),
		        COUNT(*) FILTER (WHERE status = 'COMPLETED'),
		        COUNT(*) FILTER (WHERE status = 'VOIDED')
		 FROM transactions WHERE shift_id = $1`, shiftID).
		Scan(&sum.TotalSales, &sum.TotalTransactions, &sum.VoidedCount)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx,
		`SELECT p.method, COALESCE(SUM(p.amount), 0)
		 FROM payments p JOIN transactions t ON t.id = p.transaction_id
		 WHERE t.shift_id = $1 AND t.status = 'COMPLETED'
		 GROUP BY p.method`, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var method string
		var amount int64
		if err := rows.Scan(&method, &amount); err != nil {
			return nil, err
		}
		sum.PaymentTotals[method] = amount
	}
	sum.CashSales = sum.PaymentTotals[domain.PaymentCash]
	return sum, rows.Err()
}

var shiftSorts = map[string]string{"opened_at": "s.opened_at", "closed_at": "s.closed_at", "cashier": "u.name"}

func (r *ShiftRepository) List(ctx context.Context, f domain.ShiftFilter) ([]domain.Shift, int64, error) {
	w := &where{}
	w.search(f.Search, "u.name")
	if f.UserID != nil {
		w.add("s.user_id = ?", *f.UserID)
	}
	if f.Status != "" {
		w.add("s.status = ?", f.Status)
	}
	if f.DateFrom != nil {
		w.add("s.opened_at >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		w.add("s.opened_at < ?", *f.DateTo)
	}
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*) FROM shifts s JOIN users u ON u.id = s.user_id`+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(f.Params)
	rows, err := q.Query(ctx, shiftSelect+w.sql()+orderBy(f.Params, shiftSorts, "s.opened_at DESC", "s.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Shift, error) { return scanShift(row) })
	return items, total, err
}
