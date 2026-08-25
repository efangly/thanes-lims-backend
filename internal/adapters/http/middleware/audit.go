package middleware

import (
	"reflect"
	"strings"
	"time"

	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/gofiber/fiber/v3"
)

const (
	// LocalsAuditResource lets a handler override the Resource the Audit
	// middleware derives from the request path - needed for sub-resources
	// that are audited distinctly from their parent (Sample CoC Step,
	// Calibration Event - see CONTEXT.md#audit-trail and ADR 0003).
	LocalsAuditResource localsKey = "audit_resource"
	// LocalsAuditChangeSet lets a handler attach a field-level Change Set
	// (or, for a create/append, a snapshot; for a delete, a marker) to the
	// Audit Entry the middleware writes for this request. Build it with
	// ChangeSet, Snapshot, or DeletedMarker below.
	LocalsAuditChangeSet localsKey = "audit_change_set"
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

		resource := resourceForPath(c.Path())
		if r, ok := c.Locals(LocalsAuditResource).(string); ok && r != "" {
			resource = r
		}

		var metadata map[string]any
		if cs, ok := c.Locals(LocalsAuditChangeSet).(map[string]any); ok {
			metadata = cs
		}

		entry := domainaudit.AuditLog{
			ActorID:    actorID,
			ActorRole:  actorRole,
			Method:     method,
			Path:       c.Path(),
			Resource:   resource,
			ResourceID: c.Params("id"),
			StatusCode: statusCode,
			Metadata:   metadata,
			CreatedAt:  time.Now(),
		}
		_ = logAction.Execute(c.Context(), entry)

		return handlerErr
	}
}

// resourceForPath derives an Audit Entry's default Resource from the
// request path, using rbac.Module string values where a Module maps 1:1
// onto an audited entity. Two sub-resources are audited distinctly from
// their parent per CONTEXT.md#audit-trail / ADR 0003 (Sample CoC Step,
// Calibration Event) and are special-cased here; everything else defers to
// the per-use-case override via LocalsAuditResource where a handler needs
// something more specific than the path implies. Routes outside the
// audited scope (environment, notification, auth) fall through to "".
func resourceForPath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	// segments[0] == "api", segments[1] == "v1", segments[2] == the resource collection.
	if len(segments) < 3 {
		return ""
	}

	collection := segments[2]
	switch collection {
	case "samples":
		if len(segments) >= 5 && segments[4] == "coc" {
			return "sample_coc_step"
		}
		return "sample"
	case "locations":
		return "location"
	case "users":
		return "user"
	case "tests":
		return "testresult"
	case "equipment":
		if len(segments) >= 5 && segments[4] == "calibration" {
			return "calibration_event"
		}
		return "equipment"
	case "inventory":
		return "inventory"
	case "purchase-orders":
		return "purchaseorder"
	case "documents":
		return "document"
	default:
		return ""
	}
}

// ChangeSet diffs the exported fields of before and after (which must be
// the same struct type, passed by value) and returns a Change Set: for
// every field whose value differs, {"old": ..., "new": ...} - see
// CONTEXT.md#audit-trail. Fields that didn't change are omitted.
func ChangeSet(before, after any) map[string]any {
	cs := map[string]any{}

	bv := reflect.ValueOf(before)
	av := reflect.ValueOf(after)
	if bv.Kind() != reflect.Struct || av.Kind() != reflect.Struct || bv.Type() != av.Type() {
		return cs
	}

	t := bv.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		bf := bv.Field(i).Interface()
		af := av.Field(i).Interface()
		if reflect.DeepEqual(bf, af) {
			continue
		}
		cs[field.Name] = map[string]any{"old": bf, "new": af}
	}
	return cs
}

// Snapshot returns the exported fields of v as a plain map, for the
// Change Set of a create or an append (there is no "old" state to diff
// against - see CONTEXT.md#audit-trail).
func Snapshot(v any) map[string]any {
	out := map[string]any{}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return out
	}

	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		out[field.Name] = rv.Field(i).Interface()
	}
	return out
}

// DeletedMarker is the Change Set for a delete (soft delete / Retired -
// see ADR 0003): a simple marker rather than a full pre-delete snapshot,
// kept consistent across every audited Module.
func DeletedMarker() map[string]any {
	return map[string]any{"deleted": true}
}
