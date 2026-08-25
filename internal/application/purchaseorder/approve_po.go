package purchaseorder

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portpurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

type ApprovePOUseCase struct {
	purchaseOrders portpurchaseorder.Repository
}

func NewApprovePOUseCase(purchaseOrders portpurchaseorder.Repository) *ApprovePOUseCase {
	return &ApprovePOUseCase{purchaseOrders: purchaseOrders}
}

type ApprovePOInput struct {
	ID        string
	ActorRole domainuser.Role
}

func (uc *ApprovePOUseCase) Execute(ctx context.Context, in ApprovePOInput) (purchaseorder.PurchaseOrder, error) {
	// Approve permission: same Admin/QA set the removed Can(PermApprove)
	// matrix granted (see ADR 0002 - wiring this to the normalized RBAC
	// model is a later phase, out of scope here).
	if in.ActorRole != domainuser.RoleAdmin && in.ActorRole != domainuser.RoleQA {
		return purchaseorder.PurchaseOrder{}, shared.ErrForbidden
	}

	po, err := uc.purchaseOrders.FindByID(ctx, in.ID)
	if err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	if po.Status != purchaseorder.StatusPendingApproval {
		return purchaseorder.PurchaseOrder{}, fmt.Errorf("%w: cannot approve PO from status %s", shared.ErrValidation, po.Status)
	}

	po.Status = purchaseorder.StatusSentToVendor
	return uc.purchaseOrders.Update(ctx, po)
}
