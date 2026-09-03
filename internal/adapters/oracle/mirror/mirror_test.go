package mirror

import (
	"testing"
	"unicode/utf8"

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

func TestClip(t *testing.T) {
	if got := clip("abc", 10); got != "abc" {
		t.Errorf("clip short = %q, want %q", got, "abc")
	}
	if got := clip("abcdef", 3); got != "abc" {
		t.Errorf("clip ascii = %q, want %q", got, "abc")
	}
	// "ก" is 3 bytes in UTF-8; a 4-byte budget must not split the second rune.
	if got := clip("กก", 4); got != "ก" {
		t.Errorf("clip thai = %q (% x), want %q", got, got, "ก")
	}
	if !utf8.ValidString(clip("ไมโครลิตร/หลอด", 20)) {
		t.Error("clip produced invalid UTF-8")
	}
}

func TestNullText(t *testing.T) {
	if got := nullText(""); got != nil {
		t.Errorf("nullText(%q) = %v, want nil", "", got)
	}
	if got := nullText("hi"); got != "hi" {
		t.Errorf("nullText(%q) = %v, want %q", "hi", got, "hi")
	}
}
