package pdf

import (
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/audit"
)

// AuditExport renders the given entries (already filtered/ordered by the
// caller) as a table: timestamp, actor, method+path, resource, status.
func AuditExport(entries []audit.AuditLog, from, to *time.Time) ([]byte, error) {
	pdf := newDocument()

	pdf.SetFont(fontFamily, "B", 16)
	pdf.CellFormat(0, 10, "รายงานประวัติการใช้งาน (Audit Log)", "", 1, "C", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(0, 6, "ช่วงเวลา: "+rangeLabel(from, to), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("จำนวนรายการ: %d", len(entries)), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	colWidths := []float64{28, 22, 45, 55, 15}
	headers := []string{"เวลา", "บทบาท", "การกระทำ", "ทรัพยากร", "สถานะ"}

	pdf.SetFont(fontFamily, "B", 9)
	pdf.SetFillColor(230, 230, 230)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont(fontFamily, "", 8)
	for _, e := range entries {
		actorRole := e.ActorRole
		if actorRole == "" {
			actorRole = "-"
		}
		action := e.Method + " " + e.Path
		resource := e.Resource
		if e.ResourceID != "" {
			resource = fmt.Sprintf("%s/%s", e.Resource, e.ResourceID)
		}

		pdf.CellFormat(colWidths[0], 7, e.CreatedAt.Format("2006-01-02 15:04"), "1", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, actorRole, "1", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, action, "1", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, resource, "1", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("%d", e.StatusCode), "1", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	return output(pdf)
}

func rangeLabel(from, to *time.Time) string {
	const layout = "2006-01-02"
	switch {
	case from != nil && to != nil:
		return from.Format(layout) + " ถึง " + to.Format(layout)
	case from != nil:
		return "ตั้งแต่ " + from.Format(layout)
	case to != nil:
		return "ถึง " + to.Format(layout)
	default:
		return "ทั้งหมด"
	}
}
