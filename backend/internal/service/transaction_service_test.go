package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var jakarta, _ = time.LoadLocation("Asia/Jakarta")

type txDeps struct {
	txs      *mockTxRepo
	products *mockProductRepo
	moves    *mockMovementRepo
	shifts   *mockShiftRepo
	settings *mockSettingRepo
	tx       *fakeTx
	svc      *TransactionService
}

// 2026-10-10 23:30 WIB = 2026-10-10 16:30 UTC.
var checkoutNow = time.Date(2026, 10, 10, 23, 30, 0, 0, jakarta)

func newTxDeps() *txDeps {
	d := &txDeps{txs: &mockTxRepo{}, products: &mockProductRepo{}, moves: &mockMovementRepo{},
		shifts: &mockShiftRepo{}, settings: &mockSettingRepo{}, tx: &fakeTx{}}
	d.svc = NewTransactionService(d.txs, d.products, d.moves, d.shifts, d.settings, d.tx, jakarta, fixedClock(checkoutNow.UTC()))
	return d
}

var cashier = Actor{UserID: 7, Role: domain.RoleCashier}
var owner = Actor{UserID: 1, Role: domain.RoleOwner}

func coffee() domain.Product {
	return domain.Product{ID: 1, Name: "Kopi Susu", SKU: "MNM-001", Price: 18000, Cost: 7000, Stock: 10, IsActive: true}
}
func tea() domain.Product {
	return domain.Product{ID: 2, Name: "Es Teh", SKU: "MNM-002", Price: 5000, Cost: 1500, Stock: 2, IsActive: true}
}

func (d *txDeps) expectReady(taxBP int) {
	d.txs.On("FindByIdempotencyKey", mock.Anything, "key-12345678").Return(nil, domain.ErrNotFound).Once()
	d.shifts.On("FindOpenByUser", mock.Anything, cashier.UserID).Return(&domain.Shift{ID: 3, UserID: cashier.UserID, Status: domain.ShiftOpen}, nil)
	d.settings.On("Get", mock.Anything).Return(&domain.StoreSetting{TaxRateBP: taxBP}, nil)
}

func TestCheckout_ComputesTotalsAndDecrementsStock(t *testing.T) {
	d := newTxDeps()
	d.expectReady(1100) // 11%
	d.products.On("FindByIDsForUpdate", mock.Anything, []int64{1, 2}).Return([]domain.Product{coffee(), tea()}, nil)
	d.txs.On("NextInvoiceSequence", mock.Anything, time.Date(2026, 10, 10, 0, 0, 0, 0, jakarta)).Return(42, nil)

	var saved *domain.Transaction
	d.txs.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Run(func(a mock.Arguments) {
		saved = a.Get(1).(*domain.Transaction)
		saved.ID = 99
	}).Return(nil)
	d.products.On("UpdateStock", mock.Anything, int64(1), 8).Return(nil)
	d.products.On("UpdateStock", mock.Anything, int64(2), 0).Return(nil)
	d.moves.On("Create", mock.Anything, mock.AnythingOfType("*domain.StockMovement")).Return(nil).Twice()
	d.txs.On("FindByID", mock.Anything, int64(99)).Return(&domain.Transaction{ID: 99}, nil)

	req := dto.CheckoutRequest{
		Items:        []dto.CheckoutItem{{ProductID: 2, Quantity: 1}, {ProductID: 1, Quantity: 2}, {ProductID: 2, Quantity: 1}},
		DiscountType: ptr(domain.DiscountPercent), DiscountValue: 1000, // 10%
		Payments: []dto.CheckoutPayment{{Method: domain.PaymentCash, Amount: 100000}},
	}
	_, created, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", req)
	require.NoError(t, err)
	assert.True(t, created)

	// subtotal = 2×18.000 + 2×5.000 = 46.000; discount 10% = 4.600; taxable 41.400; tax 11% = 4.554
	assert.Equal(t, int64(46000), saved.Subtotal)
	assert.Equal(t, int64(4600), saved.DiscountAmount)
	assert.Equal(t, int64(4554), saved.TaxAmount)
	assert.Equal(t, int64(45954), saved.Total)
	assert.Equal(t, int64(100000), saved.PaidAmount)
	assert.Equal(t, int64(54046), saved.ChangeAmount)
	assert.Equal(t, int64(2*7000+2*1500), saved.TotalCost)
	assert.Equal(t, "INV/20261010/0042", saved.InvoiceNumber, "invoice date uses store timezone")
	assert.Equal(t, int64(3), saved.ShiftID)
	require.Len(t, saved.Items, 2, "duplicate product lines are merged")
	require.Len(t, saved.Payments, 1)
	assert.Equal(t, int64(45954), saved.Payments[0].Amount, "applied cash = tendered − change")
	assert.Equal(t, int64(100000), saved.Payments[0].TenderedAmount)

	d.moves.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(m *domain.StockMovement) bool {
		return m.ProductID == 1 && m.Type == domain.MovementSale && m.Quantity == -2 && m.StockBefore == 10 &&
			m.StockAfter == 8 && *m.ReferenceID == 99
	}))
	assert.Equal(t, 1, d.tx.calls, "everything runs in one DB transaction")
}

