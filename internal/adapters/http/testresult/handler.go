package testresult

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create  *applicationtestresult.CreateTestResultUseCase
	submit  *applicationtestresult.SubmitResultUseCase
	approve *applicationtestresult.ApproveResultUseCase
	list    *applicationtestresult.ListTestResultsUseCase
	get     *applicationtestresult.GetTestResultUseCase
}

func NewHandler(
	create *applicationtestresult.CreateTestResultUseCase,
	submit *applicationtestresult.SubmitResultUseCase,
	approve *applicationtestresult.ApproveResultUseCase,
	list *applicationtestresult.ListTestResultsUseCase,
	get *applicationtestresult.GetTestResultUseCase,
) *Handler {
	return &Handler{create: create, submit: submit, approve: approve, list: list, get: get}
}

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

// ListBySample serves the nested GET /samples/:id/tests convenience route.
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

func (h *Handler) Get(c fiber.Ctx) error {
	t, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(t))
}

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
