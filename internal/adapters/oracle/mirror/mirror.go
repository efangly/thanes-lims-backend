// Package mirror keeps the chatbot POC's Oracle ADB in sync with the Postgres
// system of record for the four entities the chatbot queries: samples,
// test_results, inventory_items and purchase_orders
// (scripts/oracle/001_schema.sql).
//
// It works as a set of repository decorators wired in at the composition root
// (cmd/api/routes.go): every write that succeeds against Postgres is then
// best-effort MERGE'd into Oracle. A mirror failure is logged and swallowed -
// the main API must stay fully functional on Postgres alone even when the ADB
// is unreachable. cmd/oracle-backfill reconciles anything missed.
//
// The mirror tables are a lossy projection of the domain: they have no Users
// or Locations tables (custodian/location are free text), no inventory lots
// (an item's quantity is the pre-summed total), and no barcode/description/
// position columns. The Upsert* methods encode that mapping.
package mirror

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// mirrorTimeout caps a single mirror round-trip. It is deliberately short:
// mirroring must never noticeably slow the Postgres-backed request it trails.
const mirrorTimeout = 10 * time.Second

// Mirror holds the writable Oracle ADB connection (the CHATBOT_APP user -
// distinct from the chatbot's read-only CHATBOT_RO connection) and the
// per-table MERGE statements.
type Mirror struct {
	db *sql.DB
}

// New wraps a writable Oracle *sql.DB (see internal/adapters/oracle/db.New).
func New(db *sql.DB) *Mirror {
	return &Mirror{db: db}
}

// run executes fn against Oracle with its own short, cancel-detached timeout
// (the trailing request's context may already be done) and logs any failure
// without propagating it.
func run(ctx context.Context, entity, id string, fn func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), mirrorTimeout)
	defer cancel()
	if err := fn(ctx); err != nil {
		log.Printf("oracle mirror: %s %s: %v", entity, id, err)
	}
}

// nullText maps an empty optional string to a SQL NULL so it does not trip a
// CHECK constraint (e.g. test_results.flag) or store a meaningless "".
func nullText(s string) any {
	if s == "" {
		return nil
	}
	return s
}
