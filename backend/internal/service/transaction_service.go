package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
)

type TransactionService struct {
	transactions domain.TransactionRepository
	products     domain.ProductRepository
	movements    domain.StockMovementRepository
	shifts       domain.ShiftRepository
	settings     domain.StoreSettingRepository
	tx           domain.TxManager
	loc          *time.Location
	now          Clock
}

func NewTransactionService(transactions domain.TransactionRepository, products domain.ProductRepository,
	movements domain.StockMovementRepository, shifts domain.ShiftRepository, settings domain.StoreSettingRepository,
	tx domain.TxManager, loc *time.Location, now Clock) *TransactionService {
	return &TransactionService{transactions: transactions, products: products, movements: movements,
		shifts: shifts, settings: settings, tx: tx, loc: loc, now: now}
}

// Checkout creates a sale atomically: transaction, items, payments, stock decrement and
// stock movements either all succeed or none do. Product rows are locked with
// SELECT ... FOR UPDATE so concurrent checkouts can never push stock below zero.
// Re-sending the same idempotency key returns the original transaction instead of a duplicate.
// The returned bool is true when a new transaction was created.
func (s *TransactionService) Checkout(ctx context.Context, actor Actor, idempotencyKey string, req dto.CheckoutRequest) (*domain.Transaction, bool, error) {
	key := strings.TrimSpace(idempotencyKey)
	if len(key) < 8 || len(key) > 100 {
		return nil, false, apperror.BadRequest("Header Idempotency-Key wajib diisi (8–100 karakter)")
	}
	if existing, err := s.replay(ctx, actor, key); existing != nil || err != nil {
		return existing, false, err
	}

	shift, err := s.shifts.FindOpenByUser(ctx, actor.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, false, apperror.Unprocessable("Buka shift terlebih dahulu sebelum melakukan transaksi")
	}
	if err != nil {
		return nil, false, err
	}
	setting, err := s.settings.Get(ctx)
	if err != nil {
		return nil, false, err
	}

	qty := map[int64]int{}
	var ids []int64
	for _, it := range req.Items {
		if _, seen := qty[it.ProductID]; !seen {
			ids = append(ids, it.ProductID)
		}
		qty[it.ProductID] += it.Quantity
	}
	slices.Sort(ids)

	var txID int64
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		products, err := s.products.FindByIDsForUpdate(ctx, ids)
		if err != nil {
			return err
		}
		if err := ensureAllFound(ids, products); err != nil {
			return err
		}

		t := &domain.Transaction{
			IdempotencyKey: key, ShiftID: shift.ID, UserID: actor.UserID, Status: domain.TxCompleted,
			TaxRateBP: setting.TaxRateBP, Note: trimPtr(req.Note),
		}
		var stockErrs []apperror.FieldError
		for _, p := range products {
			q := qty[p.ID]
			if !p.IsActive {
				return apperror.Unprocessable(fmt.Sprintf("Produk %s sedang tidak aktif", p.Name))
			}
			if p.Stock < q {
				stockErrs = append(stockErrs, apperror.FieldError{
					Field:   fmt.Sprintf("items[product_id=%d]", p.ID),
					Message: fmt.Sprintf("stok %s tersisa %d", p.Name, p.Stock),
				})
				continue
			}
			line := p.Price * int64(q)
			t.Subtotal += line
			t.TotalCost += p.Cost * int64(q)
			t.Items = append(t.Items, domain.TransactionItem{
				ProductID: p.ID, ProductName: p.Name, SKU: p.SKU, Quantity: q,
				UnitPrice: p.Price, UnitCost: p.Cost, Subtotal: line,
			})
		}
		if len(stockErrs) > 0 {
			return &apperror.AppError{Status: http.StatusUnprocessableEntity, Message: "Stok tidak mencukupi", Errors: stockErrs}
		}

		if err := applyPricing(t, req.DiscountType, req.DiscountValue); err != nil {
			return err
		}
		if err := applyPayments(t, req.Payments); err != nil {
			return err
		}

		now := s.now()
		seq, err := s.transactions.NextInvoiceSequence(ctx, startOfDay(now, s.loc))
		if err != nil {
			return err
		}
		t.InvoiceNumber = FormatInvoiceNumber(now.In(s.loc), seq)

		if err := s.transactions.Create(ctx, t); err != nil {
			return err
		}
		ref := domain.RefTransaction
		note := "Penjualan " + t.InvoiceNumber
		for _, p := range products {
			q := qty[p.ID]
			if err := s.products.UpdateStock(ctx, p.ID, p.Stock-q); err != nil {
				return err
			}
			if err := s.movements.Create(ctx, &domain.StockMovement{
				ProductID: p.ID, UserID: actor.UserID, Type: domain.MovementSale, Quantity: -q,
				StockBefore: p.Stock, StockAfter: p.Stock - q, ReferenceType: &ref, ReferenceID: &t.ID, Note: &note,
			}); err != nil {
				return err
			}
		}
		txID = t.ID
		return nil
	})
	if errors.Is(err, domain.ErrDuplicate) {
		// A concurrent request with the same key won the race; return its result.
		if existing, rerr := s.replay(ctx, actor, key); existing != nil || rerr != nil {
			return existing, false, rerr
		}
	}
	if err != nil {
		return nil, false, err
	}
	t, err := s.transactions.FindByID(ctx, txID)
	return t, true, err
}

