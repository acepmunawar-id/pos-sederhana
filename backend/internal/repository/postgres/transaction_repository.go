package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
)

type TransactionRepository struct{ pool *pgxpool.Pool }

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

const txSelect = `SELECT t.id, t.invoice_number, t.idempotency_key, t.shift_id, t.user_id, u.name, t.subtotal,
	t.discount_type, t.discount_value, t.discount_amount, t.tax_rate_bp, t.tax_amount, t.total, t.paid_amount,
	t.change_amount, t.total_cost, t.status, t.note, t.voided_at, t.voided_by, vu.name, t.void_reason,
	t.created_at, t.updated_at
	FROM transactions t
	JOIN users u ON u.id = t.user_id
	LEFT JOIN users vu ON vu.id = t.voided_by`

func scanTx(row pgx.Row) (domain.Transaction, error) {
	var t domain.Transaction
	err := row.Scan(&t.ID, &t.InvoiceNumber, &t.IdempotencyKey, &t.ShiftID, &t.UserID, &t.CashierName, &t.Subtotal,
		&t.DiscountType, &t.DiscountValue, &t.DiscountAmount, &t.TaxRateBP, &t.TaxAmount, &t.Total, &t.PaidAmount,
		&t.ChangeAmount, &t.TotalCost, &t.Status, &t.Note, &t.VoidedAt, &t.VoidedBy, &t.VoidedByName, &t.VoidReason,
		&t.CreatedAt, &t.UpdatedAt)
	return t, mapErr(err)
}

func (r *TransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	q := conn(ctx, r.pool)
	err := q.QueryRow(ctx,
		`INSERT INTO transactions (invoice_number, idempotency_key, shift_id, user_id, subtotal, discount_type,
		 discount_value, discount_amount, tax_rate_bp, tax_amount, total, paid_amount, change_amount, total_cost,
		 status, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		 RETURNING id, created_at, updated_at`,
		t.InvoiceNumber, t.IdempotencyKey, t.ShiftID, t.UserID, t.Subtotal, t.DiscountType, t.DiscountValue,
		t.DiscountAmount, t.TaxRateBP, t.TaxAmount, t.Total, t.PaidAmount, t.ChangeAmount, t.TotalCost,
		t.Status, t.Note).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return mapErr(err)
	}

	batch := &pgx.Batch{}
	for i := range t.Items {
		it := &t.Items[i]
		it.TransactionID = t.ID
		batch.Queue(`INSERT INTO transaction_items (transaction_id, product_id, product_name, sku, quantity,
			unit_price, unit_cost, subtotal) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			t.ID, it.ProductID, it.ProductName, it.SKU, it.Quantity, it.UnitPrice, it.UnitCost, it.Subtotal).
			QueryRow(func(row pgx.Row) error { return row.Scan(&it.ID) })
	}
	for i := range t.Payments {
		p := &t.Payments[i]
		p.TransactionID = t.ID
		batch.Queue(`INSERT INTO payments (transaction_id, method, amount, tendered_amount, reference_number)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			t.ID, p.Method, p.Amount, p.TenderedAmount, p.ReferenceNumber).
			QueryRow(func(row pgx.Row) error { return row.Scan(&p.ID) })
	}
	return mapErr(q.SendBatch(ctx, batch).Close())
}

func (r *TransactionRepository) load(ctx context.Context, sql string, arg any) (*domain.Transaction, error) {
	q := conn(ctx, r.pool)
	t, err := scanTx(q.QueryRow(ctx, sql, arg))
	if err != nil {
		return nil, err
	}
	if err := r.loadChildren(ctx, q, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) loadChildren(ctx context.Context, q querier, t *domain.Transaction) error {
	rows, err := q.Query(ctx,
		`SELECT id, transaction_id, product_id, product_name, sku, quantity, unit_price, unit_cost, subtotal
		 FROM transaction_items WHERE transaction_id = $1 ORDER BY id`, t.ID)
	if err != nil {
		return err
	}
	t.Items, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.TransactionItem, error) {
		var it domain.TransactionItem
		err := row.Scan(&it.ID, &it.TransactionID, &it.ProductID, &it.ProductName, &it.SKU, &it.Quantity,
			&it.UnitPrice, &it.UnitCost, &it.Subtotal)
		return it, err
	})
	if err != nil {
		return err
	}
	rows, err = q.Query(ctx,
		`SELECT id, transaction_id, method, amount, tendered_amount, reference_number
		 FROM payments WHERE transaction_id = $1 ORDER BY id`, t.ID)
	if err != nil {
		return err
	}
	t.Payments, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Payment, error) {
		var p domain.Payment
		err := row.Scan(&p.ID, &p.TransactionID, &p.Method, &p.Amount, &p.TenderedAmount, &p.ReferenceNumber)
		return p, err
	})
	return err
}

