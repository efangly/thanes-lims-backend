//go:build integration

// Integration test for the MCP server (cmd/mcp-server, internal/adapters/mcp)
// per Phase 1.4 of the chatbot migration plan: stands up a real Postgres
// (via the same pgtest.SetupPostgres helper the rest of the repo's
// integration tests use), wires the real repositories, serves the MCP
// server over Streamable HTTP via httptest, and drives it through a real
// MCP client - exercising both auth checks (service key + JWT, including
// rejection cases) and one tool per domain end-to-end.
//
// Run with: go test -tags=integration ./internal/adapters/mcp/...
// (requires a running Docker daemon, same as every other *_integration_test.go
// in this repo).
package mcp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/jwt"
	limsmcp "github.com/efangly/thanes-lims-backend/internal/adapters/mcp"
	postgresinventory "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/inventory"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	postgrespurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/purchaseorder"
	postgressample "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	postgrestestresult "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/testresult"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	domaininventory "github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	domainpurchaseorder "github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	domaintestresult "github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// headerRoundTripper injects the service API key and (optionally) a bearer
// JWT on every outgoing request, simulating the NestJS gateway's forwarding
// contract (docs/adr/0013-mcp-server-transport.md).
type headerRoundTripper struct {
	serviceKey string
	bearer     string
}

