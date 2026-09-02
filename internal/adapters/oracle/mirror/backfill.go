package mirror

import (
	"context"
	"fmt"

	dsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	dtestresult "github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portpo "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

// Source bundles the read side of the Postgres system of record needed to
// (re)build the whole mirror in one pass. The Postgres repositories satisfy
// it directly.
type Source struct {
	Users       UserDirectory
	Locations   LocationPathResolver
	Samples     interface {
		List(context.Context, portsample.ListFilter) ([]dsample.Sample, error)
	}
	TestResults interface {
		List(context.Context, porttestresult.ListFilter) ([]dtestresult.TestResult, error)
	}
	Inventory portinventory.Repository
	POs       portpo.Repository
}

// Counts reports how many rows were upserted per table.
type Counts struct {
	InventoryItems, PurchaseOrders, Samples, TestResults int
}

// Backfill pushes every current row from src into the mirror, in the order
// the mirror's foreign keys require (inventory_items → purchase_orders,
// samples → test_results). It first removes rows left behind by
// cmd/oracle-insert-test. Unlike the per-write decorators this is not
// best-effort: any error aborts and is returned.
func Backfill(ctx context.Context, m *Mirror, src Source) (Counts, error) {
	var c Counts

	for _, stmt := range []string{
		`DELETE FROM test_results WHERE sample_id LIKE 'SMP-TEST-%'`,
		`DELETE FROM samples WHERE id LIKE 'SMP-TEST-%'`,
	} {
		if _, err := m.db.ExecContext(ctx, stmt); err != nil {
			return c, fmt.Errorf("clean junk (%s): %w", stmt, err)
		}
	}

	items, err := src.Inventory.List(ctx)
	if err != nil {
		return c, fmt.Errorf("list inventory: %w", err)
	}
	for _, it := range items {
		if err := m.UpsertInventoryItem(ctx, it); err != nil {
			return c, fmt.Errorf("upsert inventory_item %s: %w", it.ID, err)
		}
	}
	c.InventoryItems = len(items)

	pos, err := src.POs.List(ctx)
	if err != nil {
		return c, fmt.Errorf("list purchase orders: %w", err)
	}
	for _, po := range pos {
		if err := m.UpsertPurchaseOrder(ctx, po); err != nil {
			return c, fmt.Errorf("upsert purchase_order %s: %w", po.ID, err)
		}
	}
	c.PurchaseOrders = len(pos)

	samples, err := src.Samples.List(ctx, portsample.ListFilter{})
	if err != nil {
		return c, fmt.Errorf("list samples: %w", err)
	}
	for _, s := range samples {
		custodian := unresolved
		if u, err := src.Users.FindByID(ctx, s.CustodianUserID); err == nil && u.Name != "" {
			custodian = u.Name
		}
		location := unresolved
		if s.LocationID != nil {
			if p, err := src.Locations.FullPath(ctx, *s.LocationID); err == nil && p != "" {
				location = p
			}
		}
		if err := m.UpsertSample(ctx, s, custodian, location); err != nil {
			return c, fmt.Errorf("upsert sample %s: %w", s.ID, err)
		}
	}
	c.Samples = len(samples)

	results, err := src.TestResults.List(ctx, porttestresult.ListFilter{})
	if err != nil {
		return c, fmt.Errorf("list test results: %w", err)
	}
	for _, tr := range results {
		if err := m.UpsertTestResult(ctx, tr); err != nil {
			return c, fmt.Errorf("upsert test_result %s: %w", tr.ID, err)
		}
	}
	c.TestResults = len(results)

	return c, nil
}
