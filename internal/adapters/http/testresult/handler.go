package testresult

import (
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	"github.com/efangly/thanes-lims-backend/internal/adapters/pdf"
	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create    *applicationtestresult.CreateTestResultUseCase
	submit    *applicationtestresult.SubmitResultUseCase
	approve   *applicationtestresult.ApproveResultUseCase
	list      *applicationtestresult.ListTestResultsUseCase
	get       *applicationtestresult.GetTestResultUseCase
	genReport *applicationtestresult.GenerateReportUseCase
}

func NewHandler(
	create *applicationtestresult.CreateTestResultUseCase,
	submit *applicationtestresult.SubmitResultUseCase,
	approve *applicationtestresult.ApproveResultUseCase,
	list *applicationtestresult.ListTestResultsUseCase,
	get *applicationtestresult.GetTestResultUseCase,
	genReport *applicationtestresult.GenerateReportUseCase,
) *Handler {
	return &Handler{create: create, submit: submit, approve: approve, list: list, get: get, genReport: genReport}
}

// Create godoc
//
//	@Summary		สร้างรายการทดสอบใหม่
//	@Tags			test-results
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateTestResultRequest	true	"ข้อมูลการทดสอบ"
//	@Success		201		{object}	response.Envelope{data=TestResultResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/tests [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateTestResultRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	t, err := h.create.Execute(c.Context(), applicationtestresult.CreateTestResultInput{
		SampleID: req.SampleID,
		TestName: req.TestName,
		Analyst:  req.Analyst,
		RefRange: req.RefRange,
	})
	if err != nil {
		return err
	}
	return response.Created(c, toResponse(t))
}

// List godoc
//
//	@Summary		รายการผลทดสอบทั้งหมด
//	@Description	รองรับ filter ผ่าน query param `sample_id` และ `status`
//	@Tags			test-results
//	@Produce		json
//	@Security		BearerAuth
//	@Param			sample_id	query		string	false	"กรองตาม Sample ID"
//	@Param			status		query		string	false	"กรองตามสถานะ"
//	@Success		200			{object}	response.Envelope{data=[]TestResultResponse}
//	@Failure		401			{object}	response.Envelope
//	@Router			/tests [get]
func (h *Handler) List(c fiber.Ctx) error {
	var filter porttestresult.ListFilter
	if sampleID := c.Query("sample_id"); sampleID != "" {
		filter.SampleID = &sampleID
	}
	if status := c.Query("status"); status != "" {
		s := testresult.Status(status)
		filter.Status = &s
	}

	results, err := h.list.Execute(c.Context(), filter)
	if err != nil {
		return err
	}
	out := make([]TestResultResponse, len(results))
	for i, t := range results {
		out[i] = toResponse(t)
	}
	return response.OK(c, out)
}

// ListBySample godoc
//
//	@Summary		รายการผลทดสอบของตัวอย่างหนึ่ง
//	@Description	ListBySample serves the nested GET /samples/:id/tests convenience route.
//	@Tags			test-results
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Sample ID"
//	@Success		200	{object}	response.Envelope{data=[]TestResultResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/samples/{id}/tests [get]
func (h *Handler) ListBySample(c fiber.Ctx) error {
	sampleID := c.Params("id")
	results, err := h.list.Execute(c.Context(), porttestresult.ListFilter{SampleID: &sampleID})
	if err != nil {
		return err
	}
	out := make([]TestResultResponse, len(results))
	for i, t := range results {
		out[i] = toResponse(t)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงผลทดสอบตาม ID
//	@Tags			test-results
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Test Result ID"
//	@Success		200	{object}	response.Envelope{data=TestResultResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/tests/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	t, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(t))
}

// SubmitResult godoc
//
//	@Summary		บันทึกผลการทดสอบ
//	@Tags			test-results
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Test Result ID"
//	@Param			request	body		SubmitResultRequest	true	"ผลการทดสอบ"
//	@Success		200		{object}	response.Envelope{data=TestResultResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/tests/{id}/result [patch]
func (h *Handler) SubmitResult(c fiber.Ctx) error {
	var req SubmitResultRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	t, err := h.submit.Execute(c.Context(), applicationtestresult.SubmitResultInput{
		ID:     c.Params("id"),
		Result: req.Result,
		Flag:   testresult.Flag(req.Flag),
	})
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(t))
}

// GetReport godoc
//
//	@Summary		ดาวน์โหลดรายงานผลทดสอบ (PDF)
//	@Description	สร้าง PDF ผลการทดสอบ + ข้อมูลตัวอย่าง + chain-of-custody trail
//	@Tags			test-results
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Test Result ID"
//	@Success		200	{file}	byte
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/tests/{id}/report [get]
func (h *Handler) GetReport(c fiber.Ctx) error {
	id := c.Params("id")
	data, err := h.genReport.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	body, err := pdf.TestResultReport(pdf.TestResultReportData{
		Result:           data.Result,
		Sample:           data.Sample,
		CoCSteps:         data.CoCSteps,
		LocationFullPath: data.LocationFullPath,
	})
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s-report.pdf"`, id))
	return c.Send(body)
}

// Approve godoc
//
//	@Summary		อนุมัติผลทดสอบ
//	@Description	แจ้งเตือนผู้เกี่ยวข้องเมื่ออนุมัติสำเร็จ
//	@Tags			test-results
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Test Result ID"
//	@Success		200	{object}	response.Envelope{data=TestResultResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		403	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		409	{object}	response.Envelope
//	@Router			/tests/{id}/approve [patch]
func (h *Handler) Approve(c fiber.Ctx) error {
	role := fiber.Locals[domainuser.Role](c, middleware.LocalsRole)
	name := fiber.Locals[string](c, middleware.LocalsName)

	t, err := h.approve.Execute(c.Context(), applicationtestresult.ApproveResultInput{
		ID:        c.Params("id"),
		ActorRole: role,
		ActorName: name,
	})
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(t))
}
