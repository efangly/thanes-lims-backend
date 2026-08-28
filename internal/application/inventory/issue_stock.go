package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

// IssueStockUseCase records a withdrawal (เบิกออก) against one or more
// InventoryLots the user picks explicitly - never an auto-FEFO selection
// (CONTEXT.md "Stock Issue"). Each line draws one lot down by its quantity.
//
// When a line's quantity exceeds its lot's recorded balance the use case,
// unless Force is set, applies nothing and returns the real remaining
// balances so the caller can either add another lot line or re-submit with
// Force. Forcing lets the lot's Quantity go below zero (ADR 0008) - a
// negative balance is the signal of a physical-count discrepancy, not a
// validation error.
type IssueStockUseCase struct {
	items portinventory.Repository
	lots  portinventory.LotRepository
}

func NewIssueStockUseCase(items portinventory.Repository, lots portinventory.LotRepository) *IssueStockUseCase {
	return &IssueStockUseCase{items: items, lots: lots}
}

// IssueLine is one lot-specific withdrawal within a Stock Issue.
type IssueLine struct {
	LotID    string
	Quantity int
}

type IssueStockInput struct {
	ItemID string
	Lines  []IssueLine
	Force  bool
}

// Shortfall reports a line whose requested quantity is more than the lot's
// recorded balance.
type Shortfall struct {
	LotID     string
	LotNo     string
	Requested int
	Available int
}

// IssueStockResult carries the item with its recomputed on-hand Quantity.
// When Applied is false the withdrawal was not performed because one or
// more lines over-issued and Force was not set - inspect Shortfalls.
type IssueStockResult struct {
	Applied    bool
	Shortfalls []Shortfall
	Item       inventory.InventoryItem
	Lots       []inventory.InventoryLot
}

func (uc *IssueStockUseCase) Execute(ctx context.Context, in IssueStockInput) (IssueStockResult, error) {
	if len(in.Lines) == 0 {
		return IssueStockResult{}, fmt.Errorf("%w: at least one issue line is required", shared.ErrValidation)
	}

	if _, err := uc.items.FindByID(ctx, in.ItemID); err != nil {
		return IssueStockResult{}, err
	}

	seen := make(map[string]bool, len(in.Lines))
	resolved := make([]inventory.InventoryLot, len(in.Lines))
	for i, line := range in.Lines {
		lotID := strings.TrimSpace(line.LotID)
		if lotID == "" {
			return IssueStockResult{}, fmt.Errorf("%w: lot_id is required on every line", shared.ErrValidation)
		}
		if line.Quantity <= 0 {
			return IssueStockResult{}, fmt.Errorf("%w: quantity must be greater than zero", shared.ErrValidation)
		}
		if seen[lotID] {
			return IssueStockResult{}, fmt.Errorf("%w: lot %s appears more than once - combine into one line", shared.ErrValidation, lotID)
		}
		seen[lotID] = true

		lot, err := uc.lots.FindByID(ctx, lotID)
		if err != nil {
			return IssueStockResult{}, err
		}
		if lot.ItemID != in.ItemID {
			return IssueStockResult{}, fmt.Errorf("%w: lot %s does not belong to item %s", shared.ErrValidation, lotID, in.ItemID)
		}
		resolved[i] = lot
	}

	var shortfalls []Shortfall
	for i, lot := range resolved {
		if in.Lines[i].Quantity > lot.Quantity {
			shortfalls = append(shortfalls, Shortfall{
				LotID:     lot.ID,
				LotNo:     lot.LotNo,
				Requested: in.Lines[i].Quantity,
				Available: lot.Quantity,
			})
		}
	}

	if len(shortfalls) > 0 && !in.Force {
		item, err := uc.items.FindByID(ctx, in.ItemID)
		if err != nil {
			return IssueStockResult{}, err
		}
		return IssueStockResult{Applied: false, Shortfalls: shortfalls, Item: item, Lots: resolved}, nil
	}

	updated := make([]inventory.InventoryLot, len(resolved))
	for i, lot := range resolved {
		u, err := uc.lots.UpdateQuantity(ctx, lot.ID, lot.Quantity-in.Lines[i].Quantity)
		if err != nil {
			return IssueStockResult{}, err
		}
		updated[i] = u
	}

	item, err := uc.items.FindByID(ctx, in.ItemID)
	if err != nil {
		return IssueStockResult{}, err
	}
	return IssueStockResult{Applied: true, Shortfalls: shortfalls, Item: item, Lots: updated}, nil
}
