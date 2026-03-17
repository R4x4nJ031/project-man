package project

import (
	"context"

	"github.com/R4x4nJ031/project-man/internal/audit"
	"github.com/R4x4nJ031/project-man/internal/database"
	"github.com/jackc/pgx/v5"
)

func CreateProjectWithAudit(ctx context.Context, p *Project, actorUserID string) error {
	return database.WithTenantTx(ctx, p.TenantID, func(tx pgx.Tx) error {
		if err := createProjectTx(ctx, tx, p); err != nil {
			return err
		}

		return audit.InsertTx(ctx, tx, &audit.Entry{
			ActorUserID: actorUserID,
			TenantID:    p.TenantID,
			Action:      "project.create",
			Resource:    "project",
			ResourceID:  p.ID,
		})
	})
}

func UpdateProjectWithAudit(ctx context.Context, projectID string, tenantID string, name string, actorUserID string) (*Project, error) {
	var p Project

	err := database.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		updated, err := updateProjectTx(ctx, tx, projectID, tenantID, name)
		if err != nil {
			return err
		}
		p = *updated

		return audit.InsertTx(ctx, tx, &audit.Entry{
			ActorUserID: actorUserID,
			TenantID:    tenantID,
			Action:      "project.update",
			Resource:    "project",
			ResourceID:  projectID,
		})
	})
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func DeleteProjectWithAudit(ctx context.Context, projectID string, tenantID string, actorUserID string) error {
	return database.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if err := deleteProjectTx(ctx, tx, projectID, tenantID); err != nil {
			return err
		}

		return audit.InsertTx(ctx, tx, &audit.Entry{
			ActorUserID: actorUserID,
			TenantID:    tenantID,
			Action:      "project.delete",
			Resource:    "project",
			ResourceID:  projectID,
		})
	})
}
