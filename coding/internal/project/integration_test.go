//go:build integration

package project

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/R4x4nJ031/project-man/internal/database"
	"github.com/jackc/pgx/v5"
)

const integrationDatabaseURL = "postgres://project:projectpass@localhost:5432/projectdb"

func TestRLSBlocksCrossTenantDirectRead(t *testing.T) {
	ctx := integrationTestContext(t)

	project := &Project{
		ID:       "project-rls-check",
		TenantID: "tenant-acme",
		Name:     "rls-check",
	}

	if err := CreateProjectWithAudit(ctx, project, "123"); err != nil {
		t.Fatalf("create project with audit: %v", err)
	}

	var foundID string
	err := database.WithTenantTx(ctx, "tenant-beta", func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id FROM projects WHERE id=$1`, project.ID).Scan(&foundID)
	})

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows due to RLS, got %v", err)
	}
}

func TestAuditLoggingCapturesProjectMutations(t *testing.T) {
	ctx := integrationTestContext(t)

	project := &Project{
		ID:       "project-audit-check",
		TenantID: "tenant-acme",
		Name:     "audit-check",
	}

	if err := CreateProjectWithAudit(ctx, project, "123"); err != nil {
		t.Fatalf("create project with audit: %v", err)
	}

	if _, err := UpdateProjectWithAudit(ctx, project.ID, project.TenantID, "audit-check-updated", "123"); err != nil {
		t.Fatalf("update project with audit: %v", err)
	}

	if err := DeleteProjectWithAudit(ctx, project.ID, project.TenantID, "123"); err != nil {
		t.Fatalf("delete project with audit: %v", err)
	}

	rows, err := database.DB.Query(ctx, `
        SELECT action, tenant_id, actor_user_id, resource_type, resource_id
        FROM audit_logs
        WHERE resource_id=$1
        ORDER BY created_at
    `, project.ID)
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	defer rows.Close()

	var actions []string
	for rows.Next() {
		var action string
		var tenantID string
		var actorUserID string
		var resourceType string
		var resourceID string

		if err := rows.Scan(&action, &tenantID, &actorUserID, &resourceType, &resourceID); err != nil {
			t.Fatalf("scan audit log row: %v", err)
		}

		if tenantID != "tenant-acme" {
			t.Fatalf("expected tenant-acme, got %q", tenantID)
		}

		if actorUserID != "123" {
			t.Fatalf("expected actor 123, got %q", actorUserID)
		}

		if resourceType != "project" {
			t.Fatalf("expected resource type project, got %q", resourceType)
		}

		if resourceID != project.ID {
			t.Fatalf("expected resource id %q, got %q", project.ID, resourceID)
		}

		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	expected := []string{"project.create", "project.update", "project.delete"}
	if len(actions) != len(expected) {
		t.Fatalf("expected %d audit rows, got %d (%v)", len(expected), len(actions), actions)
	}

	for i := range expected {
		if actions[i] != expected[i] {
			t.Fatalf("expected action %q at index %d, got %q", expected[i], i, actions[i])
		}
	}
}

func integrationTestContext(t *testing.T) context.Context {
	t.Helper()

	if database.DB == nil {
		dbURL := os.Getenv("INTEGRATION_DATABASE_URL")
		if dbURL == "" {
			dbURL = integrationDatabaseURL
		}

		if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
			t.Fatalf("set DATABASE_URL: %v", err)
		}

		if err := database.Connect(); err != nil {
			t.Fatalf("connect database: %v", err)
		}
	}

	ctx := context.Background()

	// Keep the integration tests repeatable even when rerun against the same DB.
	_, err := database.DB.Exec(ctx, `DELETE FROM audit_logs WHERE resource_id IN ('project-rls-check', 'project-audit-check')`)
	if err != nil {
		t.Fatalf("reset audit logs: %v", err)
	}

	_, err = database.DB.Exec(ctx, `DELETE FROM projects WHERE id IN ('project-rls-check', 'project-audit-check')`)
	if err != nil {
		t.Fatalf("reset projects: %v", err)
	}

	return ctx
}
