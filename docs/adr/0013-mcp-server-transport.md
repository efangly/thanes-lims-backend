# MCP server: SDK choice, transport, and credential-forwarding contract

## Context

`docs/chatbot-poc-plan.md` documents the original chatbot POC: a Go handler (`POST /chat`)
that used Claude's tool-use loop to write and run SQL against a separate Oracle ADB mirror
of Postgres. `/Users/tng-mac-01/.claude/plans/ai-chatbot-groovy-spindle.md` (Phase 0-2)
replaces that whole approach: the AI chatbot moves to an external NestJS + LangGraph.js
service, and this Go backend stops mirroring data to Oracle. Instead it exposes an MCP
(Model Context Protocol) server so the new service can query **Postgres directly** - the
real system of record - through the existing Sample/TestResult/Inventory/PurchaseOrder
repositories, never via new SQL.

This ADR is Phase 1 of that plan: it only adds the MCP server (`cmd/mcp-server`,
`internal/adapters/mcp`). The old chatbot code (`internal/adapters/oracle/chatbot`,
`internal/adapters/anthropic/chatbot`, `internal/adapters/http/chatbot`, `/api/v1/chat`) is
untouched and keeps running; it is only removed in a later phase, gated on the NestJS
service's cutover.

## Decision: `github.com/modelcontextprotocol/go-sdk`

We evaluated the official `github.com/modelcontextprotocol/go-sdk` (maintained with
Google/Anthropic) against the community `github.com/mark3labs/mcp-go`. We chose the
**official SDK** (`v1.8.0` at the time of writing):

- It's `go get`-able and builds cleanly against this module's Go toolchain (go 1.26).
- Its Streamable HTTP transport (`mcp.NewStreamableHTTPHandler`) is a first-class,
  actively maintained implementation of the current MCP spec transport - not a bolt-on -
  with production concerns already handled: session management, `Mcp-Session-Id`
  lifecycle, SSE resumption via `EventStore`, request body size limits, and DNS-rebinding
  protection for localhost binds. Nothing about it looked unstable or unsuitable; there
  was no need to fall back to `mark3labs/mcp-go`.
- `mcp.AddTool[In, Out]` infers a tool's JSON input/output schema from Go struct tags
  (`jsonschema:"..."` per field) via `github.com/google/jsonschema-go`, which is what we
  use to hand-write explicit, LLM-facing field descriptions per
  docs/mcp-server-tools.md, rather than relying on bare Go field names.
- It plugs into any `net/http` handler chain, which made it trivial to wrap with our own
  auth middleware (below) instead of fighting a framework-specific routing layer.

`NewStreamableHTTPHandler`'s `getServer` callback returns the same `*mcp.Server` for every
request in our case (this server is stateless w.r.t. which tools exist - only the caller's
identity varies per request), which is an explicitly supported pattern per its doc comment.

## Transport: Streamable HTTP, one server per process

Per the migration plan's Phase 0 decision, the MCP server speaks **Streamable HTTP**
(not stdio, not SSE-only) so the NestJS service can reach it as a normal internal HTTP
service on the same Kubernetes network as the rest of this backend. It runs as its own
binary/process (`cmd/mcp-server`, default port `8090` via `MCP_SERVER_PORT`), entirely
separate from `cmd/api` - it does not mount into the Fiber app or share its lifecycle, so
it can be deployed, scaled, and restarted independently of the main API.

## Credential-forwarding contract

The MCP server sits behind two independent, mandatory checks (`internal/adapters/mcp/auth.go`),
applied to *every* HTTP request before any MCP method (including `tools/list`) is served:

1. **Service API key** - header `X-Service-Api-Key`, compared with
   `crypto/subtle.ConstantTimeCompare` against `MCP_SERVICE_API_KEY`. This proves the
   request came from the NestJS gateway itself (a static, deployment-scoped secret), not
   from an arbitrary holder of a leaked end-user JWT.
2. **Forwarded end-user JWT** - the standard `Authorization: Bearer <jwt>` header, carrying
   the *same* access token the end user's browser already holds (issued by this backend's
   own `/api/v1/auth/login`, `internal/adapters/jwt`). The NestJS service is expected to
   simply forward the token it received from the frontend, unmodified. It is parsed with
   the same HS256 verification and `JWT_ACCESS_SECRET` as `internal/adapters/jwt.Adapter`,
   extracting `user_id`, `role`, and `permissions`.

Both must pass, or the request is rejected outright (HTTP 401) before reaching the MCP
session/message layer - this is transport-level rejection, appropriate for missing/invalid
*credentials*, matching how `internal/adapters/http/middleware.Auth` behaves for the main
API.

RBAC (`chatbot:view`, `rbac.ModuleChatbot`/`rbac.ActionView` - the same permission the old
`/chat` endpoint required, granted to every Role per
`docs/chatbot-acceptance-checklist.md`) is checked one level down, at MCP tool-dispatch
time: `internal/adapters/mcp/server.go`'s generic `requirePermission[In, Out]` wraps every
`mcp.AddTool` registration once, checking the claims stashed in context by the auth
middleware. A caller whose JWT lacks `chatbot:view` gets an MCP tool-level error
(`CallToolResult{IsError: true}`) rather than a transport error, so the calling LLM can see
why the call failed and react, per the SDK's guidance for handler-detected failures. This
keeps the RBAC check in one place instead of duplicated inside each `tools/*.go` handler.

## Consequences

- The NestJS service must hold both `MCP_SERVICE_API_KEY` and be able to forward the
  caller's real access token - it cannot mint its own service-level JWT with elevated
  claims, by design (no new trust boundary is introduced beyond what `/api/v1` already
  grants that user).
- All 8 MCP tools are read-only by construction (typed handlers over `FindByID`/`List`,
  never `Create`/`Update`/`Delete`) - there is no MCP-level "read-only guard" to bypass, per
  the acceptance checklist's read-only guarantee.
- `cmd/mcp-server` reuses `internal/config.Load()` like `cmd/seed` does, even though it
  doesn't need every field (Storage/Redis/Partner config) - consistent with the existing
  multi-binary pattern in this repo rather than splitting config per binary.
