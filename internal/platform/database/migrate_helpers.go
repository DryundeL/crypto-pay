package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func newMigrate(migrationsPath, dsn string) (*migrate.Migrate, error) {
	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	return m, nil
}
