package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
)

// ReportRepository is a read-only repository of aggregated sales data.
// Only COMPLETED transactions are counted.
type ReportRepository struct {
	pool     *pgxpool.Pool
	timezone string
}

func NewReportRepository(pool *pgxpool.Pool, timezone string) *ReportRepository {
	return &ReportRepository{pool: pool, timezone: timezone}
}

const summaryColumns = `COUNT(*), COALESCE(SUM(t.subtotal), 0), COALESCE(SUM(t.discount_amount), 0),
	COALESCE(SUM(t.tax_amount), 0), COALESCE(SUM(t.subtotal - t.discount_amount), 0), COALESCE(SUM(t.total), 0),
	COALESCE(SUM(t.total_cost), 0), COALESCE(SUM(t.subtotal - t.discount_amount - t.total_cost), 0)`

func scanSummary(row pgx.Row, extra ...any) (domain.SalesSummary, error) {
	var s domain.SalesSummary
	dest := append(extra, &s.TransactionCount, &s.GrossSales, &s.DiscountTotal, &s.TaxTotal, &s.NetSales,
		&s.TotalRevenue, &s.CostTotal, &s.GrossProfit)
	return s, row.Scan(dest...)
}

func (r *ReportRepository) Summary(ctx context.Context, rg domain.ReportRange) (*domain.SalesSummary, error) {
	s, err := scanSummary(conn(ctx, r.pool).QueryRow(ctx,
		`SELECT `+summaryColumns+` FROM transactions t
		 WHERE t.status = 'COMPLETED' AND t.created_at >= $1 AND t.created_at < $2`, rg.From, rg.To))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ReportRepository) SalesByPeriod(ctx context.Context, rg domain.ReportRange, monthly bool) ([]domain.SalesPeriodRow, error) {
	format := "YYYY-MM-DD"
	if monthly {
		format = "YYYY-MM"
	}
	rows, err := conn(ctx, r.pool).Query(ctx,
		`SELECT TO_CHAR(t.created_at AT TIME ZONE $3, '`+format+`') AS period, `+summaryColumns+`
		 FROM transactions t
		 WHERE t.status = 'COMPLETED' AND t.created_at >= $1 AND t.created_at < $2
		 GROUP BY period ORDER BY period`, rg.From, rg.To, r.timezone)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.SalesPeriodRow, error) {
		var out domain.SalesPeriodRow
		s, err := scanSummary(row, &out.Period)
		out.SalesSummary = s
		return out, err
	})
}

func (r *ReportRepository) TopProducts(ctx context.Context, rg domain.ReportRange, limit int) ([]domain.TopProductRow, error) {
	rows, err := conn(ctx, r.pool).Query(ctx,
		`SELECT i.product_id, MAX(i.product_name), MAX(i.sku), SUM(i.quantity), SUM(i.subtotal),
		        SUM(i.subtotal - i.unit_cost * i.quantity)
		 FROM transaction_items i JOIN transactions t ON t.id = i.transaction_id
		 WHERE t.status = 'COMPLETED' AND t.created_at >= $1 AND t.created_at < $2
		 GROUP BY i.product_id
		 ORDER BY SUM(i.quantity) DESC, SUM(i.subtotal) DESC
		 LIMIT $3`, rg.From, rg.To, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.TopProductRow, error) {
		var x domain.TopProductRow
		err := row.Scan(&x.ProductID, &x.ProductName, &x.SKU, &x.Quantity, &x.Revenue, &x.Profit)
		return x, err
	})
}

func (r *ReportRepository) ByPaymentMethod(ctx context.Context, rg domain.ReportRange) ([]domain.PaymentMethodRow, error) {
	rows, err := conn(ctx, r.pool).Query(ctx,
		`SELECT p.method, COUNT(DISTINCT p.transaction_id), SUM(p.amount)
		 FROM payments p JOIN transactions t ON t.id = p.transaction_id
		 WHERE t.status = 'COMPLETED' AND t.created_at >= $1 AND t.created_at < $2
		 GROUP BY p.method ORDER BY SUM(p.amount) DESC`, rg.From, rg.To)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.PaymentMethodRow, error) {
		var x domain.PaymentMethodRow
		err := row.Scan(&x.Method, &x.TransactionCount, &x.Amount)
		return x, err
	})
}

func (r *ReportRepository) ByCashier(ctx context.Context, rg domain.ReportRange) ([]domain.CashierRow, error) {
	rows, err := conn(ctx, r.pool).Query(ctx,
		`SELECT u.id, u.name, COUNT(t.id), COALESCE(SUM(t.total), 0)
		 FROM transactions t JOIN users u ON u.id = t.user_id
		 WHERE t.status = 'COMPLETED' AND t.created_at >= $1 AND t.created_at < $2
		 GROUP BY u.id, u.name ORDER BY SUM(t.total) DESC`, rg.From, rg.To)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.CashierRow, error) {
		var x domain.CashierRow
		err := row.Scan(&x.UserID, &x.UserName, &x.TransactionCount, &x.TotalSales)
		return x, err
	})
}
