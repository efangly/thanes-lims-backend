package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	portspurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PurchaseOrderTools implements the PurchaseOrder-domain MCP tools:
// listPurchaseOrders, listPurchaseOrdersByItem.
type PurchaseOrderTools struct {
	Orders portspurchaseorder.Repository
}

type PurchaseOrderOutput struct {
	ID        string `json:"id" jsonschema:"Purchase order ID, e.g. PO-2569-0012."`
	ItemID    string `json:"itemId" jsonschema:"Inventory item ID this PO orders more of."`
	Quantity  int    `json:"quantity" jsonschema:"Quantity ordered."`
	Vendor    string `json:"vendor" jsonschema:"Vendor name the order was placed with."`
	OrderDate string `json:"orderDate" jsonschema:"RFC3339 date the order was placed."`
	Status    string `json:"status" jsonschema:"Order status: pending_approval (awaiting internal sign-off), sent_to_vendor (approved and sent), received (goods arrived), cancelled."`
}

func toPurchaseOrderOutput(po purchaseorder.PurchaseOrder) PurchaseOrderOutput {
	return PurchaseOrderOutput{
		ID:        po.ID,
		ItemID:    po.ItemID,
		Quantity:  po.Quantity,
		Vendor:    po.Vendor,
		OrderDate: po.OrderDate.Format(time.RFC3339),
		Status:    string(po.Status),
	}
}

// --- listPurchaseOrders ---

type ListPurchaseOrdersInput struct {
	// portspurchaseorder.Repository.List has no status filter, so this is
	// applied client-side in the handler after a full List call - the
	// purchase_orders table is small (tens of rows), so this avoids widening
	// the repository interface for a single chatbot query.
	Status *string `json:"status,omitempty" jsonschema:"When set, only return orders with this status: pending_approval, sent_to_vendor, received, or cancelled."`
}

type ListPurchaseOrdersOutput struct {
	Orders []PurchaseOrderOutput `json:"orders" jsonschema:"Matching purchase orders."`
	Count  int                   `json:"count" jsonschema:"Number of matching purchase orders."`
}

// List implements the listPurchaseOrders tool.
func (t *PurchaseOrderTools) List(ctx context.Context, _ *mcp.CallToolRequest, in ListPurchaseOrdersInput) (*mcp.CallToolResult, ListPurchaseOrdersOutput, error) {
	if in.Status != nil {
		switch purchaseorder.Status(*in.Status) {
		case purchaseorder.StatusPendingApproval, purchaseorder.StatusSentToVendor, purchaseorder.StatusReceived, purchaseorder.StatusCancelled:
		default:
			return errorResult[ListPurchaseOrdersOutput](fmt.Sprintf("invalid status %q: must be one of pending_approval, sent_to_vendor, received, cancelled", *in.Status))
		}
	}

	orders, err := t.Orders.List(ctx)
	if err != nil {
		return nil, ListPurchaseOrdersOutput{}, err
	}
	out := make([]PurchaseOrderOutput, 0, len(orders))
	for _, po := range orders {
		if in.Status != nil && string(po.Status) != *in.Status {
			continue
		}
		out = append(out, toPurchaseOrderOutput(po))
	}
	return nil, ListPurchaseOrdersOutput{Orders: out, Count: len(out)}, nil
}

// --- listPurchaseOrdersByItem ---

type ListPurchaseOrdersByItemInput struct {
	InventoryItemID string `json:"inventoryItemId" jsonschema:"Inventory item ID to find purchase orders for, e.g. INV-0002."`
}

// ListByItem implements the listPurchaseOrdersByItem tool.
func (t *PurchaseOrderTools) ListByItem(ctx context.Context, _ *mcp.CallToolRequest, in ListPurchaseOrdersByItemInput) (*mcp.CallToolResult, ListPurchaseOrdersOutput, error) {
	if strings.TrimSpace(in.InventoryItemID) == "" {
		return errorResult[ListPurchaseOrdersOutput]("inventoryItemId is required")
	}

	orders, err := t.Orders.List(ctx)
	if err != nil {
		return nil, ListPurchaseOrdersOutput{}, err
	}
	out := make([]PurchaseOrderOutput, 0)
	for _, po := range orders {
		if po.ItemID == in.InventoryItemID {
			out = append(out, toPurchaseOrderOutput(po))
		}
	}
	return nil, ListPurchaseOrdersOutput{Orders: out, Count: len(out)}, nil
}
