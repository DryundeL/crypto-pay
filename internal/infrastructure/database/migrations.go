package database

import (
	"fmt"
	"log"

	"github.com/DryundeL/crypto-pay/internal/infrastructure/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrationsUp applies all pending database migrations
// This function uses the same logic as cmd/migrator/main.go
func RunMigrationsUp(cfg *config.Config, migrationsPath string) error {
	m, err := migrate.New(
		"file://"+migrationsPath,
		cfg.Database.DSN(),
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	log.Printf("Connected to database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	currentVersion, _, _ := m.Version()
	if currentVersion > 0 {
		log.Printf("Current database migration version: %d", currentVersion)
	} else {
		log.Println("No migrations have been applied yet")
	}

	log.Println("Checking for pending migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	if err == migrate.ErrNoChange {
		log.Println("✓ No pending migrations - database is up to date")
		return nil
	}

	newVersion, dirty, _ := m.Version()
	if dirty {
		log.Printf("⚠ Warning: Migration version %d is marked as dirty", newVersion)
		return fmt.Errorf("migration version %d is marked as dirty", newVersion)
	}

	log.Printf("✓ Migrations applied successfully. Current version: %d", newVersion)
	return nil
}
