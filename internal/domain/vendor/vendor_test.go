package vendor_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	"github.com/stretchr/testify/assert"
)

func TestVendor_Validate(t *testing.T) {
	cases := []struct {
		name    string
		v       vendor.Vendor
		wantErr bool
	}{
		{"ok minimal", vendor.Vendor{Name: "Acme"}, false},
		{"ok full", vendor.Vendor{Name: "Acme", ContactEmail: "sales@acme.com"}, false},
		{"blank name", vendor.Vendor{Name: "   "}, true},
		{"bad email", vendor.Vendor{Name: "Acme", ContactEmail: "not-an-email"}, true},
	}

	for _, tc := range cases {
		err := tc.v.Validate()
		if tc.wantErr {
			assert.ErrorIs(t, err, shared.ErrValidation, tc.name)
		} else {
			assert.NoError(t, err, tc.name)
		}
	}
}
