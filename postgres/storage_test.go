//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/perebaj/esaj"
	"github.com/perebaj/esaj/postgres"
	"github.com/stretchr/testify/require"
)

// OpenDB create a new database for testing and return a connection to it.
// Why: For testing, we need a new database for each test to avoid side effects.
// So Opendb creates a new database with a random suffix, and after the test, it drops the database.
func OpenDB(t *testing.T) *sqlx.DB {
	t.Helper()

	cfg := postgres.Config{
		URL:             os.Getenv("POSTGRES_URL"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := sql.Open("postgres", cfg.URL)
	require.NoError(t, err, "error connecting to Postgres: %v", err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	require.NoError(t, err, "error pinging Postgres: %v", err)

	// create a new database with random suffix
	postgresURL, err := url.Parse(cfg.URL)
	require.NoError(t, err, "error parsing Postgres connection URL: %v", err)

	database := strings.TrimLeft(postgresURL.Path, "/")

	randSuffix := fmt.Sprintf("%x", time.Now().UnixNano())

	database = fmt.Sprintf("%s-%x", database, randSuffix)
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, database))
	require.NoError(t, err, "error creating database for test: %v", err)

	postgresURL.Path = "/" + database
	cfg.URL = postgresURL.String()
	testDB, err := postgres.OpenDB(cfg)
	require.NoError(t, err, "error opening test database: %v", err)

	// after run the tests, drop the database
	t.Cleanup(func() {
		defer func() {
			_ = testDB.Close()
		}()

		defer func() {
			_ = db.Close()
		}()
		_, err = db.Exec(fmt.Sprintf(`DROP DATABASE "%s" WITH (FORCE);`, database))
		require.NoError(t, err, "error dropping database for test: %v", err)
	})

	return testDB
}

// Test a ping to the database
func TestPing(t *testing.T) {
	db := OpenDB(t)
	defer func() {
		_ = db.Close()
	}()

	err := db.Ping()
	require.NoError(t, err)
}

// TestSaveProcessesSeed test the SaveProcessesSeed method
func TestSaveProcessesSeed(t *testing.T) {
	db := OpenDB(t)
	defer func() {
		_ = db.Close()
	}()

	storage := postgres.NewStorage(db)

	ps := []esaj.ProcessSeed{
		{
			ProcessID: "1",
			OAB:       "OAB",
			URL:       "URL",
		},
		{
			ProcessID: "2",
			OAB:       "OAB",
			URL:       "URL",
		},
	}

	rows, err := storage.SaveProcessSeeds(context.Background(), ps)
	require.NoError(t, err)
	require.Equal(t, int64(2), rows)

	// valide if on conflict do nothing is working.
	ps = []esaj.ProcessSeed{
		{
			ProcessID: "1",
			OAB:       "OAB",
			URL:       "URL",
		},
		{
			ProcessID: "2",
			OAB:       "OAB2",
			URL:       "URL",
		},
	}

	rows, err = storage.SaveProcessSeeds(context.Background(), ps)
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
}
