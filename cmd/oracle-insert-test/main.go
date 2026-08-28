// Command oracle-insert-test is a standalone smoke test for the Oracle ADB
// samples mirror adapter (internal/adapters/oracle/sample). It inserts one
// synthetic sample row and reads it back to prove the insert path works
// end-to-end against the real ADB. Not wired into cmd/api or go test - run
// manually, the same way cmd/oracle-ping is.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	oracledb "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/db"
	oraclesampleadapter "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/sample"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/ports/oraclesample"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.OracleDSN == "" {
		log.Fatal("ORACLE_DSN is not set")
	}

	sdb, err := oracledb.New(cfg.OracleDSN, cfg.OracleTNSAdmin)
	if err != nil {
		log.Fatalf("oracle db: %v", err)
	}
	defer sdb.Close()

	repo := oraclesampleadapter.New(sdb)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	locationID := "N/A"
	want := oraclesample.MirrorSample{
		ID:         fmt.Sprintf("SMP-TEST-%d", time.Now().Unix()),
		Name:       "oracle-insert-test smoke test",
		Type:       string(sample.TypeBlood),
		Custodian:  "cmd/oracle-insert-test",
		LocationID: &locationID,
		Status:     string(sample.StatusPending),
		ReceivedAt: time.Now().Truncate(time.Second),
	}

	if err := repo.Insert(ctx, want); err != nil {
		log.Fatalf("insert: %v", err)
	}
	log.Printf("inserted sample id=%s", want.ID)

	got, err := repo.FindByID(ctx, want.ID)
	if err != nil {
		log.Fatalf("find by id: %v", err)
	}

	locationMismatch := (got.LocationID == nil) != (want.LocationID == nil) ||
		(got.LocationID != nil && *got.LocationID != *want.LocationID)
	if got.ID != want.ID || got.Name != want.Name || got.Type != want.Type ||
		got.Custodian != want.Custodian || locationMismatch ||
		got.Status != want.Status || !got.ReceivedAt.Equal(want.ReceivedAt) {
		log.Fatalf("read-back mismatch: want %+v, got %+v", want, got)
	}

	log.Printf("OK: insert + read-back matched for id=%s", want.ID)
}
