package sample

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create       *applicationsample.CreateSampleUseCase
	list         *applicationsample.ListSamplesUseCase
	get          *applicationsample.GetSampleUseCase
	updateStatus *applicationsample.UpdateSampleStatusUseCase
	listCoC      *applicationsample.ListCoCStepsUseCase
	appendCoC    *applicationsample.AppendCoCStepUseCase
}

func NewHandler(
	create *applicationsample.CreateSampleUseCase,
	list *applicationsample.ListSamplesUseCase,
	get *applicationsample.GetSampleUseCase,
	updateStatus *applicationsample.UpdateSampleStatusUseCase,
	listCoC *applicationsample.ListCoCStepsUseCase,
	appendCoC *applicationsample.AppendCoCStepUseCase,
) *Handler {
	return &Handler{create: create, list: list, get: get, updateStatus: updateStatus, listCoC: listCoC, appendCoC: appendCoC}
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateSampleRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	s, err := h.create.Execute(c.Context(), applicationsample.CreateSampleInput{
		Name:      req.Name,
		Type:      sample.Type(req.Type),
		Custodian: req.Custodian,
		Location:  req.Location,
	})
	if err != nil {
		return err
	}
	return response.Created(c, toSampleResponse(s))
}

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

func (h *Handler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	s, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, toSampleResponse(s))
}

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

	s, err := h.updateStatus.Execute(c.Context(), applicationsample.UpdateSampleStatusInput{
		SampleID:  id,
		NewStatus: sample.Status(req.Status),
		ActorRole: role,
		ActorName: name,
	})
	if err != nil {
		return err
	}
	return response.OK(c, toSampleResponse(s))
}

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
	return response.Created(c, toCoCStepResponse(step))
}
