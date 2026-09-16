// Package tools holds the MCP tool handlers exposed by the LIMS MCP server
// (internal/adapters/mcp). Each file covers one domain (Sample, TestResult,
// Inventory, PurchaseOrder) and is a thin adapter: validate input, call the
// existing repository interface already used by the rest of the backend,
// shape the response. No business logic lives here - see
// docs/mcp-server-tools.md for the tool catalog these types define the
// contract for.
package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errorResult builds an MCP tool-level error (IsError: true, a human-
// readable message in Content) rather than a protocol-level error, so the
// calling LLM sees the failure and can react (e.g. re-ask with a valid ID)
// instead of the call simply failing. See mcp.CallToolResult.IsError docs.
func errorResult[Out any](msg string) (*mcp.CallToolResult, Out, error) {
	var zero Out
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, zero, nil
}
