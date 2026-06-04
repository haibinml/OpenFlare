package goose

import (
	"fmt"

	presslygoose "github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

const versionPagesDeploymentRootDir int64 = 202606040001

// migration202606040001 adds RootDir field to Pages deployments.
func migration202606040001(backend string, ctx Context) *presslygoose.Migration {
	return newGORMMigration(
		versionPagesDeploymentRootDir,
		"202606040001_add_pages_deployment_root_dir.go",
		backend,
		ctx,
		migratePagesDeploymentRootDir,
	)
}

func migratePagesDeploymentRootDir(ctx Context, db *gorm.DB, backend string) error {
	if err := ctx.ApplyCurrentSchema(db, backend); err != nil {
		return err
	}
	// Verify that the column exists
	if !db.Migrator().HasColumn("pages_deployments", "root_dir") {
		return fmt.Errorf("column pages_deployments.root_dir is missing")
	}
	return nil
}
