// Package postgres gather all the code related to the postgres database
package postgres

import "github.com/jmoiron/sqlx"

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
