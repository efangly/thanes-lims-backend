package partnerdevice

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// liveReadsEnabled gates /discover and /:serial/timeseries: both call the
// Partner API live (unlike every other route here, which is Postgres/Redis
// only), so they have nothing to serve without a live gRPC connection -
// only mounted when PARTNER_API_ENABLED=true (h.discover/h.timeseries are
// non-nil in that case - see NewHandler).
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService, liveReadsEnabled bool) {
	authGuard := middleware.Auth(tokens)
	requireView := middleware.RequirePermission(rbac.ModulePartnerDevice, rbac.ActionView)

	// middleware.AuthQuery is deliberately hard-restricted to genuine
	// WebSocket upgrade handshakes (see its doc comment) - an SSE request is
	// a plain GET, so /stream uses ordinary header Bearer auth like every
	// other route here. The frontend can't use a native EventSource against
	// this endpoint for the same reason (no custom header support); it
	// reads the stream via fetch + ReadableStream instead, which can set
	// Authorization like any other apiFetch call.
	pd := r.Group("/partner-devices", authGuard)
	pd.Post("/", middleware.RequirePermission(rbac.ModulePartnerDevice, rbac.ActionCreate), h.Create)
	pd.Get("/", requireView, h.List)
	// Static "/stream" and "/discover" are registered before the "/:serial"
	// wildcard so neither can be swallowed as a serial value.
	pd.Get("/stream", requireView, h.Stream)
	if liveReadsEnabled {
		pd.Get("/discover", requireView, h.Discover)
	}
	pd.Get("/:serial", requireView, h.Get)
	pd.Patch("/:serial", middleware.RequirePermission(rbac.ModulePartnerDevice, rbac.ActionEdit), h.Update)
	pd.Get("/:serial/snapshot", requireView, h.GetSnapshot)
	if liveReadsEnabled {
		pd.Get("/:serial/timeseries", requireView, h.GetTimeseries)
	}
}
