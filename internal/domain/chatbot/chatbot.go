// Package chatbot holds the domain types for the single-turn AI chatbot POC
// (see docs/chatbot-poc-plan.md). A Question is answered by an Assistant that
// translates it to SQL, runs that SQL read-only against the POC Oracle ADB,
// and narrates the rows back in Thai. This is a demo feature - the Postgres
// system of record is never touched.
package chatbot

// Question is one natural-language question (Thai or English) from a user.
type Question struct {
	Text string
}

// Answer is the narrated reply plus a bit of provenance the demo UI shows so
// stakeholders can see which SQL produced the answer and how fast.
type Answer struct {
	// Text is the Thai narration produced by the LLM from the query rows.
	Text string
	// SQLQueries are the SELECT statements actually executed, in order.
	SQLQueries []string
	// Rows is the total number of rows returned across all queries.
	Rows int
	// ElapsedMS is the wall-clock time for the whole ask (LLM + SQL).
	ElapsedMS int64
	// CacheReadTokens / CacheWriteTokens are the prompt-cache tokens served
	// from / written to Anthropic's cache across all LLM turns of this ask
	// (see docs/chatbot-poc-plan.md - the schema/system prompt is cached).
	CacheReadTokens  int64
	CacheWriteTokens int64
}
