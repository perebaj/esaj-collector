package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

// Config have the core configuration for postgres database.
type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
}

// OpenDB create a database connection
func OpenDB(cfg Config) (*sqlx.DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %v", err)
	}

	err = Migrate(db)
	if err != nil {
		return nil, fmt.Errorf("migrating postgres: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping postgres: %v", err)
	}
	return sqlx.NewDb(db, "postgres"), nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate run the migrations for the database.
func Migrate(db *sql.DB) error {
	fs, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating iofs driver: %v", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %v", err)
	}
	m, err := migrate.NewWithInstance("iofs", fs, "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
