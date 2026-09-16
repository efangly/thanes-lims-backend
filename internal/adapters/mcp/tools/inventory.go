package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portsinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InventoryTools implements the Inventory-domain MCP tools:
// listInventoryLowStock, getInventoryItemById. Lots is optional - when the
// caller doesn't need lot-level detail it can be left nil, but getInventoryItemById
// always tries to include lots since a LotRepository exists in this backend.
type InventoryTools struct {
	Items portsinventory.Repository
	Lots  portsinventory.LotRepository
}

type InventoryItemOutput struct {
	ID                 string               `json:"id" jsonschema:"Inventory item ID, e.g. INV-0001."`
	Name               string               `json:"name" jsonschema:"Item name, e.g. \"Glucose Reagent\"."`
	Category           string               `json:"category" jsonschema:"Item category, e.g. reagent, PPE, kit."`
	Quantity           int                  `json:"quantity" jsonschema:"Current on-hand quantity, summed across all lots."`
	Unit               string               `json:"unit" jsonschema:"Unit of measure for quantity, e.g. box, bottle, pair."`
	MinQuantity        int                  `json:"minQuantity" jsonschema:"Minimum quantity threshold; at or below this the item needs reordering."`
	MaxQuantity        int                  `json:"maxQuantity" jsonschema:"Maximum stocking quantity (full-stock reference level)."`
	BelowMin           bool                 `json:"belowMin" jsonschema:"True when quantity is at or below minQuantity - i.e. this item is low/out of stock."`
	DefaultVendor      string               `json:"defaultVendor,omitempty" jsonschema:"Default vendor name used for auto-reorder, if configured."`
	Manufacturer       string               `json:"manufacturer,omitempty" jsonschema:"Manufacturer name."`
	CustodianUserID    int64                `json:"custodianUserId" jsonschema:"User ID of the person responsible for this item."`
	EarliestExpireDate string               `json:"earliestExpireDate,omitempty" jsonschema:"RFC3339 date of the soonest-expiring lot, when any lot has an expiry."`
	LotCount           int                  `json:"lotCount" jsonschema:"Number of inventory lots backing this item."`
	Lots               []InventoryLotOutput `json:"lots,omitempty" jsonschema:"The item's individual received batches (lots), when available."`
}

type InventoryLotOutput struct {
	ID         string `json:"id" jsonschema:"Lot ID."`
	LotNo      string `json:"lotNo" jsonschema:"Lot number as printed on the physical batch."`
	Quantity   int    `json:"quantity" jsonschema:"Quantity remaining in this lot. Can be negative, which signals a physical-count discrepancy from a forced over-issue, not a data error."`
	ExpireDate string `json:"expireDate,omitempty" jsonschema:"RFC3339 expiry date of this lot, if it has one."`
}

func toInventoryItemOutput(i inventory.InventoryItem) InventoryItemOutput {
	out := InventoryItemOutput{
		ID:              i.ID,
		Name:            i.Name,
		Category:        i.Category,
		Quantity:        i.Quantity,
		Unit:            i.Unit,
		MinQuantity:     i.Min,
		MaxQuantity:     i.Max,
		BelowMin:        i.BelowMin(),
		DefaultVendor:   i.DefaultVendor,
		Manufacturer:    i.Manufacturer,
		CustodianUserID: i.CustodianUserID,
		LotCount:        i.LotCount,
	}
	if i.EarliestExpireDate != nil {
		out.EarliestExpireDate = i.EarliestExpireDate.Format(time.RFC3339)
	}
	return out
}

// --- listInventoryLowStock ---

type ListInventoryLowStockOutput struct {
	Items []InventoryItemOutput `json:"items" jsonschema:"Inventory items whose quantity is at or below their minQuantity threshold."`
	Count int                   `json:"count" jsonschema:"Number of matching items."`
}

// ListLowStock implements the listInventoryLowStock tool. It takes no
// input parameters.
func (t *InventoryTools) ListLowStock(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, ListInventoryLowStockOutput, error) {
	items, err := t.Items.List(ctx)
	if err != nil {
		return nil, ListInventoryLowStockOutput{}, err
	}
	out := make([]InventoryItemOutput, 0)
	for _, i := range items {
		if i.BelowMin() {
			out = append(out, toInventoryItemOutput(i))
		}
	}
	return nil, ListInventoryLowStockOutput{Items: out, Count: len(out)}, nil
}

// --- getInventoryItemById ---

type GetInventoryItemByIDInput struct {
	ID string `json:"id" jsonschema:"Inventory item ID to look up, e.g. INV-0002."`
}

// GetByID implements the getInventoryItemById tool.
func (t *InventoryTools) GetByID(ctx context.Context, _ *mcp.CallToolRequest, in GetInventoryItemByIDInput) (*mcp.CallToolResult, InventoryItemOutput, error) {
	if strings.TrimSpace(in.ID) == "" {
		return errorResult[InventoryItemOutput]("id is required")
	}
	item, err := t.Items.FindByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return errorResult[InventoryItemOutput](fmt.Sprintf("inventory item %q not found", in.ID))
		}
		return nil, InventoryItemOutput{}, err
	}
	out := toInventoryItemOutput(item)

	if t.Lots != nil {
		lots, err := t.Lots.ListByItem(ctx, in.ID)
		if err == nil {
			out.Lots = make([]InventoryLotOutput, 0, len(lots))
			for _, l := range lots {
				lo := InventoryLotOutput{ID: l.ID, LotNo: l.LotNo, Quantity: l.Quantity}
				if l.ExpireDate != nil {
					lo.ExpireDate = l.ExpireDate.Format(time.RFC3339)
				}
				out.Lots = append(out.Lots, lo)
			}
		}
	}
	return nil, out, nil
}
