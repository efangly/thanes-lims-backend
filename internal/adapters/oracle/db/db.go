package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/godror/godror"
)

// New opens a connection to the Oracle ADB (chatbot POC instance) via godror
// and tunes the underlying connection pool. Unlike the Postgres system of
// record, this is a low-traffic optional path, so the pool is kept small.
func New(dsn, tnsAdmin string) (*sql.DB, error) {
	if tnsAdmin != "" {
		if err := os.Setenv("TNS_ADMIN", tnsAdmin); err != nil {
			return nil, fmt.Errorf("oracle db: setenv TNS_ADMIN: %w", err)
		}
	}

	sdb, err := sql.Open("godror", dsn)
	if err != nil {
		return nil, fmt.Errorf("oracle db: open: %w", err)
	}

	sdb.SetMaxOpenConns(5)
	sdb.SetMaxIdleConns(2)
	sdb.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sdb.PingContext(ctx); err != nil {
		sdb.Close()
		return nil, fmt.Errorf("oracle db: ping: %w", err)
	}

	return sdb, nil
}
