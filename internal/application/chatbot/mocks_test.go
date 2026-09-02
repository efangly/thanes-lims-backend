package chatbot_test

import (
	"context"

	domainchatbot "github.com/efangly/thanes-lims-backend/internal/domain/chatbot"
	"github.com/stretchr/testify/mock"
)

type mockAssistant struct{ mock.Mock }

func (m *mockAssistant) Ask(ctx context.Context, q domainchatbot.Question) (domainchatbot.Answer, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(domainchatbot.Answer), args.Error(1)
}
