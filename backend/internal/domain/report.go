package domain

import (
	"context"
	"io"
	"time"
)

// ReportRange is a half-open interval [From, To) expressed in the store's timezone.
type ReportRange struct {
	From time.Time
	To   time.Time
}

type SalesSummary struct {
	TransactionCount int64
	GrossSales       int64 // sum of item subtotals before transaction discount
	DiscountTotal    int64
	TaxTotal         int64
	NetSales         int64 // gross − discount (excludes tax)
	TotalRevenue     int64 // amount paid by customers (net + tax)
	CostTotal        int64
	GrossProfit      int64 // net − cost
}

type SalesPeriodRow struct {
	Period string // YYYY-MM-DD or YYYY-MM
	SalesSummary
}

type TopProductRow struct {
	ProductID   int64
	ProductName string
	SKU         string
	Quantity    int64
	Revenue     int64
	Profit      int64
}

type PaymentMethodRow struct {
	Method           string
	TransactionCount int64
	Amount           int64
}

type CashierRow struct {
	UserID           int64
	UserName         string
	TransactionCount int64
	TotalSales       int64
}

type ReportRepository interface {
	Summary(ctx context.Context, r ReportRange) (*SalesSummary, error)
	SalesByPeriod(ctx context.Context, r ReportRange, monthly bool) ([]SalesPeriodRow, error)
	TopProducts(ctx context.Context, r ReportRange, limit int) ([]TopProductRow, error)
	ByPaymentMethod(ctx context.Context, r ReportRange) ([]PaymentMethodRow, error)
	ByCashier(ctx context.Context, r ReportRange) ([]CashierRow, error)
}

// FileStorage persists uploaded files and returns a public URL.
type FileStorage interface {
	Save(ctx context.Context, folder, ext string, r io.Reader) (url string, err error)
	Delete(ctx context.Context, url string) error
}