func TestCheckout_IdempotentReplayReturnsExisting(t *testing.T) {
	d := newTxDeps()
	existing := &domain.Transaction{ID: 5, UserID: cashier.UserID, InvoiceNumber: "INV/20261010/0001"}
	d.txs.On("FindByIdempotencyKey", mock.Anything, "key-12345678").Return(existing, nil)

	got, created, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Same(t, existing, got)
	d.txs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	assert.Equal(t, 0, d.tx.calls)
}

func TestCheckout_IdempotencyKeyOfOtherUserConflicts(t *testing.T) {
	d := newTxDeps()
	d.txs.On("FindByIdempotencyKey", mock.Anything, "key-12345678").Return(&domain.Transaction{ID: 5, UserID: 999}, nil)
	_, _, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{})
	assertStatus(t, err, http.StatusConflict)
}

func TestCheckout_RequiresIdempotencyKey(t *testing.T) {
	d := newTxDeps()
	_, _, err := d.svc.Checkout(context.Background(), cashier, "short", dto.CheckoutRequest{})
	assertStatus(t, err, http.StatusBadRequest)
}

func TestCheckout_RequiresOpenShift(t *testing.T) {
	d := newTxDeps()
	d.txs.On("FindByIdempotencyKey", mock.Anything, "key-12345678").Return(nil, domain.ErrNotFound)
	d.shifts.On("FindOpenByUser", mock.Anything, cashier.UserID).Return(nil, domain.ErrNotFound)
	_, _, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{})
	assertStatus(t, err, http.StatusUnprocessableEntity)
}

func TestCheckout_InsufficientStockWritesNothing(t *testing.T) {
	d := newTxDeps()
	d.expectReady(0)
	d.products.On("FindByIDsForUpdate", mock.Anything, []int64{2}).Return([]domain.Product{tea()}, nil)

	_, _, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{
		Items:    []dto.CheckoutItem{{ProductID: 2, Quantity: 3}},
		Payments: []dto.CheckoutPayment{{Method: domain.PaymentCash, Amount: 20000}},
	})
	ae := assertStatus(t, err, http.StatusUnprocessableEntity)
	require.Len(t, ae.Errors, 1)
	assert.Contains(t, ae.Errors[0].Message, "tersisa 2")
	d.txs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	d.products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckout_UnknownProduct(t *testing.T) {
	d := newTxDeps()
	d.expectReady(0)
	d.products.On("FindByIDsForUpdate", mock.Anything, []int64{1, 50}).Return([]domain.Product{coffee()}, nil)
	_, _, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{
		Items:    []dto.CheckoutItem{{ProductID: 50, Quantity: 1}, {ProductID: 1, Quantity: 1}},
		Payments: []dto.CheckoutPayment{{Method: domain.PaymentCash, Amount: 20000}},
	})
	assertStatus(t, err, http.StatusNotFound)
}

func TestCheckout_InactiveProduct(t *testing.T) {
	d := newTxDeps()
	d.expectReady(0)
	p := coffee()
	p.IsActive = false
	d.products.On("FindByIDsForUpdate", mock.Anything, []int64{1}).Return([]domain.Product{p}, nil)
	_, _, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{
		Items:    []dto.CheckoutItem{{ProductID: 1, Quantity: 1}},
		Payments: []dto.CheckoutPayment{{Method: domain.PaymentCash, Amount: 20000}},
	})
	assertStatus(t, err, http.StatusUnprocessableEntity)
}

