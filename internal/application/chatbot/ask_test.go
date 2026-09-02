package chatbot_test

import (
	"context"
	"strings"
	"testing"

	applicationchatbot "github.com/efangly/thanes-lims-backend/internal/application/chatbot"
	domainchatbot "github.com/efangly/thanes-lims-backend/internal/domain/chatbot"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsk_RejectsEmptyQuestion(t *testing.T) {
	uc := applicationchatbot.NewAskUseCase(new(mockAssistant))
	_, err := uc.Execute(context.Background(), applicationchatbot.AskInput{Question: "   "})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestAsk_RejectsTooLongQuestion(t *testing.T) {
	uc := applicationchatbot.NewAskUseCase(new(mockAssistant))
	_, err := uc.Execute(context.Background(), applicationchatbot.AskInput{Question: strings.Repeat("ก", 501)})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestAsk_DelegatesToAssistant(t *testing.T) {
	assistant := new(mockAssistant)
	want := domainchatbot.Answer{Text: "มี 3 รายการ", Rows: 3}
	assistant.On("Ask", mock.Anything, domainchatbot.Question{Text: "มี sample ค้างกี่รายการ"}).Return(want, nil)

	uc := applicationchatbot.NewAskUseCase(assistant)
	got, err := uc.Execute(context.Background(), applicationchatbot.AskInput{Question: "  มี sample ค้างกี่รายการ  "})

	assert.NoError(t, err)
	assert.Equal(t, want, got)
	assistant.AssertExpectations(t)
}
