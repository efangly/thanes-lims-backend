package document

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	docs := r.Group("/documents", authGuard)
	docs.Post("/", h.Upload)
	docs.Get("/", h.List)
	docs.Get("/:id", h.Get)
	docs.Get("/:id/download", h.Download)
	docs.Post("/:id/versions", h.NewVersion)
	docs.Get("/:id/history", h.History)
	docs.Patch("/:id/lock", h.SetLock)
}
