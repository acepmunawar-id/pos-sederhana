package dto

import (
	"time"

	"github.com/posumkm/backend/internal/domain"
)

type StockInItem struct {
	ProductID int64 `json:"product_id" validate:"required,gt=0"`
	Quantity  int   `json:"quantity" validate:"required,gt=0,lte=1000000"`
}

// StockInRequest records goods received (e.g. from a supplier) for one or more products.
type StockInRequest struct {
	Items []StockInItem `json:"items" validate:"required,min=1,max=100,dive"`
	Note  *string       `json:"note" validate:"omitempty,max=500" example:"Pembelian dari Supplier A"`
}

// StockAdjustmentRequest sets the stock to the physically counted quantity (stock opname).
type StockAdjustmentRequest struct {
	ProductID   int64  `json:"product_id" validate:"required,gt=0"`
	ActualStock int    `json:"actual_stock" validate:"gte=0,lte=1000000"`
	Note        string `json:"note" validate:"required,min=3,max=500" example:"Barang rusak"`
}

type StockMovementResponse struct {
	ID            int64      `json:"id"`
	Product       ProductRef `json:"product"`
	User          IDName     `json:"user"`
	Type          string     `json:"type" enums:"IN,ADJUSTMENT,SALE,VOID"`
	Quantity      int        `json:"quantity"`
	StockBefore   int        `json:"stock_before"`
	StockAfter    int        `json:"stock_after"`
	ReferenceType *string    `json:"reference_type"`
	ReferenceID   *int64     `json:"reference_id"`
	Note          *string    `json:"note"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ProductRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	SKU  string `json:"sku"`
}

func ToStockMovementResponse(m *domain.StockMovement) StockMovementResponse {
	return StockMovementResponse{
		ID: m.ID, Product: ProductRef{ID: m.ProductID, Name: m.ProductName, SKU: m.ProductSKU},
		User: IDName{ID: m.UserID, Name: m.UserName}, Type: m.Type, Quantity: m.Quantity,
		StockBefore: m.StockBefore, StockAfter: m.StockAfter, ReferenceType: m.ReferenceType,
		ReferenceID: m.ReferenceID, Note: m.Note, CreatedAt: m.CreatedAt,
	}
}

func ToStockMovementResponses(items []domain.StockMovement) []StockMovementResponse {
	out := make([]StockMovementResponse, len(items))
	for i := range items {
		out[i] = ToStockMovementResponse(&items[i])
	}
	return out
}