func (r *TransactionRepository) FindByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return r.load(ctx, txSelect+` WHERE t.id = $1`, id)
}

func (r *TransactionRepository) FindByIDForUpdate(ctx context.Context, id int64) (*domain.Transaction, error) {
	return r.load(ctx, txSelect+` WHERE t.id = $1 FOR UPDATE OF t`, id)
}

func (r *TransactionRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	return r.load(ctx, txSelect+` WHERE t.idempotency_key = $1`, key)
}

func (r *TransactionRepository) MarkVoided(ctx context.Context, id, voidedBy int64, reason string, at time.Time) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE transactions SET status = 'VOIDED', voided_by = $2, void_reason = $3, voided_at = $4
		 WHERE id = $1 AND status = 'COMPLETED'`, id, voidedBy, reason, at))
}

func (r *TransactionRepository) NextInvoiceSequence(ctx context.Context, day time.Time) (int, error) {
	// The upsert takes a row lock on the day's counter, serialising concurrent checkouts
	// until the surrounding transaction commits; numbers are therefore gap-free.
	var n int
	err := conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO invoice_sequences (seq_date, last_number) VALUES ($1, 1)
		 ON CONFLICT (seq_date) DO UPDATE SET last_number = invoice_sequences.last_number + 1
		 RETURNING last_number`, day.Format("2006-01-02")).Scan(&n)
	return n, mapErr(err)
}

var txSorts = map[string]string{
	"created_at": "t.created_at", "total": "t.total", "invoice_number": "t.invoice_number", "cashier": "u.name",
}

func (r *TransactionRepository) List(ctx context.Context, f domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	w := &where{}
	w.search(f.Search, "t.invoice_number", "u.name")
	if f.Status != "" {
		w.add("t.status = ?", f.Status)
	}
	if f.UserID != nil {
		w.add("t.user_id = ?", *f.UserID)
	}
	if f.ShiftID != nil {
		w.add("t.shift_id = ?", *f.ShiftID)
	}
	if f.Method != "" {
		w.add("EXISTS (SELECT 1 FROM payments p WHERE p.transaction_id = t.id AND p.method = ?)", f.Method)
	}
	if f.DateFrom != nil {
		w.add("t.created_at >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		w.add("t.created_at < ?", *f.DateTo)
	}
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*) FROM transactions t JOIN users u ON u.id = t.user_id`+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(f.Params)
	rows, err := q.Query(ctx, txSelect+w.sql()+orderBy(f.Params, txSorts, "t.created_at DESC", "t.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Transaction, error) { return scanTx(row) })
	if err != nil {
		return nil, 0, err
	}
	// Attach payment methods so lists can show them without N+1 item queries.
	if len(items) > 0 {
		ids := make([]int64, len(items))
		idx := make(map[int64]int, len(items))
		for i, t := range items {
			ids[i] = t.ID
			idx[t.ID] = i
		}
		prow, err := q.Query(ctx,
			`SELECT id, transaction_id, method, amount, tendered_amount, reference_number
			 FROM payments WHERE transaction_id = ANY($1) ORDER BY id`, ids)
		if err != nil {
			return nil, 0, err
		}
		defer prow.Close()
		for prow.Next() {
			var p domain.Payment
			if err := prow.Scan(&p.ID, &p.TransactionID, &p.Method, &p.Amount, &p.TenderedAmount, &p.ReferenceNumber); err != nil {
				return nil, 0, err
			}
			i := idx[p.TransactionID]
			items[i].Payments = append(items[i].Payments, p)
		}
		if err := prow.Err(); err != nil {
			return nil, 0, err
		}
	}
	return items, total, nil
}
