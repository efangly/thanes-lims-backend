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

// List godoc
//
//	@Summary		รายการแจ้งเตือนของผู้ใช้ปัจจุบัน
//	@Tags			notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]NotificationResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/notifications [get]
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

// MarkRead godoc
//
//	@Summary		ทำเครื่องหมายว่าอ่านแล้ว (รายการเดียว)
//	@Tags			notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Notification ID"
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/notifications/{id}/read [patch]
func (h *Handler) MarkRead(c fiber.Ctx) error {
	if err := h.markRead.Execute(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"read": true})
}

// MarkAllRead godoc
//
//	@Summary		ทำเครื่องหมายว่าอ่านแล้วทั้งหมด
//	@Tags			notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Router			/notifications/read-all [patch]
func (h *Handler) MarkAllRead(c fiber.Ctx) error {
	userID := fiber.Locals[int64](c, middleware.LocalsUserID)
	if err := h.markAllRead.Execute(c.Context(), userID); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"read": true})
}
