package domain

import (
	"context"
	"time"

	"github.com/posumkm/backend/pkg/pagination"
)

const (
	MovementIn         = "IN"
	MovementAdjustment = "ADJUSTMENT"
	MovementSale       = "SALE"
	MovementVoid       = "VOID"

	RefTransaction = "TRANSACTION"
)

type StockMovement struct {
	ID            int64
	ProductID     int64
	ProductName   string
	ProductSKU    string
	UserID        int64
	UserName      string
	Type          string
	Quantity      int // signed: positive adds stock, negative removes it
	StockBefore   int
	StockAfter    int
	ReferenceType *string
	ReferenceID   *int64
	Note          *string
	CreatedAt     time.Time
}

type StockMovementFilter struct {
	pagination.Params
	ProductID *int64
	Type      string
	DateFrom  *time.Time
	DateTo    *time.Time // exclusive
}

type StockMovementRepository interface {
	Create(ctx context.Context, m *StockMovement) error
	List(ctx context.Context, f StockMovementFilter) ([]StockMovement, int64, error)
}
