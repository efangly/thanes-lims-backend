// Package chatbot defines the ports the chatbot use case depends on: a
// read-only SQL runner against the POC Oracle ADB, and the LLM-backed
// Assistant that drives the NL->SQL->narrate loop.
package chatbot

import "context"

// SQLRunner executes a single read-only SELECT statement against the POC
// Oracle ADB and returns the result set as strings (every value stringified,
// NULL as ""). Implementations must reject anything that is not a lone
// SELECT/WITH query - the LLM proposes the SQL, so this is a trust boundary.
type SQLRunner interface {
	RunSelect(ctx context.Context, query string) (columns []string, rows [][]string, err error)
}
