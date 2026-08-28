package pdf

import (
	"testing"
	"time"
)

func TestSampleSticker_RendersEveryTemplateAndSymbology(t *testing.T) {
	data := SampleStickerData{
		ScanCode:         "SMP-BC-00001",
		SampleID:         "SMP-2569-00001",
		Name:             "เลือดผู้ป่วย OPD #01",
		TypeLabel:        "blood",
		CustodianName:    "สมชาย ใจดี",
		LocationFullPath: "Fridge-A / Shelf-2 / Slot-4",
		ReceivedAt:       time.Now(),
	}

	for _, tpl := range []string{StickerCap, StickerStem, StickerSmall, StickerMedium, "unknown"} {
		for _, sym := range []string{SymbologyCode128, SymbologyQR, ""} {
			body, err := SampleSticker(data, tpl, sym)
			if err != nil {
				t.Fatalf("template=%s symbology=%q: %v", tpl, sym, err)
			}
			if len(body) == 0 {
				t.Fatalf("template=%s symbology=%q: empty PDF", tpl, sym)
			}
		}
	}
}

func TestNormalizeSymbology_CapDefaultsToQR(t *testing.T) {
	if got := NormalizeSymbology(StickerCap, ""); got != SymbologyQR {
		t.Fatalf("cap default = %q, want qr", got)
	}
	if got := NormalizeSymbology(StickerMedium, ""); got != SymbologyCode128 {
		t.Fatalf("medium default = %q, want code128", got)
	}
	if got := NormalizeSymbology(StickerCap, SymbologyCode128); got != SymbologyCode128 {
		t.Fatalf("explicit choice not honoured: %q", got)
	}
}
