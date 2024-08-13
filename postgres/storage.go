// Package postgres gather all the code related to the postgres database
package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/perebaj/esaj"
)

// Storage is a struct that contains the database connection
type Storage struct {
	db *sqlx.DB
}

// NewStorage creates a new storage
func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
	}
}

// SaveProcessSeeds insert a list of process seeds into the database
func (s Storage) SaveProcessSeeds(ctx context.Context, ps []esaj.ProcessSeed) (int64, error) {
	result, err := s.db.NamedExecContext(ctx,
		`INSERT INTO process_seeds (process_id, oab, url) VALUES (:process_id, :oab, :url) ON CONFLICT DO NOTHING;`,
		ps,
	)
	if err != nil {
		return 0, fmt.Errorf("error inserting process_seeds: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected: %v", err)
	}

	return rows, nil
}
