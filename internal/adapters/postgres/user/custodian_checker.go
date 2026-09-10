package user

import (
	"context"

	"gorm.io/gorm"
)

// CustodianChecker implements portuser.CustodianChecker by counting the
// non-Retired Samples and Inventory Items that still point at a User as
// their Custodian. It lives in this package (not sample/inventory) only
// because it is the retire-User guard's dependency - it touches no domain
// type, just two `custodian_user_id` columns.
type CustodianChecker struct {
	db *gorm.DB
}

func NewCustodianChecker(db *gorm.DB) *CustodianChecker {
	return &CustodianChecker{db: db}
}

func (c *CustodianChecker) CountCustodianRefs(ctx context.Context, userID int64) (samples int64, inventoryItems int64, err error) {
	if err = c.db.WithContext(ctx).Table("samples").
		Where("custodian_user_id = ? AND deleted_at IS NULL", userID).
		Count(&samples).Error; err != nil {
		return 0, 0, err
	}
	if err = c.db.WithContext(ctx).Table("inventory_items").
		Where("custodian_user_id = ? AND deleted_at IS NULL", userID).
		Count(&inventoryItems).Error; err != nil {
		return 0, 0, err
	}
	return samples, inventoryItems, nil
}
