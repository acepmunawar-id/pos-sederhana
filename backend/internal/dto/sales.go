package dto

import (
	"time"

	"github.com/posumkm/backend/internal/domain"
)

// ---------- Shift ----------

type OpenShiftRequest struct {
	OpeningCash int64   `json:"opening_cash" validate:"gte=0" example:"200000"`
	Note        *string `json:"note" validate:"omitempty,max=500"`
}

type CloseShiftRequest struct {
	ActualCash int64   `json:"actual_cash" validate:"gte=0" example:"450000"`
	Note       *string `json:"note" validate:"omitempty,max=500"`
}

type ShiftResponse struct {
	ID                int64      `json:"id"`
	User              IDName     `json:"user"`
	Status            string     `json:"status" enums:"OPEN,CLOSED"`
	OpeningCash       int64      `json:"opening_cash"`
	ExpectedCash      *int64     `json:"expected_cash"`
	ActualCash        *int64     `json:"actual_cash"`
	CashDifference    *int64     `json:"cash_difference"`
	TotalSales        *int64     `json:"total_sales"`
	TotalTransactions *int       `json:"total_transactions"`
	OpeningNote       *string    `json:"opening_note"`
	ClosingNote       *string    `json:"closing_note"`
	OpenedAt          time.Time  `json:"opened_at"`
	ClosedAt          *time.Time `json:"closed_at"`
}

type ShiftSummaryResponse struct {
	TotalSales        int64            `json:"total_sales"`
	TotalTransactions int              `json:"total_transactions"`
	VoidedCount       int              `json:"voided_count"`
	CashSales         int64            `json:"cash_sales"`
	ExpectedCash      int64            `json:"expected_cash"`
	PaymentTotals     map[string]int64 `json:"payment_totals"`
}

type ShiftDetailResponse struct {
	ShiftResponse
	Summary ShiftSummaryResponse `json:"summary"`
}

func ToShiftResponse(s *domain.Shift) ShiftResponse {
	return ShiftResponse{
		ID: s.ID, User: IDName{ID: s.UserID, Name: s.UserName}, Status: s.Status, OpeningCash: s.OpeningCash,
		ExpectedCash: s.ExpectedCash, ActualCash: s.ActualCash, CashDifference: s.CashDifference,
		TotalSales: s.TotalSales, TotalTransactions: s.TotalTransactions, OpeningNote: s.OpeningNote,
		ClosingNote: s.ClosingNote, OpenedAt: s.OpenedAt, ClosedAt: s.ClosedAt,
	}
}

func ToShiftResponses(items []domain.Shift) []ShiftResponse {
	out := make([]ShiftResponse, len(items))
	for i := range items {
		out[i] = ToShiftResponse(&items[i])
	}
	return out
}

func ToShiftDetailResponse(s *domain.Shift, sum *domain.ShiftSummary) ShiftDetailResponse {
	return ShiftDetailResponse{
		ShiftResponse: ToShiftResponse(s),
		Summary: ShiftSummaryResponse{
			TotalSales: sum.TotalSales, TotalTransactions: sum.TotalTransactions, VoidedCount: sum.VoidedCount,
			CashSales: sum.CashSales, ExpectedCash: s.OpeningCash + sum.CashSales, PaymentTotals: sum.PaymentTotals,
		},
	}
}

// ---------- Checkout / Transaction ----------

type CheckoutItem struct {
	ProductID int64 `json:"product_id" validate:"required,gt=0" example:"1"`
	Quantity  int   `json:"quantity" validate:"required,gt=0,lte=10000" example:"2"`
}

type CheckoutPayment struct {
	Method string `json:"method" validate:"required,oneof=CASH QRIS TRANSFER DEBIT" example:"CASH"`
	// Amount handed over by the customer. For CASH it may exceed the bill (change is returned).
	Amount          int64   `json:"amount" validate:"gte=0" example:"50000"`
	ReferenceNumber *string `json:"reference_number" validate:"omitempty,max=100"`
}

type CheckoutRequest struct {
	Items        []CheckoutItem `json:"items" validate:"required,min=1,max=200,dive"`
	DiscountType *string        `json:"discount_type" validate:"omitempty,oneof=NOMINAL PERCENT" example:"PERCENT"`
	// NOMINAL: rupiah. PERCENT: basis points (1000 = 10%).
	DiscountValue int64             `json:"discount_value" validate:"gte=0" example:"1000"`
	Payments      []CheckoutPayment `json:"payments" validate:"required,min=1,max=4,dive"`
	Note          *string           `json:"note" validate:"omitempty,max=500"`
}

type VoidRequest struct {
	Reason string `json:"reason" validate:"required,min=3,max=500" example:"Salah input pesanan"`
}

type TransactionItemResponse struct {
	ID          int64  `json:"id"`
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Subtotal    int64  `json:"subtotal"`
}

type PaymentResponse struct {
	Method          string  `json:"method" enums:"CASH,QRIS,TRANSFER,DEBIT"`
	Amount          int64   `json:"amount"`
	TenderedAmount  int64   `json:"tendered_amount"`
	ReferenceNumber *string `json:"reference_number"`
}