func (s *TransactionService) replay(ctx context.Context, actor Actor, key string) (*domain.Transaction, error) {
	existing, err := s.transactions.FindByIdempotencyKey(ctx, key)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if existing.UserID != actor.UserID {
		return nil, apperror.Conflict("Idempotency-Key sudah dipakai oleh transaksi lain")
	}
	return existing, nil
}

// applyPricing computes discount, tax and total on t (Subtotal must be set).
// Tax is charged on the subtotal after discount.
func applyPricing(t *domain.Transaction, discountType *string, discountValue int64) error {
	if discountType != nil && discountValue > 0 {
		switch *discountType {
		case domain.DiscountNominal:
			if discountValue > t.Subtotal {
				return apperror.Validation([]apperror.FieldError{{Field: "discount_value", Message: "diskon melebihi subtotal"}})
			}
			t.DiscountAmount = discountValue
		case domain.DiscountPercent:
			if discountValue > domain.BasisPoints {
				return apperror.Validation([]apperror.FieldError{{Field: "discount_value", Message: "diskon persen maksimal 100%"}})
			}
			t.DiscountAmount = mulBP(t.Subtotal, discountValue)
		}
		dt := *discountType
		t.DiscountType, t.DiscountValue = &dt, discountValue
	}
	taxable := t.Subtotal - t.DiscountAmount
	t.TaxAmount = mulBP(taxable, int64(t.TaxRateBP))
	t.Total = taxable + t.TaxAmount
	return nil
}

// applyPayments validates the tendered amounts and derives paid, change and the applied
// amount per method. Change can only come from cash: non-cash payments may not exceed the total.
func applyPayments(t *domain.Transaction, in []dto.CheckoutPayment) error {
	seen := map[string]bool{}
	var nonCash, cash int64
	hasCash := false
	for _, p := range in {
		if seen[p.Method] {
			return apperror.Validation([]apperror.FieldError{{Field: "payments", Message: "metode bayar tidak boleh dobel"}})
		}
		seen[p.Method] = true
		if p.Method == domain.PaymentCash {
			cash, hasCash = p.Amount, true
		} else {
			nonCash += p.Amount
		}
	}
	if nonCash > t.Total {
		return apperror.Validation([]apperror.FieldError{{Field: "payments", Message: "pembayaran non-tunai melebihi total belanja"}})
	}
	paid := nonCash + cash
	if paid < t.Total {
		return apperror.Validation([]apperror.FieldError{{Field: "payments",
			Message: fmt.Sprintf("pembayaran kurang %s", FormatRupiah(t.Total-paid))}})
	}
	change := paid - t.Total
	if hasCash && cash == change && t.Total > 0 {
		return apperror.Validation([]apperror.FieldError{{Field: "payments", Message: "pembayaran tunai tidak diperlukan"}})
	}
	t.PaidAmount, t.ChangeAmount = paid, change
	t.Payments = make([]domain.Payment, 0, len(in))
	for _, p := range in {
		applied := p.Amount
		if p.Method == domain.PaymentCash {
			applied = p.Amount - change
		}
		t.Payments = append(t.Payments, domain.Payment{
			Method: p.Method, Amount: applied, TenderedAmount: p.Amount, ReferenceNumber: trimPtr(p.ReferenceNumber),
		})
	}
	return nil
}

