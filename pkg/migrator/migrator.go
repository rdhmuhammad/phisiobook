package migrator

import (
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	"github.com/rdhmuhammad/phisiobook/resource/db-changelog"
)

func Up(db *gorm.DB) error {
	goose.SetBaseFS(migrations.FS)

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	return goose.Up(sqlDB, ".")
}
