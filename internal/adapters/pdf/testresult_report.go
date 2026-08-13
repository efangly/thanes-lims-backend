package pdf

import (
	"fmt"

	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
)

// TestResultReportData is the adapter-facing view of everything the report
// needs; application/testresult.ReportData maps onto it 1:1 so this package
// stays free of an application-layer import.
type TestResultReportData struct {
	Result   testresult.TestResult
	Sample   domainsample.Sample
	CoCSteps []domainsample.CoCStep
}

func testResultStatusLabel(s testresult.Status) string {
	switch s {
	case testresult.StatusAnalyzing:
		return "กำลังวิเคราะห์"
	case testresult.StatusPendingVerification:
		return "รอตรวจสอบ"
	case testresult.StatusApproved:
		return "อนุมัติแล้ว"
	default:
		return string(s)
	}
}

func flagLabel(f testresult.Flag) string {
	switch f {
	case testresult.FlagHi:
		return "สูงกว่าเกณฑ์"
	case testresult.FlagLo:
		return "ต่ำกว่าเกณฑ์"
	case testresult.FlagOk:
		return "ปกติ"
	default:
		return string(f)
	}
}

// TestResultReport renders a single-page (or more, for a long CoC trail)
// PDF: result header, the linked sample's details, and its full
// chain-of-custody trail.
func TestResultReport(data TestResultReportData) ([]byte, error) {
	pdf := newDocument()

	pdf.SetFont(fontFamily, "B", 16)
	pdf.CellFormat(0, 10, "รายงานผลการทดสอบ", "", 1, "C", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(0, 6, "Thanes LIMS", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	labelValue := func(label, value string) {
		pdf.SetFont(fontFamily, "B", 11)
		pdf.CellFormat(45, 8, label, "", 0, "", false, 0, "")
		pdf.SetFont(fontFamily, "", 11)
		pdf.CellFormat(0, 8, value, "", 1, "", false, 0, "")
	}

	pdf.SetFont(fontFamily, "B", 13)
	pdf.CellFormat(0, 8, "ผลการทดสอบ", "", 1, "", false, 0, "")
	labelValue("รหัสผล:", data.Result.ID)
	labelValue("รายการทดสอบ:", data.Result.TestName)
	labelValue("ผู้วิเคราะห์:", data.Result.Analyst)
	labelValue("ผลลัพธ์:", data.Result.Result)
	labelValue("ค่าอ้างอิง:", data.Result.RefRange)
	labelValue("สถานะ Flag:", flagLabel(data.Result.Flag))
	labelValue("สถานะผล:", testResultStatusLabel(data.Result.Status))
	pdf.Ln(4)

	pdf.SetFont(fontFamily, "B", 13)
	pdf.CellFormat(0, 8, "ตัวอย่าง", "", 1, "", false, 0, "")
	labelValue("รหัสตัวอย่าง:", data.Sample.ID)
	labelValue("ชื่อตัวอย่าง:", data.Sample.Name)
	labelValue("ประเภท:", string(data.Sample.Type))
	labelValue("ผู้ดูแล:", data.Sample.Custodian)
	labelValue("สถานที่:", data.Sample.Location)
	labelValue("รับเข้าเมื่อ:", data.Sample.ReceivedAt.Format("2006-01-02 15:04"))
	pdf.Ln(4)

	pdf.SetFont(fontFamily, "B", 13)
	pdf.CellFormat(0, 8, "ประวัติการควบคุมตัวอย่าง (Chain of Custody)", "", 1, "", false, 0, "")
	if len(data.CoCSteps) == 0 {
		pdf.SetFont(fontFamily, "", 11)
		pdf.CellFormat(0, 8, "ไม่มีข้อมูล", "", 1, "", false, 0, "")
	}
	for i, step := range data.CoCSteps {
		pdf.SetFont(fontFamily, "", 11)
		line := fmt.Sprintf("%d. %s — %s (%s)", i+1, step.Title, step.Who, step.OccurredAt.Format("2006-01-02 15:04"))
		pdf.MultiCell(0, 7, line, "", "", false)
	}

	return output(pdf)
}
