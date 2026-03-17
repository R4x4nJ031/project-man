package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func WithTenantTx(ctx context.Context, tenantID string, fn func(pgx.Tx) error) error {
	tx, err := DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// SET LOCAL scopes the tenant context to this transaction only,
	// which prevents leakage across pooled connections.
	_, err = tx.Exec(ctx, `SELECT set_config('app.current_tenant', $1, true)`, tenantID)
	if err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
