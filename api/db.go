package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectDB opens a connection pool to Postgres, checks it works and creates
// the tables the API needs. The caller owns the pool and closes it on shutdown.
func connectDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// migrate creates the schema if it doesn't exist yet. Enough for one table;
// a bigger app would use a migration tool.
func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id         BIGSERIAL PRIMARY KEY,
			text       TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	return err
}
