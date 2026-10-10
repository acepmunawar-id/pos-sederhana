package domain

import (
	"context"
	"time"

	"github.com/posumkm/backend/pkg/pagination"
)

const (
	ShiftOpen   = "OPEN"
	ShiftClosed = "CLOSED"

	TxCompleted = "COMPLETED"
	TxVoided    = "VOIDED"

	DiscountNominal = "NOMINAL"
	DiscountPercent = "PERCENT"

	PaymentCash     = "CASH"
	PaymentQRIS     = "QRIS"
	PaymentTransfer = "TRANSFER"
	PaymentDebit    = "DEBIT"
)

type Shift struct {
	ID                int64
	UserID            int64
	UserName          string
	Status            string
	OpeningCash       int64
	ExpectedCash      *int64
	ActualCash        *int64
	CashDifference    *int64
	TotalSales        *int64
	TotalTransactions *int
	OpeningNote       *string
	ClosingNote       *string
	OpenedAt          time.Time
	ClosedAt          *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ShiftSummary aggregates completed transactions of a shift.
type ShiftSummary struct {
	TotalSales        int64
	TotalTransactions int
	VoidedCount       int
	CashSales         int64 // net cash kept in the drawer (tendered − change)
	PaymentTotals     map[string]int64
}

type ShiftFilter struct {
	pagination.Params
	UserID   *int64
	Status   string
	DateFrom *time.Time
	DateTo   *time.Time
}

type ShiftRepository interface {
	Create(ctx context.Context, s *Shift) error
	Close(ctx context.Context, s *Shift) error
	FindByID(ctx context.Context, id int64) (*Shift, error)
	FindOpenByUser(ctx context.Context, userID int64) (*Shift, error)
	Summary(ctx context.Context, shiftID int64) (*ShiftSummary, error)
	List(ctx context.Context, f ShiftFilter) ([]Shift, int64, error)
}

type Transaction struct {
	ID             int64
	InvoiceNumber  string
	IdempotencyKey string
	ShiftID        int64
	UserID         int64
	CashierName    string
	Subtotal       int64
	DiscountType   *string
	DiscountValue  int64
	DiscountAmount int64
	TaxRateBP      int
	TaxAmount      int64
	Total          int64
	PaidAmount     int64
	ChangeAmount   int64
	TotalCost      int64
	Status         string
	Note           *string
	VoidedAt       *time.Time
	VoidedBy       *int64
	VoidedByName   *string
	VoidReason     *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Items          []TransactionItem
	Payments       []Payment
}

type TransactionItem struct {
	ID            int64
	TransactionID int64
	ProductID     int64
	ProductName   string
	SKU           string
	Quantity      int
	UnitPrice     int64
	UnitCost      int64
	Subtotal      int64
}

type Payment struct {
	ID              int64
	TransactionID   int64
	Method          string
	Amount          int64
	TenderedAmount  int64
	ReferenceNumber *string
}

type TransactionFilter struct {
	pagination.Params
	Status   string
	UserID   *int64
	ShiftID  *int64
	Method   string
	DateFrom *time.Time
	DateTo   *time.Time
}

type TransactionRepository interface {
	// Create inserts the transaction together with its items and payments.
	Create(ctx context.Context, t *Transaction) error
	FindByID(ctx context.Context, id int64) (*Transaction, error)
	FindByIDForUpdate(ctx context.Context, id int64) (*Transaction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)
	MarkVoided(ctx context.Context, id, voidedBy int64, reason string, at time.Time) error
	// NextInvoiceSequence atomically increments and returns the counter for the given day.
	NextInvoiceSequence(ctx context.Context, day time.Time) (int, error)
	List(ctx context.Context, f TransactionFilter) ([]Transaction, int64, error)
}

type StoreSetting struct {
	ID            int64
	Name          string
	Address       *string
	Phone         *string
	LogoURL       *string
	TaxRateBP     int
	ReceiptFooter *string
	UpdatedAt     time.Time
}

type StoreSettingRepository interface {
	Get(ctx context.Context) (*StoreSetting, error)
	Update(ctx context.Context, s *StoreSetting) error
	UpdateLogo(ctx context.Context, url *string) error
}
