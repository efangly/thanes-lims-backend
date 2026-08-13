package audit

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/pdf"
	_ "github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	list *applicationaudit.ListAuditLogsUseCase
}

func NewHandler(list *applicationaudit.ListAuditLogsUseCase) *Handler {
	return &Handler{list: list}
}

// Export streams a PDF of audit log entries, optionally narrowed to a
// [from, to] date range via the `from`/`to` query params (RFC3339 or
// 2006-01-02).
// Export godoc
//
//	@Summary		ส่งออก audit log เป็น PDF
//	@Description	เฉพาะ admin/qa เท่านั้น รองรับกรองช่วงเวลาด้วย `from`/`to` (RFC3339 หรือ YYYY-MM-DD)
//	@Tags			audit
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			from	query	string	false	"วันที่เริ่มต้น (RFC3339 หรือ YYYY-MM-DD)"
//	@Param			to		query	string	false	"วันที่สิ้นสุด (RFC3339 หรือ YYYY-MM-DD)"
//	@Success		200	{file}		byte
//	@Failure		401	{object}	response.Envelope
//	@Failure		403	{object}	response.Envelope
//	@Router			/audit/export [get]
func (h *Handler) Export(c fiber.Ctx) error {
	filter := portaudit.ListFilter{}
	if from, ok := parseDate(c.Query("from")); ok {
		filter.From = &from
	}
	if to, ok := parseDate(c.Query("to")); ok {
		filter.To = &to
	}

	entries, err := h.list.Execute(c.Context(), filter)
	if err != nil {
		return err
	}

	body, err := pdf.AuditExport(entries, filter.From, filter.To)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="audit-export.pdf"`)
	return c.Send(body)
}

func parseDate(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t, true
	}
	return time.Time{}, false
}
