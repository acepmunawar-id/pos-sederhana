package service

import (
	"context"
	"errors"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
)

type ShiftService struct {
	shifts domain.ShiftRepository
	tx     domain.TxManager
	now    Clock
}

func NewShiftService(shifts domain.ShiftRepository, tx domain.TxManager, now Clock) *ShiftService {
	return &ShiftService{shifts: shifts, tx: tx, now: now}
}

func (s *ShiftService) Open(ctx context.Context, actor Actor, req dto.OpenShiftRequest) (*domain.Shift, error) {
	if _, err := s.shifts.FindOpenByUser(ctx, actor.UserID); err == nil {
		return nil, apperror.Conflict("Anda masih memiliki shift yang terbuka. Tutup shift tersebut terlebih dahulu")
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	sh := &domain.Shift{
		UserID: actor.UserID, Status: domain.ShiftOpen, OpeningCash: req.OpeningCash,
		OpeningNote: trimPtr(req.Note), OpenedAt: s.now(),
	}
	if err := s.shifts.Create(ctx, sh); err != nil {
		return nil, duplicate(err, "Anda masih memiliki shift yang terbuka")
	}
	return s.shifts.FindByID(ctx, sh.ID)
}

// Current returns the actor's open shift with a live summary.
func (s *ShiftService) Current(ctx context.Context, actor Actor) (*domain.Shift, *domain.ShiftSummary, error) {
	sh, err := s.shifts.FindOpenByUser(ctx, actor.UserID)
	if err != nil {
		return nil, nil, notFound(err, "Belum ada shift yang dibuka")
	}
	sum, err := s.shifts.Summary(ctx, sh.ID)
	if err != nil {
		return nil, nil, err
	}
	return sh, sum, nil
}

// Close finalises the actor's open shift: expected cash = opening cash + net cash sales,
// difference = counted cash − expected cash (negative means cash is missing).
func (s *ShiftService) Close(ctx context.Context, actor Actor, req dto.CloseShiftRequest) (*domain.Shift, *domain.ShiftSummary, error) {
	var sum *domain.ShiftSummary
	var closedID int64
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		sh, err := s.shifts.FindOpenByUser(ctx, actor.UserID)
		if err != nil {
			return notFound(err, "Belum ada shift yang dibuka")
		}
		sum, err = s.shifts.Summary(ctx, sh.ID)
		if err != nil {
			return err
		}
		expected := sh.OpeningCash + sum.CashSales
		diff := req.ActualCash - expected
		now := s.now()
		sh.Status = domain.ShiftClosed
		sh.ExpectedCash, sh.ActualCash, sh.CashDifference = &expected, &req.ActualCash, &diff
		sh.TotalSales, sh.TotalTransactions = &sum.TotalSales, &sum.TotalTransactions
		sh.ClosingNote, sh.ClosedAt = trimPtr(req.Note), &now
		closedID = sh.ID
		return notFound(s.shifts.Close(ctx, sh), "Shift sudah ditutup")
	})
	if err != nil {
		return nil, nil, err
	}
	sh, err := s.shifts.FindByID(ctx, closedID)
	return sh, sum, err
}

// Get returns a shift; cashiers may only see their own shifts.
func (s *ShiftService) Get(ctx context.Context, actor Actor, id int64) (*domain.Shift, *domain.ShiftSummary, error) {
	sh, err := s.shifts.FindByID(ctx, id)
	if err != nil {
		return nil, nil, notFound(err, "Shift tidak ditemukan")
	}
	if !actor.IsManager() && sh.UserID != actor.UserID {
		return nil, nil, apperror.NotFound("Shift tidak ditemukan")
	}
	sum, err := s.shifts.Summary(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return sh, sum, nil
}

func (s *ShiftService) List(ctx context.Context, actor Actor, f domain.ShiftFilter) ([]domain.Shift, int64, error) {
	f.Normalize()
	if !actor.IsManager() {
		f.UserID = &actor.UserID
	}
	return s.shifts.List(ctx, f)
}
