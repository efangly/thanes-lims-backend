package purchaseorder

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portpurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

// AutoReorderJob is the scheduled counterpart to CreateFromLowStockUseCase's
// manual POST /inventory/:id/reorder flow: instead of a user picking a
// vendor and clicking reorder, it scans every item and reorders those that
// are both below min and have a DefaultVendor configured.
type AutoReorderJob struct {
	inventory      portinventory.Repository
	purchaseOrders portpurchaseorder.Repository
	reorder        *CreateFromLowStockUseCase
}

func NewAutoReorderJob(inventory portinventory.Repository, purchaseOrders portpurchaseorder.Repository, reorder *CreateFromLowStockUseCase) *AutoReorderJob {
	return &AutoReorderJob{inventory: inventory, purchaseOrders: purchaseOrders, reorder: reorder}
}

type AutoReorderResult struct {
	Created          []purchaseorder.PurchaseOrder
	SkippedNoVendor  []string // item IDs below min but with no DefaultVendor set
	SkippedHasOpenPO []string // item IDs below min that already have an unreceived PO in flight
}

func isOpen(status purchaseorder.Status) bool {
	return status == purchaseorder.StatusPendingApproval || status == purchaseorder.StatusSentToVendor
}

// Run scans all inventory items and creates a purchase order for each one
// that is below its minimum threshold and has a DefaultVendor configured.
// Items already covered by an open (not yet received/cancelled) PO are
// skipped so the job doesn't pile up duplicate orders on every tick while
// stock stays low; items below min without a DefaultVendor are also
// reported as skipped rather than failing the whole run, since one missing
// vendor shouldn't block reordering the rest.
func (j *AutoReorderJob) Run(ctx context.Context) (AutoReorderResult, error) {
	items, err := j.inventory.List(ctx)
	if err != nil {
		return AutoReorderResult{}, err
	}

	existingPOs, err := j.purchaseOrders.List(ctx)
	if err != nil {
		return AutoReorderResult{}, err
	}
	openItemIDs := make(map[string]bool, len(existingPOs))
	for _, po := range existingPOs {
		if isOpen(po.Status) {
			openItemIDs[po.ItemID] = true
		}
	}

	var result AutoReorderResult
	for _, item := range items {
		if !item.BelowMin() {
			continue
		}
		if openItemIDs[item.ID] {
			result.SkippedHasOpenPO = append(result.SkippedHasOpenPO, item.ID)
			continue
		}
		if item.DefaultVendor == "" {
			result.SkippedNoVendor = append(result.SkippedNoVendor, item.ID)
			continue
		}

		po, err := j.reorder.Execute(ctx, CreateFromLowStockInput{ItemID: item.ID, Vendor: item.DefaultVendor})
		if err != nil {
			return result, err
		}
		result.Created = append(result.Created, po)
	}

	return result, nil
}
