// Command oracle-backfill is a one-time reconciler that pushes the current
// contents of the Postgres system of record into the chatbot POC's Oracle
// ADB mirror (samples, test_results, inventory_items, purchase_orders). Run
// it once after enabling the dual-write mirror, and again any time Oracle was
// unreachable for a stretch and rows were missed.
//
// Standalone, run manually like cmd/oracle-ping. On macOS export
// DYLD_LIBRARY_PATH to the Instant Client dir first (see docs/chatbot-poc-plan.md).
//
//	export DYLD_LIBRARY_PATH=/Users/tng-mac-01/oracle/instantclient/instantclient_23_26
//	go run ./cmd/oracle-backfill
package main

import (
	"context"
	"log"
	"time"

	oracledb "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/db"
	oraclemirror "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/mirror"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	postgresinventory "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/inventory"
	postgreslocation "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/location"
	postgrespurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/purchaseorder"
	postgressample "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	postgrestestresult "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/testresult"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	"github.com/efangly/thanes-lims-backend/internal/config"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.OracleDSN == "" {
		log.Fatal("ORACLE_DSN is not set")
	}

	gdb, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	oradb, err := oracledb.New(cfg.OracleDSN, cfg.OracleTNSAdmin)
	if err != nil {
		log.Fatalf("oracle: %v (on macOS export DYLD_LIBRARY_PATH to the Instant Client dir)", err)
	}
	defer oradb.Close()

	m := oraclemirror.New(oradb)
	userRepo := postgresuser.New(gdb)
	locationRepo := postgreslocation.New(gdb)
	sampleRepo := postgressample.New(gdb)
	testResultRepo := postgrestestresult.New(gdb)
	inventoryRepo := postgresinventory.New(gdb)
	purchaseOrderRepo := postgrespurchaseorder.New(gdb)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Clear rows left behind by cmd/oracle-insert-test (test_results first -
	// FK to samples).
	for _, stmt := range []string{
		`DELETE FROM test_results WHERE sample_id LIKE 'SMP-TEST-%'`,
		`DELETE FROM samples WHERE id LIKE 'SMP-TEST-%'`,
	} {
		res, err := oradb.ExecContext(ctx, stmt)
		if err != nil {
			log.Fatalf("clean junk (%s): %v", stmt, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			log.Printf("removed %d junk row(s): %s", n, stmt)
		}
	}

	// Order matters for the mirror's FKs: inventory_items before
	// purchase_orders, samples before test_results.
	items, err := inventoryRepo.List(ctx)
	if err != nil {
		log.Fatalf("list inventory: %v", err)
	}
	for _, it := range items {
		if err := m.UpsertInventoryItem(ctx, it); err != nil {
			log.Fatalf("upsert inventory_item %s: %v", it.ID, err)
		}
	}
	log.Printf("inventory_items: %d upserted", len(items))

	pos, err := purchaseOrderRepo.List(ctx)
	if err != nil {
		log.Fatalf("list purchase orders: %v", err)
	}
	for _, po := range pos {
		if err := m.UpsertPurchaseOrder(ctx, po); err != nil {
			log.Fatalf("upsert purchase_order %s: %v", po.ID, err)
		}
	}
	log.Printf("purchase_orders: %d upserted", len(pos))

	samples, err := sampleRepo.List(ctx, portsample.ListFilter{})
	if err != nil {
		log.Fatalf("list samples: %v", err)
	}
	for _, s := range samples {
		custodian := "-"
		if u, err := userRepo.FindByID(ctx, s.CustodianUserID); err == nil && u.Name != "" {
			custodian = u.Name
		}
		location := "-"
		if s.LocationID != nil {
			if p, err := locationRepo.FullPath(ctx, *s.LocationID); err == nil && p != "" {
				location = p
			}
		}
		if err := m.UpsertSample(ctx, s, custodian, location); err != nil {
			log.Fatalf("upsert sample %s: %v", s.ID, err)
		}
	}
	log.Printf("samples: %d upserted", len(samples))

	results, err := testResultRepo.List(ctx, porttestresult.ListFilter{})
	if err != nil {
		log.Fatalf("list test results: %v", err)
	}
	for _, tr := range results {
		if err := m.UpsertTestResult(ctx, tr); err != nil {
			log.Fatalf("upsert test_result %s: %v", tr.ID, err)
		}
	}
	log.Printf("test_results: %d upserted", len(results))

	log.Println("backfill complete - run ./cmd/db-sync-check to verify")
}
