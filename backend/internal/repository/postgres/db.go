// Package postgres implements the domain repositories on PostgreSQL via pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/pkg/pagination"
)

// querier is satisfied by both *pgxpool.Pool and pgx.Tx.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type txKey struct{}

// conn returns the transaction stored in ctx, or the pool when none is active.
func conn(ctx context.Context, pool *pgxpool.Pool) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

type TxManager struct{ pool *pgxpool.Pool }

func NewTxManager(pool *pgxpool.Pool) *TxManager { return &TxManager{pool: pool} }

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx) // already inside a transaction: join it
	}
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return mapErr(fmt.Errorf("commit tx: %w", err))
	}
	return nil
}

// mapErr converts driver errors into domain errors.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", domain.ErrDuplicate, pgErr.ConstraintName)
	}
	return err
}

// where builds a parameterised WHERE clause. Use "?" as the placeholder in conditions.
type where struct {
	conds []string
	args  []any
}

func (w *where) add(cond string, args ...any) {
	for _, a := range args {
		w.args = append(w.args, a)
		cond = strings.Replace(cond, "?", fmt.Sprintf("$%d", len(w.args)), 1)
	}
	w.conds = append(w.conds, cond)
}

// search adds an ILIKE OR-group across the given columns.
func (w *where) search(term string, cols ...string) {
	if term == "" {
		return
	}
	w.args = append(w.args, "%"+escapeLike(term)+"%")
	ph := fmt.Sprintf("$%d", len(w.args))
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = c + " ILIKE " + ph
	}
	w.conds = append(w.conds, "("+strings.Join(parts, " OR ")+")")
}

func (w *where) sql() string {
	if len(w.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.conds, " AND ")
}

// page appends LIMIT/OFFSET placeholders and returns the clause plus full args.
func (w *where) page(p pagination.Params) (string, []any) {
	args := append(append([]any{}, w.args...), p.Limit, p.Offset())
	return fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// orderBy returns a safe ORDER BY clause: sort must be a key of allowed, otherwise fallback is used.
// idCol is appended as a tie-breaker so pagination is stable.
func orderBy(p pagination.Params, allowed map[string]string, fallback, idCol string) string {
	col, ok := allowed[p.Sort]
	if !ok {
		return " ORDER BY " + fallback + ", " + idCol + " DESC"
	}
	return fmt.Sprintf(" ORDER BY %s %s, %s DESC", col, strings.ToUpper(p.Order), idCol)
}

func count(ctx context.Context, q querier, sql string, args []any) (int64, error) {
	var n int64
	if err := q.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func exists(ctx context.Context, q querier, sql string, args ...any) (bool, error) {
	var ok bool
	err := q.QueryRow(ctx, "SELECT EXISTS("+sql+")", args...).Scan(&ok)
	return ok, err
}

func affectedOrNotFound(tag pgconn.CommandTag, err error) error {
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
