package audit

import (
	"strconv"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/pdf"
	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
	"github.com/gofiber/fiber/v3"
)

const (
	defaultLogsLimit = 50
	maxLogsLimit     = 200
)

type Handler struct {
	list     *applicationaudit.ListAuditLogsUseCase
	listPage *applicationaudit.ListAuditLogsPageUseCase
}

func NewHandler(list *applicationaudit.ListAuditLogsUseCase, listPage *applicationaudit.ListAuditLogsPageUseCase) *Handler {
	return &Handler{list: list, listPage: listPage}
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

// ListLogs returns a paginated JSON page of audit trail entries, filterable
// by `actor_id`, `resource`, `method`, and the `from`/`to` date range (same
// parsing as Export).
// ListLogs godoc
//
//	@Summary		รายการ audit log (JSON)
//	@Description	รองรับกรองด้วย `actor_id`, `resource`, `method`, `from`, `to` และแบ่งหน้าด้วย `page`/`limit`
//	@Tags			audit
//	@Produce		json
//	@Security		BearerAuth
//	@Param			actor_id	query	int		false	"กรองตาม Actor ID"
//	@Param			resource	query	string	false	"กรองตาม resource"
//	@Param			method		query	string	false	"กรองตาม HTTP method"
//	@Param			from		query	string	false	"วันที่เริ่มต้น (RFC3339 หรือ YYYY-MM-DD)"
//	@Param			to			query	string	false	"วันที่สิ้นสุด (RFC3339 หรือ YYYY-MM-DD)"
//	@Param			page		query	int		false	"หน้า"		default(1)
//	@Param			limit		query	int		false	"จำนวนต่อหน้า"	default(50)
//	@Success		200	{object}	response.Envelope{data=[]EntryResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		403	{object}	response.Envelope
//	@Router			/audit/logs [get]
func (h *Handler) ListLogs(c fiber.Ctx) error {
	filter := portaudit.ListFilter{}
	if from, ok := parseDate(c.Query("from")); ok {
		filter.From = &from
	}
	if to, ok := parseDate(c.Query("to")); ok {
		filter.To = &to
	}
	if raw := c.Query("actor_id"); raw != "" {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
			filter.ActorID = &id
		}
	}
	if resource := c.Query("resource"); resource != "" {
		filter.Resource = resource
	}
	if method := c.Query("method"); method != "" {
		filter.Method = method
	}

	page := fiber.Query(c, "page", 1)
	if page < 1 {
		page = 1
	}
	limit := fiber.Query(c, "limit", defaultLogsLimit)
	if limit < 1 {
		limit = defaultLogsLimit
	}
	if limit > maxLogsLimit {
		limit = maxLogsLimit
	}
	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	result, err := h.listPage.Execute(c.Context(), filter)
	if err != nil {
		return err
	}

	out := make([]EntryResponse, len(result.Entries))
	for i, e := range result.Entries {
		out[i] = toEntryResponse(e)
	}
	return response.OK(c, out, PageMeta{Page: page, Limit: limit, Total: result.Total})
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
