package mcp

import (
	"log"

	portsuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// auditLog writes one line per tool-call attempt (allow or deny) to the
// standard logger, which `kubectl logs` already captures for cmd/mcp-server.
// This is the forensic trail requirePermission didn't have before: who
// (user_id/role) tried to call which tool, with what arguments, and whether
// it was allowed - see the "MCP server public" planning section in
// /Users/tng-mac-01/.claude/plans/ai-chatbot-groovy-spindle.md, which called
// this out as a gap worth closing regardless of whether the server ever
// actually becomes public. Deliberately NOT written to the audit_logs table
// (internal/domain/audit) - that table is a compliance record of *mutations*
// (ADR 0003, field-diff based), and every MCP tool call is read-only; mixing
// "who changed what" with "who asked the AI assistant for what" would make
// that table's semantics (and its PDF export) misleading.
//
func auditLog(req *mcp.CallToolRequest, outcome string, claims portsuser.Claims, reason string) {
	log.Printf(
		"mcp: audit tool=%s outcome=%s user_id=%d role=%s reason=%q args=%s",
		req.Params.Name, outcome, claims.UserID, claims.Role, reason, req.Params.Arguments,
	)
}