func (h headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if h.serviceKey != "" {
		req.Header.Set("X-Service-Api-Key", h.serviceKey)
	}
	if h.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+h.bearer)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func seedIntegrationUser(t *testing.T, db *gorm.DB, name, email string) int64 {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO users (name, email, password_hash, role_id)
		 VALUES (?, ?, 'x', (SELECT id FROM roles WHERE name = 'Admin'))`, name, email,
	).Error)
	var id int64
	require.NoError(t, db.Raw(`SELECT id FROM users WHERE email = ?`, email).Scan(&id).Error)
	return id
}

func TestMCPServer_AuthAndToolsEndToEnd(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	ctx := context.Background()

	custodianID := seedIntegrationUser(t, db, "วิภา สายใจ", "wipa@test.local")

	sampleRepo := postgressample.New(db)
	sample, err := sampleRepo.Create(ctx, domainsample.Sample{
		ID:              "SMP-IT-0001",
		Name:            "Blood panel",
		Type:            domainsample.TypeBlood,
		CustodianUserID: custodianID,
		Status:          domainsample.StatusPending,
		ReceivedAt:      time.Now().AddDate(0, 0, -10),
	})
	require.NoError(t, err)

	trRepo := postgrestestresult.New(db)
	_, err = trRepo.Create(ctx, domaintestresult.TestResult{
		ID:       "TR-IT-0001",
		SampleID: sample.ID,
		TestName: "IgG",
		Analyst:  "Somchai",
		Result:   "high",
		Flag:     domaintestresult.FlagHi,
		RefRange: "0-10",
		Status:   domaintestresult.StatusApproved,
	})
	require.NoError(t, err)

	invRepo := postgresinventory.New(db)
	item, err := invRepo.Create(ctx, domaininventory.InventoryItem{
		ID:              "INV-IT-0001",
		Name:            "Glucose Reagent",
		Category:        "reagent",
		Unit:            "bottle",
		Min:             20,
		Max:             100,
		CustodianUserID: custodianID,
	})
	require.NoError(t, err)

	lotRepo := postgresinventory.NewLotRepository(db)
	_, err = lotRepo.Create(ctx, domaininventory.InventoryLot{ID: "LOT-IT-0001", ItemID: item.ID, LotNo: "A1", Quantity: 5})
	require.NoError(t, err)

	poRepo := postgrespurchaseorder.New(db)
	_, err = poRepo.Create(ctx, domainpurchaseorder.PurchaseOrder{
		ID:        "PO-IT-0001",
		ItemID:    item.ID,
		Quantity:  10,
		Vendor:    "Vendor A",
		OrderDate: time.Now(),
		Status:    domainpurchaseorder.StatusPendingApproval,
	})
	require.NoError(t, err)

	deps := limsmcp.Deps{
		Samples:        sampleRepo,
		Users:          postgresuser.New(db),
		TestResults:    trRepo,
		InventoryItems: invRepo,
		InventoryLots:  lotRepo,
		PurchaseOrders: poRepo,
	}
	server := limsmcp.NewServer(&mcp.Implementation{Name: "lims-mcp-server-it", Version: "test"}, deps)

	tokens := jwt.New("access-secret-it", "refresh-secret-it", 15*time.Minute, 168*time.Hour)
	handler := limsmcp.NewHTTPHandler(server, tokens, "the-service-key")
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	// Every domain permission the four happy-path tool calls below need, on
	// top of chatbot:view - see server.go's requirePermission (now checks
	// chatbot:view AND a per-tool domain permission, not chatbot:view alone).
	token, err := tokens.GenerateAccessToken(domainuser.User{ID: custodianID, Name: "วิภา สายใจ", Role: domainuser.RoleAdmin}, []string{
		"chatbot:view", "sample:view", "testresult:view", "inventory:view", "purchaseorder:view",
	})
	require.NoError(t, err)

	connect := func(serviceKey, bearer string) (*mcp.ClientSession, error) {
		client := mcp.NewClient(&mcp.Implementation{Name: "it-client", Version: "test"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint:   httpSrv.URL,
			HTTPClient: &http.Client{Transport: headerRoundTripper{serviceKey: serviceKey, bearer: bearer}},
		}
		return client.Connect(context.Background(), transport, nil)
	}

	t.Run("rejects missing service key", func(t *testing.T) {
		_, err := connect("", token)
		require.Error(t, err)
	})

	t.Run("rejects invalid JWT", func(t *testing.T) {
		_, err := connect("the-service-key", "garbage")
		require.Error(t, err)
	})

	t.Run("accepts valid credentials and exercises one tool per domain", func(t *testing.T) {
		session, err := connect("the-service-key", token)
		require.NoError(t, err)
		defer session.Close()

		sampleResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "getSampleById",
			Arguments: map[string]any{"id": "SMP-IT-0001"},
		})
		require.NoError(t, err)
		require.False(t, sampleResult.IsError, "getSampleById should succeed")

		trResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "searchTestResults",
			Arguments: map[string]any{"flag": "hi"},
		})
		require.NoError(t, err)
		require.False(t, trResult.IsError, "searchTestResults should succeed")

		invResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "getInventoryItemById",
			Arguments: map[string]any{"id": "INV-IT-0001"},
		})
		require.NoError(t, err)
		require.False(t, invResult.IsError, "getInventoryItemById should succeed")

		poResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "listPurchaseOrdersByItem",
			Arguments: map[string]any{"inventoryItemId": "INV-IT-0001"},
		})
		require.NoError(t, err)
		require.False(t, poResult.IsError, "listPurchaseOrdersByItem should succeed")
	})

	t.Run("rejects tool call for JWT without chatbot:view", func(t *testing.T) {
		noPermToken, err := tokens.GenerateAccessToken(domainuser.User{ID: custodianID, Name: "No Perm", Role: domainuser.RoleGeneral}, []string{})
		require.NoError(t, err)
		session, err := connect("the-service-key", noPermToken)
		require.NoError(t, err)
		defer session.Close()

		result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "listInventoryLowStock",
			Arguments: map[string]any{},
		})
		require.NoError(t, err)
		require.True(t, result.IsError, "tool call without chatbot:view must be rejected")
	})

	t.Run("rejects tool call for JWT with chatbot:view but missing the tool's domain permission", func(t *testing.T) {
		// Has chatbot:view (may use the AI assistant) but no inventory:view
		// (may not see Inventory data through the ordinary REST API either)
		// - per-tool RBAC must still deny this, not just check chatbot:view.
		partialToken, err := tokens.GenerateAccessToken(domainuser.User{ID: custodianID, Name: "Sample Only", Role: domainuser.RoleGeneral}, []string{
			"chatbot:view", "sample:view",
		})
		require.NoError(t, err)
		session, err := connect("the-service-key", partialToken)
		require.NoError(t, err)
		defer session.Close()

		result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "listInventoryLowStock",
			Arguments: map[string]any{},
		})
		require.NoError(t, err)
		require.True(t, result.IsError, "tool call without the domain-specific permission must be rejected even with chatbot:view")
	})
}
