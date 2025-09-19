package history

import (
	openai "github.com/sashabaranov/go-openai"
)

type Messages []openai.ChatCompletionMessage

type Store interface {
	Save(history Messages) error
	Load() (Messages, error)
	Clear() error
}

type Memory struct {
	msgs Messages
}

func (m *Memory) Save(history Messages) error {
	m.msgs = history
	return nil
}

func (m *Memory) Load() (Messages, error) {
	_m := make(Messages, len(m.msgs))
	copy(_m, m.msgs)
	return _m, nil
}

func (m *Memory) Clear() error {
	m.msgs = make(Messages, 0)
	return nil
}

type History struct {
	maxHistory int
	store      Store
}

func NewHistory() *History {
	return NewHistoryWithStore(20, &Memory{})
}

func NewHistoryWithStore(maxHistory int, store Store) *History {
	return &History{
		maxHistory: maxHistory,
		store:      store,
	}
}

func (a *History) GetMessages() Messages {
	msgs, err := a.store.Load()
	if err != nil {
		msgs = make(Messages, 0)
	}
	return msgs
}

func (a *History) AddMessage(input string, content string, calls []openai.ToolCall, results []string) {
	msgs, err := a.store.Load()
	if err != nil {
		msgs = make(Messages, 0)
	}
	if len(input) > 0 {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})
	}
	if len(content) > 0 {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
	}
	if len(calls) > 0 {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			ToolCalls: calls,
		})
		for i, call := range calls {
			msgs = append(msgs, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Content:    results[i],
			})
		}
	}

	a.store.Save(msgs)
}

func (a *History) ClearHistory() {
	a.store.Clear()
}

func (a *History) TrimHistory() {
	max := a.maxHistory
	if max <= 0 || max > 20 {
		max = 20
	}
	msgs, err := a.store.Load()
	if err != nil {
		msgs = make(Messages, 0)
	}
	defer func() {
		a.store.Save(msgs)
	}()
	for len(msgs) > max {
		role := msgs[0].Role
		msgs = msgs[1:]
		switch role {
		//case openai.ChatMessageRoleSystem:
		case openai.ChatMessageRoleUser:
			if len(msgs) > 1 {
				if msgs[0].Role == openai.ChatMessageRoleAssistant {
					msgs = msgs[1:]
				}
			}
		}
		i := 0
		for ; i < len(msgs); i++ {
			if msgs[i].Role != openai.ChatMessageRoleTool {
				msgs = msgs[i:]
				return
			}
		}
	}
}
