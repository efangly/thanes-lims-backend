package mirror

import (
	"testing"

	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portpo "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

// The decorators must stay drop-in replacements for the Postgres repos at the
// composition root.
var (
	_ portsample.SampleRepository = (*sampleRepo)(nil)
	_ porttestresult.Repository   = (*testResultRepo)(nil)
	_ portinventory.Repository    = (*inventoryRepo)(nil)
	_ portinventory.LotRepository = (*lotRepo)(nil)
	_ portpo.Repository           = (*purchaseOrderRepo)(nil)
)

func TestNullText(t *testing.T) {
	if got := nullText(""); got != nil {
		t.Errorf("nullText(%q) = %v, want nil", "", got)
	}
	if got := nullText("hi"); got != "hi" {
		t.Errorf("nullText(%q) = %v, want %q", "hi", got, "hi")
	}
}
