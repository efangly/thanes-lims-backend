// Command db-sync-check is a standalone, read-only diagnostic that compares
// row counts (and Sample IDs) between the Postgres system of record and the
// chatbot POC's Oracle ADB. It proves empirically that the two databases are
// NOT synchronised: they hold independent datasets and a new insert into
// Postgres never reaches Oracle.
//
// Not wired into cmd/api - run manually like cmd/oracle-ping. On macOS you
// must export DYLD_LIBRARY_PATH to the Instant Client dir before running
// (see docs/chatbot-poc-plan.md, Phase 0).
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	oracledb "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/db"
	postgresdb "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	"github.com/efangly/thanes-lims-backend/internal/config"
)

var tables = []string{"samples", "test_results", "inventory_items", "purchase_orders"}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.OracleDSN == "" {
		log.Fatal("ORACLE_DSN is not set - cannot compare against Oracle")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- Postgres (system of record) ---
	gdb, err := postgresdb.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	pg, err := gdb.DB()
	if err != nil {
		log.Fatalf("postgres sql.DB: %v", err)
	}
	defer pg.Close()

	// --- Oracle ADB (chatbot POC) ---
	ora, err := oracledb.New(cfg.OracleDSN, cfg.OracleTNSAdmin)
	if err != nil {
		log.Fatalf("oracle: %v (on macOS: export DYLD_LIBRARY_PATH to the Instant Client dir)", err)
	}
	defer ora.Close()

	fmt.Println("=== Row count: Postgres vs Oracle ADB ===")
	fmt.Printf("%-18s %12s %12s   %s\n", "table", "postgres", "oracle", "match?")
	fmt.Println("-------------------------------------------------------------")
	for _, t := range tables {
		pgN, pgErr := count(ctx, pg, "SELECT COUNT(*) FROM "+t)
		orN, orErr := count(ctx, ora, "SELECT COUNT(*) FROM "+t)
		fmt.Printf("%-18s %12s %12s   %s\n", t, fmtN(pgN, pgErr), fmtN(orN, orErr), matchMark(pgN, orN, pgErr, orErr))
	}

	// Oracle-only junk rows left behind by cmd/oracle-insert-test.
	junk, junkErr := count(ctx, ora, "SELECT COUNT(*) FROM samples WHERE id LIKE 'SMP-TEST-%'")
	fmt.Printf("\noracle samples with id LIKE 'SMP-TEST-%%' (from cmd/oracle-insert-test): %s\n", fmtN(junk, junkErr))

	// --- Sample ID overlap ---
	pgIDs, err := ids(ctx, pg, "SELECT id FROM samples")
	if err != nil {
		log.Fatalf("postgres sample ids: %v", err)
	}
	orIDs, err := ids(ctx, ora, "SELECT id FROM samples")
	if err != nil {
		log.Fatalf("oracle sample ids: %v", err)
	}
	var overlap []string
	for id := range pgIDs {
		if orIDs[id] {
			overlap = append(overlap, id)
		}
	}
	fmt.Printf("\ndistinct sample ids: postgres=%d oracle=%d overlap=%d\n", len(pgIDs), len(orIDs), len(overlap))
	if len(overlap) > 0 {
		fmt.Printf("overlapping ids: %v\n", overlap)
	}
	fmt.Printf("sample id prefixes: postgres=%v oracle=%v\n", prefixes(pgIDs), prefixes(orIDs))

	fmt.Println("\n=== Conclusion ===")
	if len(overlap) == 0 {
		fmt.Println("No shared sample IDs and counts differ per table => the two databases are")
		fmt.Println("NOT synchronised. Oracle holds an independent synthetic seed; inserts via the")
		fmt.Println("API land in Postgres only.")
	} else {
		fmt.Println("Some sample IDs appear in both DBs - investigate whether a sync path exists.")
	}
}

func count(ctx context.Context, db *sql.DB, q string) (int64, error) {
	var n int64
	err := db.QueryRowContext(ctx, q).Scan(&n)
	return n, err
}

func ids(ctx context.Context, db *sql.DB, q string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func prefixes(set map[string]bool) []string {
	seen := make(map[string]bool)
	for id := range set {
		p := id
		if len(id) > 9 {
			p = id[:9] // e.g. "SMP-2569-"
		}
		seen[p] = true
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

func fmtN(n int64, err error) string {
	if err != nil {
		return "ERR"
	}
	return fmt.Sprintf("%d", n)
}

func matchMark(a, b int64, ae, be error) string {
	if ae != nil || be != nil {
		return "?"
	}
	if a == b {
		return "="
	}
	return "DIFFER"
}
