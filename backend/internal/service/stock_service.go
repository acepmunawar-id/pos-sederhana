package service

import (
	"context"
	"fmt"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
)

type StockService struct {
	products  domain.ProductRepository
	movements domain.StockMovementRepository
	tx        domain.TxManager
}

func NewStockService(products domain.ProductRepository, movements domain.StockMovementRepository,
	tx domain.TxManager) *StockService {
	return &StockService{products: products, movements: movements, tx: tx}
}

// StockIn adds received goods to one or more products atomically.
func (s *StockService) StockIn(ctx context.Context, actor Actor, req dto.StockInRequest) ([]domain.StockMovement, error) {
	qty := map[int64]int{}
	var ids []int64
	for _, it := range req.Items {
		if _, seen := qty[it.ProductID]; !seen {
			ids = append(ids, it.ProductID)
		}
		qty[it.ProductID] += it.Quantity
	}
	note := trimPtr(req.Note)

	var out []domain.StockMovement
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		products, err := s.products.FindByIDsForUpdate(ctx, ids)
		if err != nil {
			return err
		}
		if err := ensureAllFound(ids, products); err != nil {
			return err
		}
		for _, p := range products {
			m, err := s.move(ctx, actor, p, domain.MovementIn, qty[p.ID], note)
			if err != nil {
				return err
			}
			out = append(out, *m)
		}
		return nil
	})
	return out, err
}

// Adjust sets a product's stock to the physically counted quantity and records the difference.
func (s *StockService) Adjust(ctx context.Context, actor Actor, req dto.StockAdjustmentRequest) (*domain.StockMovement, error) {
	var out *domain.StockMovement
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		products, err := s.products.FindByIDsForUpdate(ctx, []int64{req.ProductID})
		if err != nil {
			return err
		}
		if len(products) == 0 {
			return apperror.NotFound("Produk tidak ditemukan")
		}
		p := products[0]
		delta := req.ActualStock - p.Stock
		if delta == 0 {
			return apperror.BadRequest("Stok aktual sama dengan stok sistem, tidak ada yang perlu disesuaikan")
		}
		note := req.Note
		out, err = s.move(ctx, actor, p, domain.MovementAdjustment, delta, &note)
		return err
	})
	return out, err
}

func (s *StockService) Movements(ctx context.Context, f domain.StockMovementFilter) ([]domain.StockMovement, int64, error) {
	f.Normalize()
	return s.movements.List(ctx, f)
}

// LowStock lists active products at or below their minimum stock.
func (s *StockService) LowStock(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int64, error) {
	f.Normalize()
	active := true
	f.IsActive, f.LowStock = &active, true
	return s.products.List(ctx, f)
}

func (s *StockService) move(ctx context.Context, actor Actor, p domain.Product, typ string, delta int, note *string) (*domain.StockMovement, error) {
	after := p.Stock + delta
	if after < 0 {
		return nil, apperror.Unprocessable(fmt.Sprintf("Stok %s tidak mencukupi", p.Name))
	}
	if err := s.products.UpdateStock(ctx, p.ID, after); err != nil {
		return nil, err
	}
	m := &domain.StockMovement{
		ProductID: p.ID, ProductName: p.Name, ProductSKU: p.SKU, UserID: actor.UserID, Type: typ,
		Quantity: delta, StockBefore: p.Stock, StockAfter: after, Note: note,
	}
	if err := s.movements.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func ensureAllFound(ids []int64, products []domain.Product) error {
	if len(products) == len(ids) {
		return nil
	}
	found := make(map[int64]bool, len(products))
	for _, p := range products {
		found[p.ID] = true
	}
	for _, id := range ids {
		if !found[id] {
			return apperror.NotFound(fmt.Sprintf("Produk dengan ID %d tidak ditemukan", id))
		}
	}
	return nil
}
