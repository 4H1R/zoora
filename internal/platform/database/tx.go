package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromCtx retrieves the transactional *gorm.DB stashed by Transactor.RunInTx, if any.
func TxFromCtx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok
}

// DB returns the tx-scoped DB if running inside a transaction, else fallback.
// Always applies WithContext.
func DB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := TxFromCtx(ctx); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

// Transactor runs functions inside a DB transaction. The tx is propagated via ctx.
type Transactor struct {
	db *gorm.DB
}

func NewTransactor(db *gorm.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := TxFromCtx(ctx); ok {
		// Nested — just reuse existing tx.
		return fn(ctx)
	}
	return t.db.Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}
