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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	c, err := oraclemirror.Backfill(ctx, oraclemirror.New(oradb), oraclemirror.Source{
		Users:       postgresuser.New(gdb),
		Locations:   postgreslocation.New(gdb),
		Samples:     postgressample.New(gdb),
		TestResults: postgrestestresult.New(gdb),
		Inventory:   postgresinventory.New(gdb),
		POs:         postgrespurchaseorder.New(gdb),
	})
	if err != nil {
		log.Fatalf("backfill: %v", err)
	}

	log.Printf("backfill complete: inventory_items=%d purchase_orders=%d samples=%d test_results=%d",
		c.InventoryItems, c.PurchaseOrders, c.Samples, c.TestResults)
	log.Println("run ./cmd/db-sync-check to verify")
}
