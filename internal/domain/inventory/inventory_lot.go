package inventory

import "time"

// InventoryLot is one received batch of stock for an InventoryItem
// (CONTEXT.md "Inventory Lot"). An item's on-hand Quantity is the sum of its
// lots' Quantities, never written directly. A lot's Quantity is allowed to
// go below zero when a Stock Issue is deliberately forced past its recorded
// balance (ADR 0008) - a negative value signals a physical-count
// discrepancy to investigate, not a data-integrity violation.
type InventoryLot struct {
	ID         string
	ItemID     string
	LotNo      string
	ExpireDate *time.Time
	Quantity   int
}