// Void cancels a completed transaction and returns its items to stock.
func (s *TransactionService) Void(ctx context.Context, actor Actor, id int64, req dto.VoidRequest) (*domain.Transaction, error) {
	if !actor.IsManager() {
		return nil, apperror.Forbidden("Hanya Owner atau Admin yang dapat membatalkan transaksi")
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := s.transactions.FindByIDForUpdate(ctx, id)
		if err != nil {
			return notFound(err, "Transaksi tidak ditemukan")
		}
		if t.Status == domain.TxVoided {
			return apperror.Conflict("Transaksi sudah dibatalkan sebelumnya")
		}
		qty := map[int64]int{}
		var ids []int64
		for _, it := range t.Items {
			if _, seen := qty[it.ProductID]; !seen {
				ids = append(ids, it.ProductID)
			}
			qty[it.ProductID] += it.Quantity
		}
		slices.Sort(ids)
		// Soft-deleted products are not returned here; their stock no longer matters.
		products, err := s.products.FindByIDsForUpdate(ctx, ids)
		if err != nil {
			return err
		}
		ref := domain.RefTransaction
		note := "Void " + t.InvoiceNumber + ": " + strings.TrimSpace(req.Reason)
		for _, p := range products {
			q := qty[p.ID]
			if err := s.products.UpdateStock(ctx, p.ID, p.Stock+q); err != nil {
				return err
			}
			if err := s.movements.Create(ctx, &domain.StockMovement{
				ProductID: p.ID, UserID: actor.UserID, Type: domain.MovementVoid, Quantity: q,
				StockBefore: p.Stock, StockAfter: p.Stock + q, ReferenceType: &ref, ReferenceID: &t.ID, Note: &note,
			}); err != nil {
				return err
			}
		}
		return s.transactions.MarkVoided(ctx, id, actor.UserID, strings.TrimSpace(req.Reason), s.now())
	})
	if err != nil {
		return nil, err
	}
	return s.transactions.FindByID(ctx, id)
}

// Get returns a transaction; cashiers may only see their own.
func (s *TransactionService) Get(ctx context.Context, actor Actor, id int64) (*domain.Transaction, error) {
	t, err := s.transactions.FindByID(ctx, id)
	if err != nil {
		return nil, notFound(err, "Transaksi tidak ditemukan")
	}
	if !actor.IsManager() && t.UserID != actor.UserID {
		return nil, apperror.NotFound("Transaksi tidak ditemukan")
	}
	return t, nil
}

func (s *TransactionService) List(ctx context.Context, actor Actor, f domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	f.Normalize()
	if !actor.IsManager() {
		f.UserID = &actor.UserID
	}
	return s.transactions.List(ctx, f)
}

func (s *TransactionService) Receipt(ctx context.Context, actor Actor, id int64) (*domain.Transaction, *domain.StoreSetting, string, error) {
	t, err := s.Get(ctx, actor, id)
	if err != nil {
		return nil, nil, "", err
	}
	st, err := s.settings.Get(ctx)
	if err != nil {
		return nil, nil, "", err
	}
	return t, st, BuildReceiptText(st, t, s.loc), nil
}

// FormatInvoiceNumber renders INV/YYYYMMDD/0001.
func FormatInvoiceNumber(day time.Time, seq int) string {
	return fmt.Sprintf("INV/%s/%04d", day.Format("20060102"), seq)
}
