package mirror

import (
	"context"
	"database/sql"

	dinventory "github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

const mergeInventoryItemSQL = `
MERGE INTO inventory_items t
USING (SELECT :id AS id FROM dual) s ON (t.id = s.id)
WHEN MATCHED THEN UPDATE SET
    name = :name, category = :cat, quantity = :qty, unit = :unit,
    min_qty = :minq, max_qty = :maxq, default_vendor = :vendor
WHEN NOT MATCHED THEN INSERT
    (id, name, category, quantity, unit, min_qty, max_qty, default_vendor)
VALUES (:id, :name, :cat, :qty, :unit, :minq, :maxq, :vendor)
`

// UpsertInventoryItem MERGEs one item into the mirror. quantity is the derived
// on-hand total (sum of lots); the mirror has no lots table, so callers pass
// an item whose Quantity has been populated by a repository read.
func (m *Mirror) UpsertInventoryItem(ctx context.Context, i dinventory.InventoryItem) error {
	_, err := m.db.ExecContext(ctx, mergeInventoryItemSQL,
		sql.Named("id", i.ID),
		sql.Named("name", i.Name),
		sql.Named("cat", nullText(i.Category)),
		sql.Named("qty", i.Quantity),
		sql.Named("unit", i.Unit),
		sql.Named("minq", i.Min),
		sql.Named("maxq", i.Max),
		sql.Named("vendor", nullText(i.DefaultVendor)),
	)
	return err
}

// itemReader re-reads an item so the mirror gets its derived Quantity.
type itemReader interface {
	FindByID(ctx context.Context, id string) (dinventory.InventoryItem, error)
}

type inventoryRepo struct {
	portinventory.Repository
	mirror *Mirror
}

// WrapInventory returns next wrapped so item writes also mirror to Oracle.
func WrapInventory(next portinventory.Repository, m *Mirror) portinventory.Repository {
	return &inventoryRepo{Repository: next, mirror: m}
}

func (r *inventoryRepo) Create(ctx context.Context, i dinventory.InventoryItem) (dinventory.InventoryItem, error) {
	created, err := r.Repository.Create(ctx, i)
	if err != nil {
		return created, err
	}
	r.mirrorItem(ctx, created)
	return created, nil
}

func (r *inventoryRepo) Update(ctx context.Context, i dinventory.InventoryItem) (dinventory.InventoryItem, error) {
	updated, err := r.Repository.Update(ctx, i)
	if err != nil {
		return updated, err
	}
	r.mirrorItem(ctx, updated)
	return updated, nil
}

func (r *inventoryRepo) UpdateDefaultVendor(ctx context.Context, id string, vendor string) (dinventory.InventoryItem, error) {
	updated, err := r.Repository.UpdateDefaultVendor(ctx, id, vendor)
	if err != nil {
		return updated, err
	}
	r.mirrorItem(ctx, updated)
	return updated, nil
}

func (r *inventoryRepo) mirrorItem(ctx context.Context, i dinventory.InventoryItem) {
	run(ctx, "inventory_item", i.ID, func(ctx context.Context) error {
		return r.mirror.UpsertInventoryItem(ctx, i)
	})
}

// lotRepo decorates the lot repository. The mirror has no lots table, but a
// lot write changes an item's derived on-hand quantity, so after each
// successful lot write it re-reads the parent item and re-MERGEs it.
type lotRepo struct {
	portinventory.LotRepository
	mirror *Mirror
	items  itemReader
}

// WrapInventoryLot returns next wrapped so lot writes re-mirror the parent
// item's quantity. items reads the authoritative derived quantity.
func WrapInventoryLot(next portinventory.LotRepository, m *Mirror, items itemReader) portinventory.LotRepository {
	return &lotRepo{LotRepository: next, mirror: m, items: items}
}

func (r *lotRepo) Create(ctx context.Context, l dinventory.InventoryLot) (dinventory.InventoryLot, error) {
	created, err := r.LotRepository.Create(ctx, l)
	if err != nil {
		return created, err
	}
	r.remirrorItem(ctx, created.ItemID)
	return created, nil
}

func (r *lotRepo) UpdateQuantity(ctx context.Context, id string, quantity int) (dinventory.InventoryLot, error) {
	updated, err := r.LotRepository.UpdateQuantity(ctx, id, quantity)
	if err != nil {
		return updated, err
	}
	r.remirrorItem(ctx, updated.ItemID)
	return updated, nil
}

func (r *lotRepo) remirrorItem(ctx context.Context, itemID string) {
	run(ctx, "inventory_item", itemID, func(ctx context.Context) error {
		item, err := r.items.FindByID(ctx, itemID)
		if err != nil {
			return err
		}
		return r.mirror.UpsertInventoryItem(ctx, item)
	})
}
