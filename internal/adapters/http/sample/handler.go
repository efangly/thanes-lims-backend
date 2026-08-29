package sample

import (
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	"github.com/efangly/thanes-lims-backend/internal/adapters/pdf"
	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create          *applicationsample.CreateSampleUseCase
	list            *applicationsample.ListSamplesUseCase
	get             *applicationsample.GetSampleUseCase
	updateStatus    *applicationsample.UpdateSampleStatusUseCase
	listCoC         *applicationsample.ListCoCStepsUseCase
	appendCoC       *applicationsample.AppendCoCStepUseCase
	assignLocation  *applicationsample.AssignLocationUseCase
	generateBarcode *applicationsample.GenerateBarcodeUseCase
	stickerData     *applicationsample.StickerDataUseCase
}

func NewHandler(
	create *applicationsample.CreateSampleUseCase,
	list *applicationsample.ListSamplesUseCase,
	get *applicationsample.GetSampleUseCase,
	updateStatus *applicationsample.UpdateSampleStatusUseCase,
	listCoC *applicationsample.ListCoCStepsUseCase,
	appendCoC *applicationsample.AppendCoCStepUseCase,
	assignLocation *applicationsample.AssignLocationUseCase,
	generateBarcode *applicationsample.GenerateBarcodeUseCase,
	stickerData *applicationsample.StickerDataUseCase,
) *Handler {
	return &Handler{
		create:          create,
		list:            list,
		get:             get,
		updateStatus:    updateStatus,
		listCoC:         listCoC,
		appendCoC:       appendCoC,
		assignLocation:  assignLocation,
		generateBarcode: generateBarcode,
		stickerData:     stickerData,
	}
}

// Create godoc
//
//	@Summary		สร้างตัวอย่างใหม่
//	@Description	สร้าง sample record ใหม่พร้อมเริ่ม chain-of-custody
//	@Tags			samples
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateSampleRequest	true	"ข้อมูลตัวอย่าง"
//	@Success		201		{object}	response.Envelope{data=SampleResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/samples [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateSampleRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	s, err := h.create.Execute(c.Context(), applicationsample.CreateSampleInput{
		Name:            req.Name,
		Type:            sample.Type(req.Type),
		CustodianUserID: req.CustodianUserID,
		LocationID:      req.LocationID,
		BarcodeID:       req.BarcodeID,
		Description:     req.Description,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(s))
	return response.Created(c, toSampleResponse(s))
}

// List godoc
//
//	@Summary		รายการตัวอย่างทั้งหมด
//	@Description	รองรับ filter ผ่าน query param `status`, `type`, `barcode_id` (exact), `custodian_user_id`, `location` (บางส่วนของชื่อ Location)
//	@Tags			samples
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status				query	string	false	"กรองตามสถานะ"
//	@Param			type				query	string	false	"กรองตามประเภท"
//	@Param			barcode_id			query	string	false	"กรองด้วย Barcode ID (สแกน, ตรงเป๊ะ)"
//	@Param			custodian_user_id	query	integer	false	"กรองตามผู้ดูแล"
//	@Param			location			query	string	false	"กรองด้วยชื่อ Location (บางส่วน)"
//	@Param			location_id			query	string	false	"กรองด้วย Location ID ตรงเป๊ะ (เช่น ดูตัวอย่างทั้งหมดในกล่อง)"
//	@Success		200		{object}	response.Envelope{data=[]SampleResponse}
//	@Failure		401		{object}	response.Envelope
//	@Router			/samples [get]
func (h *Handler) List(c fiber.Ctx) error {
	var filter portsample.ListFilter
	if status := c.Query("status"); status != "" {
		s := sample.Status(status)
		filter.Status = &s
	}
	if typ := c.Query("type"); typ != "" {
		t := sample.Type(typ)
		filter.Type = &t
	}
	if barcodeID := c.Query("barcode_id"); barcodeID != "" {
		filter.BarcodeID = &barcodeID
	}
	if custodian := fiber.Query[int64](c, "custodian_user_id"); custodian != 0 {
		filter.CustodianUserID = &custodian
	}
	if loc := c.Query("location"); loc != "" {
		filter.LocationText = &loc
	}
	if locID := c.Query("location_id"); locID != "" {
		filter.LocationID = &locID
	}

	samples, err := h.list.Execute(c.Context(), filter)
	if err != nil {
		return err
	}
	out := make([]SampleResponse, len(samples))
	for i, s := range samples {
		out[i] = toSampleResponse(s)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลตัวอย่างตาม ID
//	@Tags			samples
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Sample ID"
//	@Success		200	{object}	response.Envelope{data=SampleResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/samples/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	s, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, toSampleResponse(s))
}

// UpdateStatus godoc
//
//	@Summary		เปลี่ยนสถานะตัวอย่าง
//	@Description	เปลี่ยนสถานะ sample และบันทึก chain-of-custody step ให้อัตโนมัติ
//	@Tags			samples
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Sample ID"
//	@Param			request	body		UpdateStatusRequest	true	"สถานะใหม่"
//	@Success		200		{object}	response.Envelope{data=SampleResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/samples/{id}/status [patch]
func (h *Handler) UpdateStatus(c fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	role := fiber.Locals[domainuser.Role](c, middleware.LocalsRole)
	name := fiber.Locals[string](c, middleware.LocalsName)

	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	s, err := h.updateStatus.Execute(c.Context(), applicationsample.UpdateSampleStatusInput{
		SampleID:  id,
		NewStatus: sample.Status(req.Status),
		ActorRole: role,
		ActorName: name,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, s))
	return response.OK(c, toSampleResponse(s))
}

// AssignLocation godoc
//
//	@Summary		กำหนดตำแหน่งจัดเก็บ (put-away) ของตัวอย่าง
//	@Description	ผูก sample เข้ากับ leaf Location - ต้องเป็น Location ที่ไม่มีลูก และไม่มี sample อื่นที่ยัง active ครองอยู่
//	@Tags			samples
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Sample ID"
//	@Param			request	body		AssignLocationRequest	true	"Location ID"
//	@Success		200		{object}	response.Envelope{data=SampleResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/samples/{id}/location [patch]
func (h *Handler) AssignLocation(c fiber.Ctx) error {
	id := c.Params("id")

	var req AssignLocationRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	hasID := req.LocationID != ""
	hasCode := req.LocationBarcodeCode != ""
	if hasID == hasCode {
		return fmt.Errorf("%w: exactly one of location_id or location_barcode_code is required", shared.ErrValidation)
	}

	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	var s sample.Sample
	if hasCode {
		s, err = h.assignLocation.ExecuteByBarcode(c.Context(), id, req.LocationBarcodeCode, req.Position)
	} else {
		s, err = h.assignLocation.Execute(c.Context(), id, req.LocationID, req.Position)
	}
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, s))
	return response.OK(c, toSampleResponse(s))
}

