// Package chatbot is the use case layer for the AI chatbot POC: validate the
// question and delegate to the Assistant port (see docs/chatbot-poc-plan.md).
package chatbot

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	domainchatbot "github.com/efangly/thanes-lims-backend/internal/domain/chatbot"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portchatbot "github.com/efangly/thanes-lims-backend/internal/ports/chatbot"
)

// maxQuestionLen bounds a single question - the POC is single-turn Q&A, not a
// place to paste documents.
const maxQuestionLen = 500

type AskUseCase struct {
	assistant portchatbot.Assistant
}

func NewAskUseCase(assistant portchatbot.Assistant) *AskUseCase {
	return &AskUseCase{assistant: assistant}
}

type AskInput struct {
	Question string
}

func (uc *AskUseCase) Execute(ctx context.Context, in AskInput) (domainchatbot.Answer, error) {
	q := strings.TrimSpace(in.Question)
	if q == "" {
		return domainchatbot.Answer{}, fmt.Errorf("%w: question must not be empty", shared.ErrValidation)
	}
	if utf8.RuneCountInString(q) > maxQuestionLen {
		return domainchatbot.Answer{}, fmt.Errorf("%w: question must be at most %d characters", shared.ErrValidation, maxQuestionLen)
	}
	return uc.assistant.Ask(ctx, domainchatbot.Question{Text: q})
}
