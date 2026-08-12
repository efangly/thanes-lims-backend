package document

import (
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
	})
	if err != nil {
		return err
	}
	return response.Created(c, toResponse(d))
}

func (h *Handler) List(c fiber.Ctx) error {
	docs, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]DocumentResponse, len(docs))
	for i, d := range docs {
		out[i] = toResponse(d)
	}
	return response.OK(c, out)
}

func (h *Handler) Get(c fiber.Ctx) error {
	d, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(d))
}

func (h *Handler) Download(c fiber.Ctx) error {
	url, err := h.downloadURL.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"url": url})
}

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

	d, err := h.newVersion.Execute(c.Context(), applicationdocument.CreateNewVersionInput{
		DocumentID:  c.Params("id"),
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
	return response.Created(c, toResponse(d))
}

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

func (h *Handler) SetLock(c fiber.Ctx) error {
	var req LockRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	role := fiber.Locals[domainuser.Role](c, middleware.LocalsRole)
	d, err := h.setLock.Execute(c.Context(), c.Params("id"), req.Locked, role)
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(d))
}
