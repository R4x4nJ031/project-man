package project

import (
	"context"
	"fmt"

	"github.com/R4x4nJ031/project-man/internal/database"
	"github.com/jackc/pgx/v5"
)

func CreateProject(ctx context.Context, p *Project) error {
	return database.WithTenantTx(ctx, p.TenantID, func(tx pgx.Tx) error {
		return createProjectTx(ctx, tx, p)
	})
}

func ListProjects(ctx context.Context, tenantID string) ([]Project, error) {
	var projects []Project

	err := database.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(
			ctx,
			`SELECT id, tenant_id, name, created_at
         FROM projects
         WHERE tenant_id=$1`,
			tenantID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var p Project

			err := rows.Scan(
				&p.ID,
				&p.TenantID,
				&p.Name,
				&p.CreatedAt,
			)
			if err != nil {
				return err
			}

			projects = append(projects, p)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func GetProjectByID(ctx context.Context, projectID string, tenantID string) (*Project, error) {
	var p Project
	err := database.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		row := tx.QueryRow(
			ctx,
			`SELECT id, tenant_id, name, created_at
         FROM projects
         WHERE id=$1 AND tenant_id=$2`,
			projectID,
			tenantID,
		)

		return row.Scan(
			&p.ID,
			&p.TenantID,
			&p.Name,
			&p.CreatedAt,
		)
	})

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func UpdateProject(ctx context.Context, projectID string, tenantID string, name string) (*Project, error) {
	var p Project
	err := database.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		updated, err := updateProjectTx(ctx, tx, projectID, tenantID, name)
		if err != nil {
			return err
		}

		p = *updated
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func DeleteProject(ctx context.Context, projectID string, tenantID string) error {
	return database.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return deleteProjectTx(ctx, tx, projectID, tenantID)
	})
}

func createProjectTx(ctx context.Context, tx pgx.Tx, p *Project) error {
	row := tx.QueryRow(
		ctx,
		`INSERT INTO projects (id, tenant_id, name)
         VALUES ($1, $2, $3)
         RETURNING id, tenant_id, name, created_at`,
		p.ID,
		p.TenantID,
		p.Name,
	)

	return row.Scan(
		&p.ID,
		&p.TenantID,
		&p.Name,
		&p.CreatedAt,
	)
}

func updateProjectTx(ctx context.Context, tx pgx.Tx, projectID string, tenantID string, name string) (*Project, error) {
	row := tx.QueryRow(
		ctx,
		`UPDATE projects
         SET name=$1
         WHERE id=$2 AND tenant_id=$3
         RETURNING id, tenant_id, name, created_at`,
		name,
		projectID,
		tenantID,
	)

	var p Project
	err := row.Scan(
		&p.ID,
		&p.TenantID,
		&p.Name,
		&p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func deleteProjectTx(ctx context.Context, tx pgx.Tx, projectID string, tenantID string) error {
	result, err := tx.Exec(
		ctx,
		`DELETE FROM projects
         WHERE id=$1 AND tenant_id=$2`,
		projectID,
		tenantID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
