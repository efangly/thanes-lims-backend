package mcp_test

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	limsmcp "github.com/efangly/thanes-lims-backend/internal/adapters/mcp"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// fakeInventoryRepo is the minimal portsinventory.Repository needed to let a
// permission-granted call to listInventoryLowStock actually reach the
// handler body (and so exercise the "allow" branch of the audit log) without
// a real database - only List is ever called by that tool.
type fakeInventoryRepo struct{}

func (fakeInventoryRepo) Create(context.Context, inventory.InventoryItem) (inventory.InventoryItem, error) {
	panic("not used by this test")
}
func (fakeInventoryRepo) FindByID(context.Context, string) (inventory.InventoryItem, error) {
	panic("not used by this test")
}
func (fakeInventoryRepo) List(context.Context) ([]inventory.InventoryItem, error) {
	return nil, nil
}
func (fakeInventoryRepo) Update(context.Context, inventory.InventoryItem) (inventory.InventoryItem, error) {
	panic("not used by this test")
}
func (fakeInventoryRepo) UpdateDefaultVendor(context.Context, string, string) (inventory.InventoryItem, error) {
	panic("not used by this test")
}

var _ portsinventory.Repository = fakeInventoryRepo{}

// captureLog redirects the standard logger's output for the duration of fn,
// restoring it afterward - log.Printf is a shared/global sink, so tests that
// assert on its output must not run in parallel with each other.
func captureLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)
	fn()
	return buf.String()
}

func TestRequirePermission_AuditLogsAllowAndDeny(t *testing.T) {
	server := limsmcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, limsmcp.Deps{InventoryItems: fakeInventoryRepo{}})
	tokens := newTestTokens()
	handler := limsmcp.NewHTTPHandler(server, tokens, "secret-key")
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	callTool := func(bearer, name string, args map[string]any) {
		client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "0.0.0"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint: httpSrv.URL,
			HTTPClient: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				req.Header.Set("X-Service-Api-Key", "secret-key")
				req.Header.Set("Authorization", "Bearer "+bearer)
				return http.DefaultTransport.RoundTrip(req)
			})},
		}
		session, err := client.Connect(context.Background(), transport, nil)
		assert.NoError(t, err)
		defer session.Close()
		_, _ = session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	}

	t.Run("allowed call logs outcome=allow with caller identity", func(t *testing.T) {
		token, err := tokens.GenerateAccessToken(user.User{ID: 7, Name: "Somchai", Role: user.RoleAdmin}, []string{"chatbot:view", "inventory:view"})
		assert.NoError(t, err)

		out := captureLog(t, func() {
			callTool(token, "listInventoryLowStock", map[string]any{})
		})

		assert.Contains(t, out, "mcp: audit")
		assert.Contains(t, out, "tool=listInventoryLowStock")
		assert.Contains(t, out, "outcome=allow")
		assert.Contains(t, out, "user_id=7")
		assert.True(t, strings.Contains(out, "role=admin"), "expected role=admin in: %s", out)
	})

	t.Run("denied call (missing domain permission) logs outcome=deny with reason", func(t *testing.T) {
		token, err := tokens.GenerateAccessToken(user.User{ID: 9, Name: "No Inventory", Role: user.RoleGeneral}, []string{"chatbot:view"})
		assert.NoError(t, err)

		out := captureLog(t, func() {
			callTool(token, "listInventoryLowStock", map[string]any{})
		})

		assert.Contains(t, out, "outcome=deny")
		assert.Contains(t, out, "user_id=9")
		assert.Contains(t, out, `reason="missing inventory:view"`)
	})
}
