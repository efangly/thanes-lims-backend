// Command mcp-server runs the LIMS MCP (Model Context Protocol) server: a
// standalone binary, entirely separate from cmd/api, that exposes read-only
// Sample/TestResult/Inventory/PurchaseOrder tools over MCP Streamable HTTP
// for the external NestJS+LangGraph.js chatbot service to call. It reads
// Postgres directly through the same repositories cmd/api wires up - see
// docs/adr/0013-mcp-server-transport.md and docs/mcp-server-tools.md.
//
// This binary intentionally does not import anything from
// internal/adapters/http/chatbot, internal/adapters/anthropic/chatbot, or
// internal/adapters/oracle - the old chatbot POC being replaced. It mounts
// no Fiber routes and has nothing to do with /api/v1/chat.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/jwt"
	limsmcp "github.com/efangly/thanes-lims-backend/internal/adapters/mcp"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	postgresinventory "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/inventory"
	postgrespurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/purchaseorder"
	postgressample "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	postgrestestresult "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/testresult"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.MCPServiceAPIKey == "" {
		log.Fatal("MCP_SERVICE_API_KEY must be set - refusing to start an unauthenticated MCP server")
	}

	gdb, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	tokens := jwt.New(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	deps := limsmcp.Deps{
		Samples:        postgressample.New(gdb),
		Users:          postgresuser.New(gdb),
		TestResults:    postgrestestresult.New(gdb),
		InventoryItems: postgresinventory.New(gdb),
		InventoryLots:  postgresinventory.NewLotRepository(gdb),
		PurchaseOrders: postgrespurchaseorder.New(gdb),
	}

	server := limsmcp.NewServer(&mcp.Implementation{Name: "lims-mcp-server", Version: "1.0.0"}, deps)
	handler := limsmcp.NewHTTPHandler(server, tokens, cfg.MCPServiceAPIKey)

	addr := ":" + cfg.MCPServerPort
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("mcp-server: listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("mcp-server: listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("mcp-server: shutdown: %v", err)
	}
}
