package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------- Stock ----------

func TestStockIn_MergesItemsAndRecordsMovements(t *testing.T) {
	products, moves := &mockProductRepo{}, &mockMovementRepo{}
	svc := NewStockService(products, moves, &fakeTx{})
	products.On("FindByIDsForUpdate", mock.Anything, []int64{2, 1}).Return([]domain.Product{coffee(), tea()}, nil)
	products.On("UpdateStock", mock.Anything, int64(1), 15).Return(nil)
	products.On("UpdateStock", mock.Anything, int64(2), 12).Return(nil)
	moves.On("Create", mock.Anything, mock.Anything).Return(nil)

	out, err := svc.StockIn(context.Background(), owner, dto.StockInRequest{Items: []dto.StockInItem{
		{ProductID: 2, Quantity: 4}, {ProductID: 1, Quantity: 5}, {ProductID: 2, Quantity: 6}}})
	require.NoError(t, err)
	require.Len(t, out, 2)
	for _, m := range out {
		assert.Equal(t, domain.MovementIn, m.Type)
		assert.Equal(t, m.StockBefore+m.Quantity, m.StockAfter)
	}
	products.AssertExpectations(t)
}

func TestAdjust_RecordsSignedDifference(t *testing.T) {
	products, moves := &mockProductRepo{}, &mockMovementRepo{}
	svc := NewStockService(products, moves, &fakeTx{})
	products.On("FindByIDsForUpdate", mock.Anything, []int64{1}).Return([]domain.Product{coffee()}, nil)
	products.On("UpdateStock", mock.Anything, int64(1), 7).Return(nil)
	moves.On("Create", mock.Anything, mock.Anything).Return(nil)

	m, err := svc.Adjust(context.Background(), owner, dto.StockAdjustmentRequest{ProductID: 1, ActualStock: 7, Note: "Rusak"})
	require.NoError(t, err)
	assert.Equal(t, -3, m.Quantity)
	assert.Equal(t, domain.MovementAdjustment, m.Type)
	assert.Equal(t, "Rusak", *m.Note)
}

func TestAdjust_NoChangeRejected(t *testing.T) {
	products := &mockProductRepo{}
	svc := NewStockService(products, &mockMovementRepo{}, &fakeTx{})
	products.On("FindByIDsForUpdate", mock.Anything, []int64{1}).Return([]domain.Product{coffee()}, nil)
	_, err := svc.Adjust(context.Background(), owner, dto.StockAdjustmentRequest{ProductID: 1, ActualStock: 10, Note: "cek"})
	assertStatus(t, err, http.StatusBadRequest)
}

// ---------- Shift ----------

func TestShiftOpen_RejectsSecondOpenShift(t *testing.T) {
	shifts := &mockShiftRepo{}
	svc := NewShiftService(shifts, &fakeTx{}, fixedClock(time.Now()))
	shifts.On("FindOpenByUser", mock.Anything, cashier.UserID).Return(&domain.Shift{ID: 1}, nil)
	_, err := svc.Open(context.Background(), cashier, dto.OpenShiftRequest{OpeningCash: 100000})
	assertStatus(t, err, http.StatusConflict)
}

func TestShiftClose_ComputesCashDifference(t *testing.T) {
	shifts := &mockShiftRepo{}
	now := time.Date(2026, 10, 10, 22, 0, 0, 0, time.UTC)
	svc := NewShiftService(shifts, &fakeTx{}, fixedClock(now))
	shifts.On("FindOpenByUser", mock.Anything, cashier.UserID).
		Return(&domain.Shift{ID: 4, UserID: cashier.UserID, Status: domain.ShiftOpen, OpeningCash: 200000}, nil)
	shifts.On("Summary", mock.Anything, int64(4)).Return(&domain.ShiftSummary{TotalSales: 350000, TotalTransactions: 12, CashSales: 250000}, nil)
	var closed *domain.Shift
	shifts.On("Close", mock.Anything, mock.Anything).Run(func(a mock.Arguments) { closed = a.Get(1).(*domain.Shift) }).Return(nil)
	shifts.On("FindByID", mock.Anything, int64(4)).Return(&domain.Shift{ID: 4, Status: domain.ShiftClosed}, nil)

	_, sum, err := svc.Close(context.Background(), cashier, dto.CloseShiftRequest{ActualCash: 445000})
	require.NoError(t, err)
	assert.Equal(t, int64(450000), *closed.ExpectedCash)
	assert.Equal(t, int64(-5000), *closed.CashDifference, "Rp 5.000 short")
	assert.Equal(t, int64(350000), *closed.TotalSales)
	assert.Equal(t, 12, *closed.TotalTransactions)
	assert.Equal(t, now, *closed.ClosedAt)
	assert.Equal(t, int64(250000), sum.CashSales)
}

func TestShiftClose_NoOpenShift(t *testing.T) {
	shifts := &mockShiftRepo{}
	svc := NewShiftService(shifts, &fakeTx{}, fixedClock(time.Now()))
	shifts.On("FindOpenByUser", mock.Anything, cashier.UserID).Return(nil, domain.ErrNotFound)
	_, _, err := svc.Close(context.Background(), cashier, dto.CloseShiftRequest{})
	assertStatus(t, err, http.StatusNotFound)
}

