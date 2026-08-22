// Command oracle-ping is a standalone connectivity smoke test for the
// Select AI chatbot POC's Oracle ADB instance (godror + wallet). It is not
// wired into cmd/api - the chatbot module is a separate, optional feature
// on top of a separate database from the Postgres system of record.
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/config"
	_ "github.com/godror/godror"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.OracleDSN == "" {
		log.Fatal("ORACLE_DSN is not set")
	}
	if cfg.OracleTNSAdmin != "" {
		if err := os.Setenv("TNS_ADMIN", cfg.OracleTNSAdmin); err != nil {
			log.Fatalf("setenv TNS_ADMIN: %v", err)
		}
	}

	db, err := sql.Open("godror", cfg.OracleDSN)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dbName, host string
	query := `SELECT SYS_CONTEXT('USERENV','DB_NAME'), SYS_CONTEXT('USERENV','SERVER_HOST') FROM dual`
	if err := db.QueryRowContext(ctx, query).Scan(&dbName, &host); err != nil {
		log.Fatalf("query: %v", err)
	}

	log.Printf("connected OK: db_name=%s host=%s", dbName, host)
}
