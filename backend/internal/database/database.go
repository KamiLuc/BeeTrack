package database

import (
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(url string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	return db, nil
}

func Migrate(db *gorm.DB, migrations fs.FS) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
