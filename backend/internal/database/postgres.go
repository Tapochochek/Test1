package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"orenkod/backend/migrations"
)

func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	waitCtx, cancel := context.WithTimeout(ctx, 70*time.Second)
	defer cancel()
	for {
		if err = db.PingContext(waitCtx); err == nil {
			break
		}
		if waitCtx.Err() != nil {
			_ = db.Close()
			return nil, fmt.Errorf("can't connect to PostgreSQL: %w", err)
		}
		time.Sleep(2 * time.Second)
	}

	if err = ApplySchema(waitCtx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err = Seed(waitCtx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func ApplySchema(ctx context.Context, db *sql.DB) error {
	schema, err := migrations.Files.ReadFile("001_init.sql")
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
