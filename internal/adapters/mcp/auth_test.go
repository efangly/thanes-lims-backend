package mcp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/jwt"
	limsmcp "github.com/efangly/thanes-lims-backend/internal/adapters/mcp"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func newTestTokens() *jwt.Adapter {
	return jwt.New("access-secret", "refresh-secret", 15*time.Minute, 168*time.Hour)
}

func passthrough(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := limsmcp.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "no claims", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(claims.Name))
	})
}

func TestAuthMiddleware_RejectsMissingServiceKey(t *testing.T) {
	tokens := newTestTokens()
	handler := limsmcp.AuthMiddleware(passthrough(t), tokens, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer whatever")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_RejectsWrongServiceKey(t *testing.T) {
	tokens := newTestTokens()
	handler := limsmcp.AuthMiddleware(passthrough(t), tokens, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Service-Api-Key", "wrong")
	req.Header.Set("Authorization", "Bearer whatever")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_RejectsMissingJWT(t *testing.T) {
	tokens := newTestTokens()
	handler := limsmcp.AuthMiddleware(passthrough(t), tokens, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Service-Api-Key", "secret-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_RejectsInvalidJWT(t *testing.T) {
	tokens := newTestTokens()
	handler := limsmcp.AuthMiddleware(passthrough(t), tokens, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Service-Api-Key", "secret-key")
	req.Header.Set("Authorization", "Bearer not-a-real-jwt")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_AllowsValidServiceKeyAndJWT(t *testing.T) {
	tokens := newTestTokens()
	token, err := tokens.GenerateAccessToken(user.User{ID: 42, Name: "Somchai", Role: user.RoleScientist}, []string{"chatbot:view"})
	assert.NoError(t, err)

	handler := limsmcp.AuthMiddleware(passthrough(t), tokens, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Service-Api-Key", "secret-key")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Somchai", rec.Body.String())
}

func TestHasPermission(t *testing.T) {
	claims := portsuser.Claims{Permissions: []string{"sample:view", "chatbot:view"}}
	assert.True(t, limsmcp.HasPermission(claims, "chatbot:view"))
	assert.False(t, limsmcp.HasPermission(claims, "chatbot:edit"))
}

// TestNewServer_ToolRequiresPermission exercises the shared requirePermission
// wrapper (via NewServer's registration) end-to-end over an in-memory MCP
// transport: a session without chatbot:view in context must get an
// IsError tool result, not a Go panic or a protocol-level failure.
func TestNewServer_ToolRequiresPermission(t *testing.T) {
	server := limsmcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, limsmcp.Deps{})

	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "0.0.0"}, nil)
	t1, t2 := mcp.NewInMemoryTransports()

	ctxNoPerm := context.Background()
	if _, err := server.Connect(ctxNoPerm, t1, nil); err != nil {
		t.Fatal(err)
	}
	session, err := client.Connect(context.Background(), t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "listInventoryLowStock",
		Arguments: map[string]any{},
	})
	assert.NoError(t, err)
	assert.True(t, result.IsError, "tool call without chatbot:view claims in context must be rejected")
}

// TestNewServer_ToolRequiresDomainPermission exercises the newer half of
// requirePermission: chatbot:view alone is not enough, the caller must also
// carry the tool's specific domain permission (e.g. "inventory:view" for
// listInventoryLowStock). Goes through the real HTTP handler + AuthMiddleware
// (not an in-memory transport) since that's what actually stashes JWT claims
// into context - no Postgres/Docker needed, since limsmcp.Deps{} is never
// touched: the permission check rejects before the handler body would reach
// the (nil) repository.
func TestNewServer_ToolRequiresDomainPermission(t *testing.T) {
	server := limsmcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, limsmcp.Deps{})
	tokens := newTestTokens()
	handler := limsmcp.NewHTTPHandler(server, tokens, "secret-key")
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	// chatbot:view present, inventory:view absent.
	token, err := tokens.GenerateAccessToken(user.User{ID: 1, Name: "Sample Only", Role: user.RoleGeneral}, []string{"chatbot:view", "sample:view"})
	assert.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "0.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: httpSrv.URL,
		HTTPClient: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("X-Service-Api-Key", "secret-key")
			req.Header.Set("Authorization", "Bearer "+token)
			return http.DefaultTransport.RoundTrip(req)
		})},
	}
	session, err := client.Connect(context.Background(), transport, nil)
	assert.NoError(t, err)
	defer session.Close()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "listInventoryLowStock",
		Arguments: map[string]any{},
	})
	assert.NoError(t, err)
	assert.True(t, result.IsError, "tool call with chatbot:view but without inventory:view must still be rejected")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