func TestShiftList_CashierScopedToSelf(t *testing.T) {
	shifts := &mockShiftRepo{}
	svc := NewShiftService(shifts, &fakeTx{}, fixedClock(time.Now()))
	other := int64(99)
	shifts.On("List", mock.Anything, mock.MatchedBy(func(f domain.ShiftFilter) bool {
		return f.UserID != nil && *f.UserID == cashier.UserID
	})).Return([]domain.Shift{}, int64(0), nil)
	_, _, err := svc.List(context.Background(), cashier, domain.ShiftFilter{UserID: &other})
	require.NoError(t, err)
	shifts.AssertExpectations(t)
}

// ---------- Catalog ----------

func TestCategoryDelete_BlockedWhenUsed(t *testing.T) {
	repo := &mockCategoryRepo{}
	svc := NewCategoryService(repo)
	repo.On("FindByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1}, nil)
	repo.On("CountProducts", mock.Anything, int64(1)).Return(int64(3), nil)
	err := svc.Delete(context.Background(), 1)
	ae := assertStatus(t, err, http.StatusConflict)
	assert.Contains(t, ae.Message, "3 produk")
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func TestProductCreate_ValidatesReferencesAndUniqueness(t *testing.T) {
	products, cats, units := &mockProductRepo{}, &mockCategoryRepo{}, &mockUnitRepo{}
	svc := NewProductService(products, cats, units, &mockMovementRepo{}, &fakeTx{}, nil)
	cats.On("FindByID", mock.Anything, int64(9)).Return(nil, domain.ErrNotFound)
	units.On("FindByID", mock.Anything, int64(1)).Return(&domain.Unit{ID: 1}, nil)
	products.On("SKUExists", mock.Anything, "MNM-001", int64(0)).Return(true, nil)
	products.On("BarcodeExists", mock.Anything, "899", int64(0)).Return(false, nil)

	_, err := svc.Create(context.Background(), owner, dto.CreateProductRequest{
		CategoryID: 9, UnitID: 1, SKU: " MNM-001 ", Barcode: ptr("899"), Name: "Kopi", Price: 1000})
	ae := assertStatus(t, err, http.StatusUnprocessableEntity)
	fields := []string{}
	for _, e := range ae.Errors {
		fields = append(fields, e.Field)
	}
	assert.ElementsMatch(t, []string{"category_id", "sku"}, fields)
}

func TestProductCreate_InitialStockRecordsMovement(t *testing.T) {
	products, cats, units, moves := &mockProductRepo{}, &mockCategoryRepo{}, &mockUnitRepo{}, &mockMovementRepo{}
	svc := NewProductService(products, cats, units, moves, &fakeTx{}, nil)
	cats.On("FindByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1}, nil)
	units.On("FindByID", mock.Anything, int64(1)).Return(&domain.Unit{ID: 1}, nil)
	products.On("SKUExists", mock.Anything, "NEW-1", int64(0)).Return(false, nil)
	products.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) { a.Get(1).(*domain.Product).ID = 20 }).Return(nil)
	moves.On("Create", mock.Anything, mock.MatchedBy(func(m *domain.StockMovement) bool {
		return m.ProductID == 20 && m.Type == domain.MovementIn && m.Quantity == 12 && m.StockAfter == 12
	})).Return(nil)
	products.On("FindByID", mock.Anything, int64(20)).Return(&domain.Product{ID: 20}, nil)

	_, err := svc.Create(context.Background(), owner, dto.CreateProductRequest{
		CategoryID: 1, UnitID: 1, SKU: "NEW-1", Barcode: ptr("  "), Name: "Baru", Price: 1000, InitialStock: 12})
	require.NoError(t, err)
	moves.AssertExpectations(t)
	products.AssertNotCalled(t, "BarcodeExists", mock.Anything, mock.Anything, mock.Anything)
}

// ---------- Formatting ----------

func TestFormatRupiah(t *testing.T) {
	assert.Equal(t, "Rp 0", FormatRupiah(0))
	assert.Equal(t, "Rp 500", FormatRupiah(500))
	assert.Equal(t, "Rp 25.000", FormatRupiah(25000))
	assert.Equal(t, "Rp 1.250.000", FormatRupiah(1250000))
	assert.Equal(t, "-Rp 5.000", FormatRupiah(-5000))
}

func TestFormatPercentBP(t *testing.T) {
	assert.Equal(t, "11%", FormatPercentBP(1100))
	assert.Equal(t, "12,5%", FormatPercentBP(1250))
	assert.Equal(t, "0,75%", FormatPercentBP(75))
}

func TestBuildReceiptText_FitsWidth(t *testing.T) {
	st := &domain.StoreSetting{Name: "Warung Kopi Nusantara Dengan Nama Sangat Panjang Sekali", ReceiptFooter: ptr("Terima kasih")}
	tr := &domain.Transaction{InvoiceNumber: "INV/20261010/0001", CashierName: "Siti", Subtotal: 36000, Total: 36000,
		PaidAmount: 50000, ChangeAmount: 14000, CreatedAt: time.Now(), Status: domain.TxVoided,
		Items:    []domain.TransactionItem{{ProductName: "Kopi Susu Gula Aren Ukuran Jumbo Spesial", Quantity: 2, UnitPrice: 18000, Subtotal: 36000}},
		Payments: []domain.Payment{{Method: domain.PaymentCash, Amount: 36000, TenderedAmount: 50000}}}
	text := BuildReceiptText(st, tr, jakarta)
	for _, line := range splitLines(text) {
		assert.LessOrEqual(t, len([]rune(line)), receiptWidth, "line too wide: %q", line)
	}
	assert.Contains(t, text, "Kembalian")
	assert.Contains(t, text, "DIBATALKAN")
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return out
}
