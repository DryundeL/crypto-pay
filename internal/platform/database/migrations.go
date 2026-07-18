package database

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
)

// LogIdentity is used only for safe startup logs (no password).
type LogIdentity struct {
	User   string
	Host   string
	Port   string
	DBName string
}

// RunMigrationsUp applies all pending database migrations.
func RunMigrationsUp(dsn string, identity LogIdentity, migrationsPath string) error {
	m, err := newMigrate(migrationsPath, dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	log.Printf("Connected to database: %s@%s:%s/%s", identity.User, identity.Host, identity.Port, identity.DBName)

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
