// Package service holds all business logic. Services depend only on domain interfaces.
package service

import (
	"errors"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/pkg/apperror"
)

// Actor is the authenticated user performing an action.
type Actor struct {
	UserID int64
	Role   string
}

func (a Actor) IsManager() bool { return a.Role == domain.RoleOwner || a.Role == domain.RoleAdmin }

// Clock lets tests control "now".
type Clock func() time.Time

// notFound maps domain.ErrNotFound to a 404 with the given message, passing other errors through.
func notFound(err error, msg string) error {
	if errors.Is(err, domain.ErrNotFound) {
		return apperror.NotFound(msg)
	}
	return err
}

// duplicate maps domain.ErrDuplicate (a lost race on a unique index) to a 409.
func duplicate(err error, msg string) error {
	if errors.Is(err, domain.ErrDuplicate) {
		return apperror.Conflict(msg)
	}
	return err
}

// mulBP returns round-half-up(amount × bp / 10000) using integer arithmetic only.
func mulBP(amount int64, bp int64) int64 {
	return (amount*bp + domain.BasisPoints/2) / domain.BasisPoints
}

// startOfDay returns midnight of t's date in loc.
func startOfDay(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}
