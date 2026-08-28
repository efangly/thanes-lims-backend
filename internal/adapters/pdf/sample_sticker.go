package pdf

import (
	"fmt"
	"time"

	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/barcode"
)

// SampleStickerData is the adapter-facing view of one Sample's label. The
// tiny templates use only ScanCode + SampleID; the larger ones use the rest.
type SampleStickerData struct {
	ScanCode         string // what the barcode/QR encodes
	SampleID         string
	Name             string
	TypeLabel        string
	CustodianName    string
	LocationFullPath string
	ReceivedAt       time.Time
}

// Sticker symbology.
const (
	SymbologyCode128 = "code128"
	SymbologyQR      = "qr"
)

// Sticker templates - physical label sizes in millimetres.
const (
	StickerCap    = "cap"    // 9.5 x 6.4  - microtube cap dot
	StickerStem   = "stem"   // 20.5 x 6.5 - microtube stem
	StickerSmall  = "small"  // 40 x 20    - vial / cryo
	StickerMedium = "medium" // 60 x 30    - general purpose
)

type stickerTemplate struct {
	w, h float64
}

var stickerTemplates = map[string]stickerTemplate{
	StickerCap:    {9.5, 6.4},
	StickerStem:   {20.5, 6.5},
	StickerSmall:  {40, 20},
	StickerMedium: {60, 30},
}

// StickerTemplateSize returns the label's millimetre dimensions and whether
// the template key is known.
func StickerTemplateSize(template string) (w, h float64, ok bool) {
	t, ok := stickerTemplates[template]
	return t.w, t.h, ok
}

// NormalizeSymbology returns a valid symbology for the template, defaulting
// to QR for the tiny cap label (a 1D barcode is unscannable at 9.5mm) and to
// Code128 otherwise. An explicit valid choice is always honoured.
func NormalizeSymbology(template, symbology string) string {
	switch symbology {
	case SymbologyCode128, SymbologyQR:
		return symbology
	}
	if template == StickerCap {
		return SymbologyQR
	}
	return SymbologyCode128
}

// SampleSticker renders one Sample label as a single-page PDF sized to the
// chosen template. Unknown template keys fall back to StickerMedium.
func SampleSticker(data SampleStickerData, template, symbology string) ([]byte, error) {
	tpl, ok := stickerTemplates[template]
	if !ok {
		template, tpl = StickerMedium, stickerTemplates[StickerMedium]
	}
	symbology = NormalizeSymbology(template, symbology)

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: tpl.w, Ht: tpl.h},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddUTF8FontFromBytes(fontFamily, "", sarabunRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", sarabunBold)
	pdf.AddPage()

	const pad = 0.6

	registerSymbol := func() string {
		if symbology == SymbologyQR {
			return barcode.RegisterQR(pdf, data.ScanCode, qr.M, qr.Unicode)
		}
		return barcode.RegisterCode128(pdf, data.ScanCode)
	}

	switch template {
	case StickerCap:
		key := registerSymbol()
		s := tpl.h - 2*pad
		barcode.Barcode(pdf, key, (tpl.w-s)/2, pad, s, s, false)

	case StickerStem:
		key := registerSymbol()
		symW := tpl.h - 2*pad
		if symbology == SymbologyCode128 {
			symW = 12
		}
		barcode.Barcode(pdf, key, pad, pad, symW, tpl.h-2*pad, false)
		pdf.SetFont(fontFamily, "B", 5)
		pdf.SetXY(pad+symW+0.5, tpl.h/2-2)
		pdf.CellFormat(tpl.w-symW-2*pad-0.5, 4, data.ScanCode, "", 0, "L", false, 0, "")

	case StickerSmall:
		key := registerSymbol()
		symH := 10.0
		symW := tpl.w - 2*pad
		if symbology == SymbologyQR {
			symW = symH
		}
		barcode.Barcode(pdf, key, (tpl.w-symW)/2, pad, symW, symH, false)
		pdf.SetFont(fontFamily, "", 6)
		pdf.SetXY(pad, symH+pad)
		pdf.CellFormat(tpl.w-2*pad, 3.5, data.ScanCode, "", 2, "C", false, 0, "")
		pdf.SetFont(fontFamily, "B", 7)
		pdf.CellFormat(tpl.w-2*pad, 4, data.SampleID, "", 2, "C", false, 0, "")

	default: // StickerMedium
		key := registerSymbol()
		symSize := 22.0
		symW := symSize
		if symbology == SymbologyCode128 {
			symW = 24
		}
		barcode.Barcode(pdf, key, pad, pad, symW, symSize, false)
		pdf.SetFont(fontFamily, "", 6)
		pdf.SetXY(pad, pad+symSize)
		pdf.CellFormat(symW, 3, data.ScanCode, "", 0, "C", false, 0, "")

		textX := pad + symW + 1.5
		textW := tpl.w - textX - pad
		pdf.SetXY(textX, pad)
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(textW, 4.5, data.SampleID, "", 2, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 7)
		pdf.CellFormat(textW, 4, truncate(data.Name, 34), "", 2, "L", false, 0, "")
		if data.TypeLabel != "" {
			pdf.CellFormat(textW, 4, "ประเภท: "+data.TypeLabel, "", 2, "L", false, 0, "")
		}
		if data.CustodianName != "" && data.CustodianName != "-" {
			pdf.CellFormat(textW, 4, "ผู้ดูแล: "+truncate(data.CustodianName, 26), "", 2, "L", false, 0, "")
		}
		if data.LocationFullPath != "" && data.LocationFullPath != "-" {
			pdf.CellFormat(textW, 4, truncate(data.LocationFullPath, 34), "", 2, "L", false, 0, "")
		}
		if !data.ReceivedAt.IsZero() {
			pdf.CellFormat(textW, 4, "รับเข้า: "+data.ReceivedAt.Format("2006-01-02"), "", 2, "L", false, 0, "")
		}
	}

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("render sample sticker: %w", err)
	}
	return output(pdf)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
