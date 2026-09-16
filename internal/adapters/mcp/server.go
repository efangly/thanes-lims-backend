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

// requiredPermission is the single RBAC gate for every tool this server
// exposes - "chatbot:view" (rbac.ModuleChatbot/rbac.ActionView), granted to
// every Role per the acceptance checklist's RBAC decision. There is no
// finer-grained per-domain permission: the old /api/v1/chat endpoint used
// the same single gate.
var requiredPermission = rbac.Permission{Module: rbac.ModuleChatbot, Action: rbac.ActionView}.Key()

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
// AuthMiddleware) carry "chatbot:view" before the handler body ever runs -
// so the RBAC check is written once here, not duplicated in each tools/*.go
// handler.
func NewServer(impl *mcp.Implementation, deps Deps) *mcp.Server {
	server := mcp.NewServer(impl, nil)

	sampleTools := &mcptools.SampleTools{Samples: deps.Samples, Users: deps.Users}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getSampleById",
		Description: "Look up a single Sample by its ID. Returns full sample detail including status, custodian, and location.",
	}, requirePermission(sampleTools.GetByID))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listSamplesByStatus",
		Description: "List Samples in a given lifecycle status, optionally filtered to those older than N days since receipt (e.g. pending samples stuck longer than a week).",
	}, requirePermission(sampleTools.ListByStatus))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listSamplesByCustodianName",
		Description: "List all Samples whose custodian (the user responsible for them) matches the given full name.",
	}, requirePermission(sampleTools.ListByCustodianName))

	testResultTools := &mcptools.TestResultTools{Results: deps.TestResults}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "searchTestResults",
		Description: "Search Test Results, optionally filtered by Sample ID, result status, and/or abnormality flag (hi/lo/ok).",
	}, requirePermission(testResultTools.Search))

	inventoryTools := &mcptools.InventoryTools{Items: deps.InventoryItems, Lots: deps.InventoryLots}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listInventoryLowStock",
		Description: "List Inventory Items whose on-hand quantity is at or below their configured minimum quantity - i.e. items due for reorder.",
	}, requirePermission(inventoryTools.ListLowStock))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getInventoryItemById",
		Description: "Look up a single Inventory Item by its ID, including its individual received lots (batches) when available.",
	}, requirePermission(inventoryTools.GetByID))

	poTools := &mcptools.PurchaseOrderTools{Orders: deps.PurchaseOrders}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPurchaseOrders",
		Description: "List Purchase Orders, optionally filtered by status (pending_approval, sent_to_vendor, received, cancelled).",
	}, requirePermission(poTools.List))
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPurchaseOrdersByItem",
		Description: "List all Purchase Orders that order a given Inventory Item.",
	}, requirePermission(poTools.ListByItem))

	return server
}

// requirePermission wraps a typed MCP tool handler so it only runs when the
// context carries JWT claims (set by AuthMiddleware from the forwarded
// end-user access token) that include "chatbot:view". Applied once per
// mcp.AddTool call above - the single shared dispatch wrapper the RBAC check
// lives in, instead of being duplicated inside each tools/*.go handler body.
func requirePermission[In, Out any](h mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		var zero Out
		claims, ok := ClaimsFromContext(ctx)
		if !ok || !HasPermission(claims, requiredPermission) {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "permission denied: missing " + requiredPermission}},
			}, zero, nil
		}
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