func TestCheckout_ConcurrentDuplicateKeyReturnsWinner(t *testing.T) {
	d := newTxDeps()
	d.expectReady(0)
	d.products.On("FindByIDsForUpdate", mock.Anything, []int64{1}).Return([]domain.Product{coffee()}, nil)
	d.txs.On("NextInvoiceSequence", mock.Anything, mock.Anything).Return(1, nil)
	d.txs.On("Create", mock.Anything, mock.Anything).Return(domain.ErrDuplicate)
	winner := &domain.Transaction{ID: 77, UserID: cashier.UserID}
	d.txs.On("FindByIdempotencyKey", mock.Anything, "key-12345678").Return(winner, nil).Once()

	got, created, err := d.svc.Checkout(context.Background(), cashier, "key-12345678", dto.CheckoutRequest{
		Items:    []dto.CheckoutItem{{ProductID: 1, Quantity: 1}},
		Payments: []dto.CheckoutPayment{{Method: domain.PaymentCash, Amount: 20000}},
	})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Same(t, winner, got)
}

func TestApplyPricing(t *testing.T) {
	cases := []struct {
		name        string
		subtotal    int64
		taxBP       int
		dType       *string
		dValue      int64
		wantDisc    int64
		wantTax     int64
		wantTotal   int64
		wantErrCode int
	}{
		{name: "no discount, no tax", subtotal: 25000, wantTotal: 25000},
		{name: "nominal discount + 10% tax", subtotal: 25000, taxBP: 1000, dType: ptr(domain.DiscountNominal), dValue: 5000,
			wantDisc: 5000, wantTax: 2000, wantTotal: 22000},
		{name: "percent rounds half up", subtotal: 12345, dType: ptr(domain.DiscountPercent), dValue: 1250, // 12.5%
			wantDisc: 1543, wantTotal: 10802},
		{name: "100% discount", subtotal: 9000, taxBP: 1100, dType: ptr(domain.DiscountPercent), dValue: 10000,
			wantDisc: 9000, wantTotal: 0},
		{name: "nominal above subtotal", subtotal: 1000, dType: ptr(domain.DiscountNominal), dValue: 2000, wantErrCode: 422},
		{name: "percent above 100", subtotal: 1000, dType: ptr(domain.DiscountPercent), dValue: 10001, wantErrCode: 422},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &domain.Transaction{Subtotal: tc.subtotal, TaxRateBP: tc.taxBP}
			err := applyPricing(tr, tc.dType, tc.dValue)
			if tc.wantErrCode != 0 {
				assertStatus(t, err, tc.wantErrCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantDisc, tr.DiscountAmount)
			assert.Equal(t, tc.wantTax, tr.TaxAmount)
			assert.Equal(t, tc.wantTotal, tr.Total)
		})
	}
}

func TestApplyPayments(t *testing.T) {
	pay := func(m string, a int64) dto.CheckoutPayment { return dto.CheckoutPayment{Method: m, Amount: a} }
	t.Run("exact non-cash", func(t *testing.T) {
		tr := &domain.Transaction{Total: 22000}
		require.NoError(t, applyPayments(tr, []dto.CheckoutPayment{pay(domain.PaymentQRIS, 22000)}))
		assert.Equal(t, int64(0), tr.ChangeAmount)
	})
	t.Run("split with cash change", func(t *testing.T) {
		tr := &domain.Transaction{Total: 22000}
		require.NoError(t, applyPayments(tr, []dto.CheckoutPayment{pay(domain.PaymentQRIS, 10000), pay(domain.PaymentCash, 20000)}))
		assert.Equal(t, int64(8000), tr.ChangeAmount)
		assert.Equal(t, int64(30000), tr.PaidAmount)
		assert.Equal(t, int64(12000), tr.Payments[1].Amount)
	})
	t.Run("underpaid", func(t *testing.T) {
		ae := assertStatus(t, applyPayments(&domain.Transaction{Total: 5500}, []dto.CheckoutPayment{pay(domain.PaymentCash, 1000)}), 422)
		assert.Equal(t, "pembayaran kurang Rp 4.500", ae.Errors[0].Message)
	})
	t.Run("non-cash overpay", func(t *testing.T) {
		assertStatus(t, applyPayments(&domain.Transaction{Total: 5000}, []dto.CheckoutPayment{pay(domain.PaymentDebit, 6000)}), 422)
	})
	t.Run("duplicate method", func(t *testing.T) {
		assertStatus(t, applyPayments(&domain.Transaction{Total: 5000},
			[]dto.CheckoutPayment{pay(domain.PaymentCash, 2000), pay(domain.PaymentCash, 3000)}), 422)
	})
	t.Run("unneeded cash", func(t *testing.T) {
		assertStatus(t, applyPayments(&domain.Transaction{Total: 5000},
			[]dto.CheckoutPayment{pay(domain.PaymentQRIS, 5000), pay(domain.PaymentCash, 1000)}), 422)
	})
}

