package document

import (
	"strconv"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	upload      *applicationdocument.UploadDocumentUseCase
	newVersion  *applicationdocument.CreateNewVersionUseCase
	setLock     *applicationdocument.SetLockUseCase
	list        *applicationdocument.ListDocumentsUseCase
	get         *applicationdocument.GetDocumentUseCase
	downloadURL *applicationdocument.GetDownloadURLUseCase
	history     *applicationdocument.ListHistoryUseCase
}

func NewHandler(
	upload *applicationdocument.UploadDocumentUseCase,
	newVersion *applicationdocument.CreateNewVersionUseCase,
	setLock *applicationdocument.SetLockUseCase,
	list *applicationdocument.ListDocumentsUseCase,
	get *applicationdocument.GetDocumentUseCase,
	downloadURL *applicationdocument.GetDownloadURLUseCase,
	history *applicationdocument.ListHistoryUseCase,
) *Handler {
	return &Handler{upload: upload, newVersion: newVersion, setLock: setLock, list: list, get: get, downloadURL: downloadURL, history: history}
}

// Upload godoc
//
//	@Summary		อัปโหลดเอกสารใหม่
//	@Tags			documents
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file			formData	file	true	"ไฟล์เอกสาร"
//	@Param			name			formData	string	true	"ชื่อเอกสาร"
//	@Param			type			formData	string	true	"ประเภทเอกสาร"
//	@Param			access_level	formData	string	true	"ระดับการเข้าถึง"
//	@Param			equipment_id	formData	string	false	"ผูกกับอุปกรณ์ (optional)"
//	@Param			calibration_event_id	formData	integer	false	"ผูกกับรายการสอบเทียบ (optional, สำหรับใบรับรอง)"
//	@Success		201				{object}	response.Envelope{data=DocumentResponse}
//	@Failure		400				{object}	response.Envelope
//	@Failure		401				{object}	response.Envelope
//	@Router			/documents [post]
func (h *Handler) Upload(c fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	name := c.FormValue("name")
	docType := c.FormValue("type")
	accessLevel := c.FormValue("access_level")
	uploaderName := fiber.Locals[string](c, middleware.LocalsName)

	d, err := h.upload.Execute(c.Context(), applicationdocument.UploadDocumentInput{
		Name:        name,
		Type:        document.Type(docType),
		Filename:    fh.Filename,
		ContentType: fh.Header.Get("Content-Type"),
		Size:        fh.Size,
		Content:     f,
		AccessLevel: accessLevel,
		UploadedBy:  uploaderName,
		EquipmentID: c.FormValue("equipment_id"),
		CalibrationEventID: func() int64 {
			n, _ := strconv.ParseInt(c.FormValue("calibration_event_id"), 10, 64)
			return n
		}(),
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(d))
	return response.Created(c, toResponse(d))
}

// List godoc
//
//	@Summary		รายการเอกสารทั้งหมด
//	@Tags			documents
//	@Produce		json
//	@Security		BearerAuth
//	@Param			equipment_id	query	string	false	"กรองเฉพาะเอกสารที่ผูกกับอุปกรณ์นี้"
//	@Param			calibration_event_id	query	integer	false	"กรองเฉพาะเอกสารที่ผูกกับรายการสอบเทียบนี้"
//	@Success		200	{object}	response.Envelope{data=[]DocumentResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/documents [get]
func (h *Handler) List(c fiber.Ctx) error {
	var (
		docs []document.Document
		err  error
	)
	if equipmentID := c.Query("equipment_id"); equipmentID != "" {
		docs, err = h.list.ExecuteByEquipment(c.Context(), equipmentID)
	} else if cev := c.Query("calibration_event_id"); cev != "" {
		var id int64
		id, err = strconv.ParseInt(cev, 10, 64)
		if err == nil {
			docs, err = h.list.ExecuteByCalibrationEvent(c.Context(), id)
		}
	} else {
		docs, err = h.list.Execute(c.Context())
	}
	if err != nil {
		return err
	}
	out := make([]DocumentResponse, len(docs))
	for i, d := range docs {
		out[i] = toResponse(d)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลเอกสารตาม ID
//	@Tags			documents
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	response.Envelope{data=DocumentResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/documents/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	d, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(d))
}

// Download godoc
//
//	@Summary		ขอลิงก์ดาวน์โหลดเอกสาร
//	@Description	คืน pre-signed URL สำหรับดาวน์โหลดไฟล์จาก object storage
//	@Tags			documents
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/documents/{id}/download [get]
func (h *Handler) Download(c fiber.Ctx) error {
	url, err := h.downloadURL.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"url": url})
}

// NewVersion godoc
//
//	@Summary		อัปโหลดเวอร์ชันใหม่ของเอกสาร
//	@Tags			documents
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Document ID"
//	@Param			file		formData	file	true	"ไฟล์เวอร์ชันใหม่"
//	@Param			change_note	formData	string	false	"หมายเหตุการเปลี่ยนแปลง"
//	@Success		201			{object}	response.Envelope{data=DocumentResponse}
//	@Failure		400			{object}	response.Envelope
//	@Failure		401			{object}	response.Envelope
//	@Failure		404			{object}	response.Envelope
//	@Router			/documents/{id}/versions [post]
func (h *Handler) NewVersion(c fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	uploaderName := fiber.Locals[string](c, middleware.LocalsName)

	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	d, err := h.newVersion.Execute(c.Context(), applicationdocument.CreateNewVersionInput{
		DocumentID:  id,
		Filename:    fh.Filename,
		ContentType: fh.Header.Get("Content-Type"),
		Size:        fh.Size,
		Content:     f,
		ChangeNote:  c.FormValue("change_note"),
		UploadedBy:  uploaderName,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, d))
	return response.Created(c, toResponse(d))
}

// History godoc
//
//	@Summary		ประวัติการแก้ไขเอกสาร
//	@Tags			documents
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	response.Envelope{data=[]HistoryResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/documents/{id}/history [get]
func (h *Handler) History(c fiber.Ctx) error {
	items, err := h.history.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	out := make([]HistoryResponse, len(items))
	for i, item := range items {
		out[i] = toHistoryResponse(item)
	}
	return response.OK(c, out)
}

// SetLock godoc
//
//	@Summary		ล็อค/ปลดล็อคเอกสาร
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string		true	"Document ID"
//	@Param			request	body		LockRequest	true	"สถานะล็อค"
//	@Success		200		{object}	response.Envelope{data=DocumentResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		403		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/documents/{id}/lock [patch]
func (h *Handler) SetLock(c fiber.Ctx) error {
	var req LockRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	role := fiber.Locals[domainuser.Role](c, middleware.LocalsRole)
	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	d, err := h.setLock.Execute(c.Context(), id, req.Locked, role)
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, d))
	return response.OK(c, toResponse(d))
}
