package chatbot

import (
	"context"

	domainchatbot "github.com/efangly/thanes-lims-backend/internal/domain/chatbot"
)

// Assistant answers a Question by running a tool-use loop against an LLM: the
// model calls a run_sql tool, the adapter executes it via SQLRunner, feeds the
// rows back, and the model narrates a final Thai answer.
type Assistant interface {
	Ask(ctx context.Context, q domainchatbot.Question) (domainchatbot.Answer, error)
}