// GenerateBarcode godoc
//
//	@Summary		สร้าง Barcode ID ให้ตัวอย่าง
//	@Description	สร้าง Barcode ID อัตโนมัติ (SMP-BC-xxxxx) ถ้าตัวอย่างยังไม่มี; ถ้ามีอยู่แล้วคืนค่าเดิม
//	@Tags			samples
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Sample ID"
//	@Success		200	{object}	response.Envelope{data=SampleResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/samples/{id}/barcode [post]
func (h *Handler) GenerateBarcode(c fiber.Ctx) error {
	id := c.Params("id")

	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	s, err := h.generateBarcode.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, s))
	return response.OK(c, toSampleResponse(s))
}

// Sticker godoc
//
//	@Summary		พิมพ์สติกเกอร์บาร์โค้ดของตัวอย่าง (PDF)
//	@Description	เรนเดอร์ label PDF ขนาดตาม template (cap|stem|small|medium) พร้อมบาร์โค้ด code128 หรือ qr
//	@Tags			samples
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			id			path	string	true	"Sample ID"
//	@Param			template	query	string	false	"cap | stem | small | medium (ค่าเริ่มต้น medium)"
//	@Param			symbology	query	string	false	"code128 | qr"
//	@Success		200	{file}		byte
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/samples/{id}/sticker [get]
func (h *Handler) Sticker(c fiber.Ctx) error {
	id := c.Params("id")
	template := c.Query("template", pdf.StickerMedium)
	symbology := c.Query("symbology")

	data, err := h.stickerData.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	body, err := pdf.SampleSticker(pdf.SampleStickerData{
		ScanCode:         data.ScanCode,
		SampleID:         data.Sample.ID,
		Name:             data.Sample.Name,
		TypeLabel:        string(data.Sample.Type),
		CustodianName:    data.CustodianName,
		LocationFullPath: data.LocationFullPath,
		ReceivedAt:       data.Sample.ReceivedAt,
	}, template, symbology)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`inline; filename="%s-sticker.pdf"`, id))
	return c.Send(body)
}

// ListCoC godoc
//
//	@Summary		ประวัติ chain-of-custody ของตัวอย่าง
//	@Tags			samples
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Sample ID"
//	@Success		200	{object}	response.Envelope{data=[]CoCStepResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/samples/{id}/coc [get]
func (h *Handler) ListCoC(c fiber.Ctx) error {
	id := c.Params("id")
	steps, err := h.listCoC.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	out := make([]CoCStepResponse, len(steps))
	for i, s := range steps {
		out[i] = toCoCStepResponse(s)
	}
	return response.OK(c, out)
}

// AppendCoC godoc
//
//	@Summary		เพิ่ม chain-of-custody step ด้วยตนเอง
//	@Tags			samples
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Sample ID"
//	@Param			request	body		AppendCoCStepRequest	true	"ข้อมูล step"
//	@Success		201		{object}	response.Envelope{data=CoCStepResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/samples/{id}/coc [post]
func (h *Handler) AppendCoC(c fiber.Ctx) error {
	id := c.Params("id")

	var req AppendCoCStepRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	name := fiber.Locals[string](c, middleware.LocalsName)

	step, err := h.appendCoC.Execute(c.Context(), applicationsample.AppendCoCStepInput{
		SampleID: id,
		Title:    req.Title,
		Meta:     req.Meta,
		Who:      name,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditResource, "sample_coc_step")
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(step))
	return response.Created(c, toCoCStepResponse(step))
}
