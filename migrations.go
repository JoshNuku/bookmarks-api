package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// runMigrations applies all pending SQL migrations to the database.
func runMigrations(db *sql.DB) error {
	// 1. Wrap the embedded filesystem into a driver that golang-migrate understands
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	// 2. Wrap the existing *sql.DB connection pool into the postgres driver
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration database driver: %w", err)
	}

	// 3. Initialize the migrate instance with our embedded source and db connection
	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", dbDriver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migration instance: %w", err)
	}

	// 4. Run all pending "up" migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	return nil
}
