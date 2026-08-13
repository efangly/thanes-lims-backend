// Package pdf renders application PDFs (test result report, audit export).
// It embeds the Sarabun font (SIL OFL, see fonts/OFL.txt) so Thai text - the
// language every domain label in this system is written in - renders
// correctly; gofpdf's core fonts only cover Latin-1.
package pdf

import (
	"bytes"
	_ "embed"

	"github.com/jung-kurt/gofpdf"
)

//go:embed fonts/Sarabun-Regular.ttf
var sarabunRegular []byte

//go:embed fonts/Sarabun-Bold.ttf
var sarabunBold []byte

const fontFamily = "Sarabun"

// newDocument returns an A4 portrait Fpdf with the Sarabun font registered
// and selected, ready for content.
func newDocument() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(fontFamily, "", sarabunRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", sarabunBold)
	pdf.AddPage()
	pdf.SetFont(fontFamily, "", 11)
	return pdf
}

func output(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
