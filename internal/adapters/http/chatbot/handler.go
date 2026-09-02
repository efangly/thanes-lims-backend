package chatbot

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationchatbot "github.com/efangly/thanes-lims-backend/internal/application/chatbot"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	ask *applicationchatbot.AskUseCase
}

func NewHandler(ask *applicationchatbot.AskUseCase) *Handler {
	return &Handler{ask: ask}
}

// Ask godoc
//
//	@Summary		ถามข้อมูลห้องแล็บด้วยภาษาธรรมชาติ (POC)
//	@Description	ถาม-ตอบ single-turn จากข้อมูล Sample/TestResult/Inventory/PurchaseOrder บน Oracle ADB ผ่าน Claude API
//	@Tags			chatbot
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		ChatRequest	true	"คำถาม"
//	@Success		200	{object}	response.Envelope{data=ChatResponse}
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		403	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/chat [post]
func (h *Handler) Ask(c fiber.Ctx) error {
	var req ChatRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	ans, err := h.ask.Execute(c.Context(), applicationchatbot.AskInput{Question: req.Question})
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(ans))
}
