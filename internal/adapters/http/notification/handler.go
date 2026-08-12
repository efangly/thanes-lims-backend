package notification

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	applicationnotification "github.com/efangly/thanes-lims-backend/internal/application/notification"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	list        *applicationnotification.ListNotificationsUseCase
	markRead    *applicationnotification.MarkReadUseCase
	markAllRead *applicationnotification.MarkAllReadUseCase
}

func NewHandler(
	list *applicationnotification.ListNotificationsUseCase,
	markRead *applicationnotification.MarkReadUseCase,
	markAllRead *applicationnotification.MarkAllReadUseCase,
) *Handler {
	return &Handler{list: list, markRead: markRead, markAllRead: markAllRead}
}

func (h *Handler) List(c fiber.Ctx) error {
	userID := fiber.Locals[int64](c, middleware.LocalsUserID)
	items, err := h.list.Execute(c.Context(), userID)
	if err != nil {
		return err
	}
	out := make([]NotificationResponse, len(items))
	for i, n := range items {
		out[i] = toResponse(n)
	}
	return response.OK(c, out)
}

func (h *Handler) MarkRead(c fiber.Ctx) error {
	if err := h.markRead.Execute(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"read": true})
}

func (h *Handler) MarkAllRead(c fiber.Ctx) error {
	userID := fiber.Locals[int64](c, middleware.LocalsUserID)
	if err := h.markAllRead.Execute(c.Context(), userID); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"read": true})
}
