package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Otnielkalit/backend-pos/internal/infrastructure/config"
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver with database/sql
	"github.com/jmoiron/sqlx"
)


// NewPostgres creates and validates a new sqlx DB connection using the pgx driver.
// It applies connection pool settings from config and pings the database to confirm connectivity.
func NewPostgres(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	// pgx/stdlib registers "pgx" as a database/sql driver name
	db, err := sqlx.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return db, nil
}

// MustNewPostgres is like NewPostgres but panics on error.
// Use only in main.go during startup where a failed DB connection is unrecoverable.
func MustNewPostgres(cfg config.DatabaseConfig) *sqlx.DB {
	db, err := NewPostgres(cfg)
	if err != nil {
		panic(err)
	}
	return db
}

// The pgx stdlib driver is automatically registered as "pgx" with database/sql
// when this package is imported (via init() inside pgx/v5/stdlib itself).
// No explicit registration needed here.

