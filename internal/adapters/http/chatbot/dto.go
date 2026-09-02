package chatbot

import domainchatbot "github.com/efangly/thanes-lims-backend/internal/domain/chatbot"

// ChatRequest is the single-turn question body for POST /chat.
type ChatRequest struct {
	Question string `json:"question" validate:"required,max=500"`
}

// ChatResponse carries the Thai answer plus the SQL that produced it and a
// rough latency figure - the demo UI shows both so stakeholders can see the
// chatbot is answering from real data.
type ChatResponse struct {
	Answer           string   `json:"answer"`
	SQLQueries       []string `json:"sql_queries"`
	Rows             int      `json:"rows"`
	ElapsedMs        int64    `json:"elapsed_ms"`
	CacheReadTokens  int64    `json:"cache_read_tokens"`
	CacheWriteTokens int64    `json:"cache_write_tokens"`
}

func toResponse(a domainchatbot.Answer) ChatResponse {
	return ChatResponse{
		Answer:           a.Text,
		SQLQueries:       a.SQLQueries,
		Rows:             a.Rows,
		ElapsedMs:        a.ElapsedMS,
		CacheReadTokens:  a.CacheReadTokens,
		CacheWriteTokens: a.CacheWriteTokens,
	}
}
