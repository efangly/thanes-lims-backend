// Package chatbot is the Oracle ADB adapter that runs LLM-proposed SELECT
// statements read-only for the chatbot POC. Defence in depth: the DB user
// (CHATBOT_RO) is granted SELECT only, and every statement is also validated
// and run inside a read-only transaction here.
package chatbot

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	// maxRows caps how many rows we hand back to the LLM - the demo questions
	// are all aggregates or short lists, and a big result set just wastes
	// context/tokens.
	maxRows = 200
	// queryTimeout bounds a single SELECT so a bad LLM query can't hang the
	// request.
	queryTimeout = 15 * time.Second
)

type Runner struct {
	db *sql.DB
}

func New(db *sql.DB) *Runner {
	return &Runner{db: db}
}

// forbiddenTokens catches DML/DDL and comment markers as whole words. The
// CHATBOT_RO grant already blocks writes at the DB; this keeps the failure
// early and legible (and blocks comment-smuggled multi-statements).
var forbiddenTokens = regexp.MustCompile(`(?i)\b(insert|update|delete|merge|drop|alter|create|grant|revoke|truncate|begin|declare|call|execute|commit|rollback)\b|--|/\*`)

// validateSelect enforces "one bare SELECT/WITH statement". Returns the
// cleaned query (trailing ';' removed) on success.
func validateSelect(query string) (string, error) {
	q := strings.TrimSpace(query)
	q = strings.TrimSuffix(q, ";")
	q = strings.TrimSpace(q)
	if q == "" {
		return "", fmt.Errorf("empty query")
	}
	if strings.Contains(q, ";") {
		return "", fmt.Errorf("only a single statement is allowed")
	}
	lower := strings.ToLower(q)
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return "", fmt.Errorf("only SELECT/WITH queries are allowed")
	}
	if forbiddenTokens.MatchString(q) {
		return "", fmt.Errorf("query contains a disallowed keyword or comment")
	}
	return q, nil
}

func (r *Runner) RunSelect(ctx context.Context, query string) ([]string, [][]string, error) {
	q, err := validateSelect(query)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	// Read-only transaction: a second guardrail behind the CHATBOT_RO grant.
	if _, err := conn.ExecContext(ctx, "SET TRANSACTION READ ONLY"); err != nil {
		return nil, nil, fmt.Errorf("set read only: %w", err)
	}

	rows, err := conn.QueryContext(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var out [][]string
	for rows.Next() {
		if len(out) >= maxRows {
			break
		}
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		rec := make([]string, len(cols))
		for i, v := range raw {
			rec[i] = stringify(v)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return cols, out, nil
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case time.Time:
		return t.Format(time.RFC3339)
	default:
		return fmt.Sprint(t)
	}
}