type TransactionResponse struct {
	ID             int64                     `json:"id"`
	InvoiceNumber  string                    `json:"invoice_number" example:"INV/20261010/0001"`
	ShiftID        int64                     `json:"shift_id"`
	Cashier        IDName                    `json:"cashier"`
	Subtotal       int64                     `json:"subtotal"`
	DiscountType   *string                   `json:"discount_type"`
	DiscountValue  int64                     `json:"discount_value"`
	DiscountAmount int64                     `json:"discount_amount"`
	TaxRateBP      int                       `json:"tax_rate_bp"`
	TaxAmount      int64                     `json:"tax_amount"`
	Total          int64                     `json:"total"`
	PaidAmount     int64                     `json:"paid_amount"`
	ChangeAmount   int64                     `json:"change_amount"`
	GrossProfit    *int64                    `json:"gross_profit,omitempty"`
	Status         string                    `json:"status" enums:"COMPLETED,VOIDED"`
	Note           *string                   `json:"note"`
	VoidedAt       *time.Time                `json:"voided_at"`
	VoidedBy       *IDName                   `json:"voided_by"`
	VoidReason     *string                   `json:"void_reason"`
	ItemCount      int                       `json:"item_count"`
	Items          []TransactionItemResponse `json:"items,omitempty"`
	Payments       []PaymentResponse         `json:"payments"`
	CreatedAt      time.Time                 `json:"created_at"`
}

// ToTransactionResponse maps a transaction. includeProfit exposes cost-based fields (Owner/Admin only).
func ToTransactionResponse(t *domain.Transaction, includeProfit bool) TransactionResponse {
	res := TransactionResponse{
		ID: t.ID, InvoiceNumber: t.InvoiceNumber, ShiftID: t.ShiftID,
		Cashier:  IDName{ID: t.UserID, Name: t.CashierName},
		Subtotal: t.Subtotal, DiscountType: t.DiscountType, DiscountValue: t.DiscountValue,
		DiscountAmount: t.DiscountAmount, TaxRateBP: t.TaxRateBP, TaxAmount: t.TaxAmount, Total: t.Total,
		PaidAmount: t.PaidAmount, ChangeAmount: t.ChangeAmount, Status: t.Status, Note: t.Note,
		VoidedAt: t.VoidedAt, VoidReason: t.VoidReason, CreatedAt: t.CreatedAt,
		Payments: make([]PaymentResponse, len(t.Payments)),
	}
	if includeProfit {
		gp := t.Subtotal - t.DiscountAmount - t.TotalCost
		res.GrossProfit = &gp
	}
	if t.VoidedBy != nil {
		name := ""
		if t.VoidedByName != nil {
			name = *t.VoidedByName
		}
		res.VoidedBy = &IDName{ID: *t.VoidedBy, Name: name}
	}
	for i, p := range t.Payments {
		res.Payments[i] = PaymentResponse{Method: p.Method, Amount: p.Amount, TenderedAmount: p.TenderedAmount,
			ReferenceNumber: p.ReferenceNumber}
	}
	if t.Items != nil {
		res.Items = make([]TransactionItemResponse, len(t.Items))
		for i, it := range t.Items {
			res.ItemCount += it.Quantity
			res.Items[i] = TransactionItemResponse{ID: it.ID, ProductID: it.ProductID, ProductName: it.ProductName,
				SKU: it.SKU, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Subtotal: it.Subtotal}
		}
	}
	return res
}

func ToTransactionResponses(items []domain.Transaction, includeProfit bool) []TransactionResponse {
	out := make([]TransactionResponse, len(items))
	for i := range items {
		out[i] = ToTransactionResponse(&items[i], includeProfit)
	}
	return out
}

// ReceiptResponse bundles everything needed to render/print a receipt.
type ReceiptResponse struct {
	Store       StoreSettingResponse `json:"store"`
	Transaction TransactionResponse  `json:"transaction"`
	Text        string               `json:"text"` // 32-column plain-text receipt, ready to share
}

// ---------- Store settings ----------

type UpdateStoreSettingRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	Address       *string `json:"address" validate:"omitempty,max=500"`
	Phone         *string `json:"phone" validate:"omitempty,max=30"`
	TaxRateBP     int     `json:"tax_rate_bp" validate:"gte=0,lte=10000" example:"1100"`
	ReceiptFooter *string `json:"receipt_footer" validate:"omitempty,max=500"`
}

type StoreSettingResponse struct {
	Name          string    `json:"name"`
	Address       *string   `json:"address"`
	Phone         *string   `json:"phone"`
	LogoURL       *string   `json:"logo_url"`
	TaxRateBP     int       `json:"tax_rate_bp"`
	ReceiptFooter *string   `json:"receipt_footer"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToStoreSettingResponse(s *domain.StoreSetting) StoreSettingResponse {
	return StoreSettingResponse{Name: s.Name, Address: s.Address, Phone: s.Phone, LogoURL: s.LogoURL,
		TaxRateBP: s.TaxRateBP, ReceiptFooter: s.ReceiptFooter, UpdatedAt: s.UpdatedAt}
}
