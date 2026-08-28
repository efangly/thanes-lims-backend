package vendor

import (
	"context"

	domainvendor "github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	portvendor "github.com/efangly/thanes-lims-backend/internal/ports/vendor"
)

type ListVendorsUseCase struct {
	vendors portvendor.Repository
}

func NewListVendorsUseCase(vendors portvendor.Repository) *ListVendorsUseCase {
	return &ListVendorsUseCase{vendors: vendors}
}

func (uc *ListVendorsUseCase) Execute(ctx context.Context) ([]domainvendor.Vendor, error) {
	return uc.vendors.List(ctx)
}

type GetVendorUseCase struct {
	vendors portvendor.Repository
}

func NewGetVendorUseCase(vendors portvendor.Repository) *GetVendorUseCase {
	return &GetVendorUseCase{vendors: vendors}
}

func (uc *GetVendorUseCase) Execute(ctx context.Context, id string) (domainvendor.Vendor, error) {
	return uc.vendors.FindByID(ctx, id)
}
