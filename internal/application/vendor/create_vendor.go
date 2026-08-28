package vendor

import (
	"context"
	"errors"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainvendor "github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portvendor "github.com/efangly/thanes-lims-backend/internal/ports/vendor"
)

type CreateVendorUseCase struct {
	vendors portvendor.Repository
	idgen   portidgen.SequenceGenerator
}

func NewCreateVendorUseCase(vendors portvendor.Repository, idgen portidgen.SequenceGenerator) *CreateVendorUseCase {
	return &CreateVendorUseCase{vendors: vendors, idgen: idgen}
}

type CreateVendorInput struct {
	Name         string
	ContactName  string
	ContactPhone string
	ContactEmail string
	Address      string
}

// Execute creates a Vendor. Name must be unique across all non-Retired
// Vendors (see CONTEXT.md#vendors).
func (uc *CreateVendorUseCase) Execute(ctx context.Context, in CreateVendorInput) (domainvendor.Vendor, error) {
	v := domainvendor.Vendor{
		Name:         strings.TrimSpace(in.Name),
		ContactName:  in.ContactName,
		ContactPhone: in.ContactPhone,
		ContactEmail: in.ContactEmail,
		Address:      in.Address,
	}
	if err := v.Validate(); err != nil {
		return domainvendor.Vendor{}, err
	}

	if _, err := uc.vendors.FindByName(ctx, v.Name); err == nil {
		return domainvendor.Vendor{}, shared.ErrConflict
	} else if !errors.Is(err, shared.ErrNotFound) {
		return domainvendor.Vendor{}, err
	}

	id, err := nextVendorID(ctx, uc.idgen)
	if err != nil {
		return domainvendor.Vendor{}, err
	}
	v.ID = id

	return uc.vendors.Create(ctx, v)
}
