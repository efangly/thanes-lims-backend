package vendor

import domainvendor "github.com/efangly/thanes-lims-backend/internal/domain/vendor"

type CreateVendorRequest struct {
	Name         string `json:"name" validate:"required"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email"`
	Address      string `json:"address"`
}

type UpdateVendorRequest struct {
	Name         string `json:"name" validate:"required"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email"`
	Address      string `json:"address"`
}

type VendorResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	Address      string `json:"address"`
}

func toResponse(v domainvendor.Vendor) VendorResponse {
	return VendorResponse{
		ID:           v.ID,
		Name:         v.Name,
		ContactName:  v.ContactName,
		ContactPhone: v.ContactPhone,
		ContactEmail: v.ContactEmail,
		Address:      v.Address,
	}
}