func TestVoid_RestoresStockAndMarksVoided(t *testing.T) {
	d := newTxDeps()
	orig := &domain.Transaction{ID: 10, InvoiceNumber: "INV/20261010/0003", Status: domain.TxCompleted,
		Items: []domain.TransactionItem{{ProductID: 2, Quantity: 1}, {ProductID: 1, Quantity: 3}}}
	d.txs.On("FindByIDForUpdate", mock.Anything, int64(10)).Return(orig, nil)
	d.products.On("FindByIDsForUpdate", mock.Anything, []int64{1, 2}).Return([]domain.Product{coffee(), tea()}, nil)
	d.products.On("UpdateStock", mock.Anything, int64(1), 13).Return(nil)
	d.products.On("UpdateStock", mock.Anything, int64(2), 3).Return(nil)
	d.moves.On("Create", mock.Anything, mock.MatchedBy(func(m *domain.StockMovement) bool {
		return m.Type == domain.MovementVoid && m.Quantity > 0
	})).Return(nil).Twice()
	d.txs.On("MarkVoided", mock.Anything, int64(10), owner.UserID, "Salah input", mock.Anything).Return(nil)
	d.txs.On("FindByID", mock.Anything, int64(10)).Return(&domain.Transaction{ID: 10, Status: domain.TxVoided}, nil)

	got, err := d.svc.Void(context.Background(), owner, 10, dto.VoidRequest{Reason: " Salah input "})
	require.NoError(t, err)
	assert.Equal(t, domain.TxVoided, got.Status)
	d.products.AssertExpectations(t)
	d.moves.AssertExpectations(t)
}

func TestVoid_AlreadyVoided(t *testing.T) {
	d := newTxDeps()
	d.txs.On("FindByIDForUpdate", mock.Anything, int64(10)).Return(&domain.Transaction{ID: 10, Status: domain.TxVoided}, nil)
	_, err := d.svc.Void(context.Background(), owner, 10, dto.VoidRequest{Reason: "x y z"})
	assertStatus(t, err, http.StatusConflict)
	d.products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
}

func TestVoid_CashierForbidden(t *testing.T) {
	d := newTxDeps()
	_, err := d.svc.Void(context.Background(), cashier, 10, dto.VoidRequest{Reason: "x y z"})
	assertStatus(t, err, http.StatusForbidden)
}

func TestGet_CashierCannotSeeOthersTransaction(t *testing.T) {
	d := newTxDeps()
	d.txs.On("FindByID", mock.Anything, int64(4)).Return(&domain.Transaction{ID: 4, UserID: 999}, nil)
	_, err := d.svc.Get(context.Background(), cashier, 4)
	assertStatus(t, err, http.StatusNotFound)

	got, err := d.svc.Get(context.Background(), owner, 4)
	require.NoError(t, err)
	assert.Equal(t, int64(4), got.ID)
}

func TestFormatInvoiceNumber(t *testing.T) {
	assert.Equal(t, "INV/20260101/0001", FormatInvoiceNumber(time.Date(2026, 1, 1, 0, 0, 0, 0, jakarta), 1))
	assert.Equal(t, "INV/20261231/12345", FormatInvoiceNumber(time.Date(2026, 12, 31, 0, 0, 0, 0, jakarta), 12345))
}

func assertStatus(t *testing.T, err error, status int) *apperror.AppError {
	t.Helper()
	require.Error(t, err)
	var ae *apperror.AppError
	require.True(t, errors.As(err, &ae), "expected *AppError, got %T: %v", err, err)
	assert.Equal(t, status, ae.Status, ae.Message)
	return ae
}
