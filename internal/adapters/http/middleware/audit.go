package middleware

import (
	"time"

	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/gofiber/fiber/v3"
)

// Audit records POST/PATCH/DELETE requests to the audit_logs table after the
// handler completes. Logging failures are swallowed - auditing must never
// break the request it's observing.
func Audit(logAction *applicationaudit.LogActionUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		handlerErr := c.Next()

		method := c.Method()
		if method != fiber.MethodPost && method != fiber.MethodPatch && method != fiber.MethodDelete {
			return handlerErr
		}

		var actorID *int64
		if v, ok := c.Locals(LocalsUserID).(int64); ok {
			actorID = &v
		}
		var actorRole string
		if r, ok := c.Locals(LocalsRole).(domainuser.Role); ok {
			actorRole = string(r)
		}

		// fiber.Config.ErrorHandler hasn't run yet at this point (it only
		// runs after the whole middleware chain, including this one,
		// unwinds), so c.Response().StatusCode() would still read the
		// pre-error default for a failed request - derive the real status
		// from the returned error instead.
		statusCode := c.Response().StatusCode()
		if handlerErr != nil {
			statusCode, _ = StatusAndCodeForError(handlerErr)
		}

		entry := domainaudit.AuditLog{
			ActorID:    actorID,
			ActorRole:  actorRole,
			Method:     method,
			Path:       c.Path(),
			ResourceID: c.Params("id"),
			StatusCode: statusCode,
			CreatedAt:  time.Now(),
		}
		_ = logAction.Execute(c.Context(), entry)

		return handlerErr
	}
}
