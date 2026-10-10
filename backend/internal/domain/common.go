// Package domain contains the business entities and repository contracts.
// It has no dependency on any database or HTTP library.
package domain

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned by repositories when a row does not exist.
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate is returned by repositories when a unique constraint is violated.
	ErrDuplicate = errors.New("duplicate record")
)

// TxManager runs fn inside a single database transaction. Repositories called with
// the ctx passed to fn automatically participate in that transaction.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Role codes.
const (
	RoleOwner   = "owner"
	RoleAdmin   = "admin"
	RoleCashier = "cashier"
)

// BasisPoints is the denominator for percentage values stored as integers (1100 = 11%).
const BasisPoints = 10000
