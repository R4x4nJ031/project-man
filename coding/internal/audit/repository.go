package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Entry struct {
	ID          string
	ActorUserID string
	TenantID    string
	Action      string
	Resource    string
	ResourceID  string
	CreatedAt   time.Time
}

func InsertTx(ctx context.Context, tx pgx.Tx, entry *Entry) error {
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}

	row := tx.QueryRow(
		ctx,
		`INSERT INTO audit_logs (
            id, actor_user_id, tenant_id, action, resource_type, resource_id
        ) VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING created_at`,
		entry.ID,
		entry.ActorUserID,
		entry.TenantID,
		entry.Action,
		entry.Resource,
		entry.ResourceID,
	)

	return row.Scan(&entry.CreatedAt)
}
