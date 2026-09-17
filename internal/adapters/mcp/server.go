package mcp

import (
	"context"
	"net/http"

	mcptools "github.com/efangly/thanes-lims-backend/internal/adapters/mcp/tools"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portsinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portspurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
	portssample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	portstestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	portsuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// chatbotViewPermission is checked on every tool regardless of domain - it's
// the "this user may use the AI assistant at all" switch, kept separate from
// the per-domain permissions below so a Role's AI access can be revoked
// without touching its ordinary CRUD permissions (or vice versa).
//
// Originally this was the *only* gate (every tool just required
// "chatbot:view", matching the old /api/v1/chat endpoint's single check).
// Per-domain permissions were added on top once the acceptance checklist's
// "all roles get chatbot:view" decision stopped being sufficient on its own -
// see the "MCP server public" planning section in the migration plan
// (/Users/tng-mac-01/.claude/plans/ai-chatbot-groovy-spindle.md) for why:
// a caller should only be able to pull Sample data through the chatbot if
// their Role can already see Sample data through the ordinary UI/API.
var chatbotViewPermission = rbac.Permission{Module: rbac.ModuleChatbot, Action: rbac.ActionView}.Key()

// Per-domain permissions, checked in addition to chatbotViewPermission -
// same Module/Action pairs the rest of the backend already enforces on the
// regular REST endpoints (internal/domain/rbac/permission.go), reused
// as-is: no new permission or migration was needed, every user's JWT
// already carries these.
var (
	samplePermission        = rbac.Permission{Module: rbac.ModuleSample, Action: rbac.ActionView}.Key()
	testResultPermission    = rbac.Permission{Module: rbac.ModuleTestResult, Action: rbac.ActionView}.Key()
	inventoryPermission     = rbac.Permission{Module: rbac.ModuleInventory, Action: rbac.ActionView}.Key()
	purchaseOrderPermission = rbac.Permission{Module: rbac.ModulePurchaseOrder, Action: rbac.ActionView}.Key()
)

// Deps bundles the repositories the MCP tool handlers are constructed
// from - the same repository interfaces the rest of the backend already
// uses, wired against the real Postgres connection by cmd/mcp-server/main.go.
type Deps struct {
	Samples        portssample.SampleRepository
	Users          portsuser.UserRepository
	TestResults    portstestresult.Repository
	InventoryItems portsinventory.Repository
	InventoryLots  portsinventory.LotRepository
	PurchaseOrders portspurchaseorder.Repository
}

// NewServer builds the MCP server and registers every LIMS tool. Every tool
// is registered through requirePermission, a single shared wrapper (below)
// that checks the caller's forwarded JWT claims (stashed in context by
// AuthMiddleware) carry both "chatbot:view" AND the tool's specific
// per-domain permission before the handler body ever runs - so the RBAC
// check is written once here, not duplicated in each tools/*.go handler.
func NewServer(impl *mcp.Implementation, deps Deps) *mcp.Server {
	server := mcp.NewServer(impl, nil)

	sampleTools := &mcptools.SampleTools{Samples: deps.Samples, Users: deps.Users}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getSampleById",
		Description: "Look up a single Sample by its ID. Returns full sample detail including status, custodian, and location.",
	}, requirePermission(sampleTools.GetByID, samplePermission))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listSamplesByStatus",
		Description: "List Samples in a given lifecycle status, optionally filtered to those older than N days since receipt (e.g. pending samples stuck longer than a week).",
	}, requirePermission(sampleTools.ListByStatus, samplePermission))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listSamplesByCustodianName",
		Description: "List all Samples whose custodian (the user responsible for them) matches the given full name.",
	}, requirePermission(sampleTools.ListByCustodianName, samplePermission))

	testResultTools := &mcptools.TestResultTools{Results: deps.TestResults}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "searchTestResults",
		Description: "Search Test Results, optionally filtered by Sample ID, result status, and/or abnormality flag (hi/lo/ok).",
	}, requirePermission(testResultTools.Search, testResultPermission))

	inventoryTools := &mcptools.InventoryTools{Items: deps.InventoryItems, Lots: deps.InventoryLots}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listInventoryLowStock",
		Description: "List Inventory Items whose on-hand quantity is at or below their configured minimum quantity - i.e. items due for reorder.",
	}, requirePermission(inventoryTools.ListLowStock, inventoryPermission))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getInventoryItemById",
		Description: "Look up a single Inventory Item by its ID, including its individual received lots (batches) when available.",
	}, requirePermission(inventoryTools.GetByID, inventoryPermission))

	poTools := &mcptools.PurchaseOrderTools{Orders: deps.PurchaseOrders}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPurchaseOrders",
		Description: "List Purchase Orders, optionally filtered by status (pending_approval, sent_to_vendor, received, cancelled).",
	}, requirePermission(poTools.List, purchaseOrderPermission))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPurchaseOrdersByItem",
		Description: "List all Purchase Orders that order a given Inventory Item.",
	}, requirePermission(poTools.ListByItem, purchaseOrderPermission))

	return server
}

// requirePermission wraps a typed MCP tool handler so it only runs when the
// context carries JWT claims (set by AuthMiddleware from the forwarded
// end-user access token) that include BOTH "chatbot:view" AND the given
// domainPermission (e.g. "sample:view") - a caller must be allowed to use
// the AI assistant at all, and be allowed to see this specific domain's
// data through the ordinary REST API, before the tool body ever runs.
// Applied once per mcp.AddTool call above - the single shared dispatch
// wrapper the RBAC check lives in, instead of being duplicated inside each
// tools/*.go handler body.
func requirePermission[In, Out any](h mcp.ToolHandlerFor[In, Out], domainPermission string) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		var zero Out
		claims, ok := ClaimsFromContext(ctx)
		if !ok || !HasPermission(claims, chatbotViewPermission) {
			auditLog(req, "deny", claims, "missing "+chatbotViewPermission)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "permission denied: missing " + chatbotViewPermission}},
			}, zero, nil
		}
		if !HasPermission(claims, domainPermission) {
			auditLog(req, "deny", claims, "missing "+domainPermission)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "permission denied: missing " + domainPermission}},
			}, zero, nil
		}
		auditLog(req, "allow", claims, "")
		return h(ctx, req, in)
	}
}

// NewHTTPHandler mounts server behind the Streamable HTTP transport, wrapped
// in AuthMiddleware so both the service API key and the forwarded end-user
// JWT are validated on every request before any MCP method (including
// tools/list) is served.
func NewHTTPHandler(server *mcp.Server, tokens portsuser.TokenService, serviceAPIKey string) http.Handler {
	streamable := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	return AuthMiddleware(streamable, tokens, serviceAPIKey)
}
