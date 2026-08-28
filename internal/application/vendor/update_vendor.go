package vendor

import (
	"context"
	"errors"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainvendor "github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	portvendor "github.com/efangly/thanes-lims-backend/internal/ports/vendor"
)

type UpdateVendorUseCase struct {
	vendors portvendor.Repository
}

func NewUpdateVendorUseCase(vendors portvendor.Repository) *UpdateVendorUseCase {
	return &UpdateVendorUseCase{vendors: vendors}
}

type UpdateVendorInput struct {
	ID           string
	Name         string
	ContactName  string
	ContactPhone string
	ContactEmail string
	Address      string
}

func (uc *UpdateVendorUseCase) Execute(ctx context.Context, in UpdateVendorInput) (domainvendor.Vendor, error) {
	existing, err := uc.vendors.FindByID(ctx, in.ID)
	if err != nil {
		return domainvendor.Vendor{}, err
	}

	existing.Name = strings.TrimSpace(in.Name)
	existing.ContactName = in.ContactName
	existing.ContactPhone = in.ContactPhone
	existing.ContactEmail = in.ContactEmail
	existing.Address = in.Address

	if err := existing.Validate(); err != nil {
		return domainvendor.Vendor{}, err
	}

	// Name uniqueness: only a clash with a *different* Vendor is a conflict.
	if byName, err := uc.vendors.FindByName(ctx, existing.Name); err == nil {
		if byName.ID != existing.ID {
			return domainvendor.Vendor{}, shared.ErrConflict
		}
	} else if !errors.Is(err, shared.ErrNotFound) {
		return domainvendor.Vendor{}, err
	}

	return uc.vendors.Update(ctx, existing)
}
