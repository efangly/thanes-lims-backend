package environment

import (
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
)

// PartnerDevice maps one physical third-party sensor unit (identified by
// its Serial, per docs/partner-api-guide.md) to the Location whose Gauge
// its readings are evaluated against - see CONTEXT.md#environment and ADR
// 0011. Location must reference an existing Gauge; it is never created
// here. Active lets an admin pause polling for a device without deleting
// the mapping (e.g. a unit temporarily removed for service).
type PartnerDevice struct {
	Serial   string
	Location string
	Active   bool
}

func (d PartnerDevice) Validate() error {
	if strings.TrimSpace(d.Serial) == "" || strings.TrimSpace(d.Location) == "" {
		return shared.ErrValidation
	}
	return nil
}
