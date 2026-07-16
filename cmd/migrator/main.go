package main

import (
	"flag"
	"log"

	"github.com/DryundeL/crypto-pay/internal/infrastructure/config"
	"github.com/DryundeL/crypto-pay/internal/infrastructure/database"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		command = flag.String("command", "up", "migration command: up, down, force, version")
		steps   = flag.Int("steps", 0, "number of migration steps (0 = all)")
		version = flag.Int("version", 0, "target version for force command")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Migrations directory is located in the project root
	migrationsPath := "migrations"

	// Create migrate instance
	// golang-migrate automatically creates and manages schema_migrations table
	m, err := migrate.New(
		"file://"+migrationsPath,
		cfg.Database.DSN(),
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	log.Printf("Connected to database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Check current version before migration
	currentVersion, _, _ := m.Version()
	if currentVersion > 0 {
		log.Printf("Current database migration version: %d", currentVersion)
	} else {
		log.Println("No migrations have been applied yet")
	}

	// Execute migration command
	switch *command {
	case "up":
		if *steps > 0 {
			log.Printf("Applying %d migration(s)...", *steps)
			err = m.Steps(*steps)
			if err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Failed to run migrations up: %v", err)
			}
			if err == migrate.ErrNoChange {
				log.Println("✓ No pending migrations - database is up to date")
			} else {
				newVersion, dirty, _ := m.Version()
				if dirty {
					log.Printf("⚠ Warning: Migration version %d is marked as dirty", newVersion)
				} else {
					log.Printf("✓ Migrations applied successfully. Current version: %d", newVersion)
				}
			}
		} else {
			// Use the shared function for applying all migrations
			if err := database.RunMigrationsUp(cfg, migrationsPath); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
		}

	case "down":
		if *steps > 0 {
			log.Printf("Rolling back %d migration(s)...", *steps)
			err = m.Steps(-*steps)
		} else {
			log.Println("Rolling back last migration...")
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
		if err == migrate.ErrNoChange {
			log.Println("✓ No migrations to rollback")
		} else {
			newVersion, dirty, _ := m.Version()
			if dirty {
				log.Printf("⚠ Warning: Migration version %d is marked as dirty", newVersion)
			} else {
				log.Printf("✓ Migrations rolled back successfully. Current version: %d", newVersion)
			}
		}

	case "force":
		if *version == 0 {
			log.Fatal("Version is required for force command. Use -version flag")
		}
		err = m.Force(*version)
		if err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
		log.Printf("Migration version forced to %d", *version)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Println("No migrations have been applied yet")
				log.Println("The schema_migrations table will be created automatically on first migration")
			} else {
				log.Fatalf("Failed to get migration version: %v", err)
			}
		} else {
			if dirty {
				log.Printf("⚠ Current migration version: %d (DIRTY - migration may have failed)", version)
				log.Println("Use -command=force -version=<version> to mark as clean if needed")
			} else {
				log.Printf("✓ Current migration version: %d", version)
				log.Println("All migrations up to this version have been applied successfully")
			}
		}

	default:
		log.Fatalf("Unknown command: %s. Use: up, down, force, version", *command)
	}
}
