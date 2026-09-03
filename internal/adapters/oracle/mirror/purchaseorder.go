package mirror

import (
	"context"
	"database/sql"

	dpo "github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	portpo "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

const mergePurchaseOrderSQL = `
MERGE INTO purchase_orders t
USING (SELECT :id AS id FROM dual) s ON (t.id = s.id)
WHEN MATCHED THEN UPDATE SET
    item_id = :item, quantity = :qty, vendor = :vendor,
    order_date = :odate, status = :status
WHEN NOT MATCHED THEN INSERT
    (id, item_id, quantity, vendor, order_date, status)
VALUES (:id, :item, :qty, :vendor, :odate, :status)
`

// UpsertPurchaseOrder MERGEs one PO into the mirror (1:1 mapping).
func (m *Mirror) UpsertPurchaseOrder(ctx context.Context, po dpo.PurchaseOrder) error {
	_, err := m.db.ExecContext(ctx, mergePurchaseOrderSQL,
		sql.Named("id", po.ID),
		sql.Named("item", po.ItemID),
		sql.Named("qty", po.Quantity),
		sql.Named("vendor", clip(po.Vendor, 200)),
		sql.Named("odate", po.OrderDate),
		sql.Named("status", string(po.Status)),
	)
	return err
}

type purchaseOrderRepo struct {
	portpo.Repository
	mirror *Mirror
}

// WrapPO returns next wrapped so Create/Update also mirror to Oracle.
func WrapPO(next portpo.Repository, m *Mirror) portpo.Repository {
	return &purchaseOrderRepo{Repository: next, mirror: m}
}

func (r *purchaseOrderRepo) Create(ctx context.Context, po dpo.PurchaseOrder) (dpo.PurchaseOrder, error) {
	created, err := r.Repository.Create(ctx, po)
	if err != nil {
		return created, err
	}
	run(ctx, "purchase_order", created.ID, func(ctx context.Context) error {
		return r.mirror.UpsertPurchaseOrder(ctx, created)
	})
	return created, nil
}

func (r *purchaseOrderRepo) Update(ctx context.Context, po dpo.PurchaseOrder) (dpo.PurchaseOrder, error) {
	updated, err := r.Repository.Update(ctx, po)
	if err != nil {
		return updated, err
	}
	run(ctx, "purchase_order", updated.ID, func(ctx context.Context) error {
		return r.mirror.UpsertPurchaseOrder(ctx, updated)
	})
	return updated, nil
}
