package openaiagent

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"
)

type Agent struct {
	llmClient    *openai.Client
	tool         Tool
	model        string
	systemPrompt string // System prompt for the agent
	msgs         Messages
	maxHistory   int
}

func NewAgent(llmClient *openai.Client, systemPrompt string, tool Tool, model string, maxHistory int) *Agent {
	return &Agent{
		llmClient:    llmClient,
		tool:         tool,
		systemPrompt: systemPrompt,
		model:        model,
		maxHistory:   maxHistory,
	}
}

func trimHistory(msgs Messages) Messages {
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
			return msgs[i:]
		}
	}
	return nil
}

func (a *Agent) Tool() Tool {
	return a.tool
}

func (a *Agent) Chat(ctx context.Context, input string, calls []openai.ToolCall, results []string, onStream func(text string)) (string, []openai.ToolCall, error) {
	messages := make([]openai.ChatCompletionMessage, 0, a.maxHistory+2)
	if len(a.systemPrompt) > 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.systemPrompt,
		})
	}
	history := a.msgs
	if a.maxHistory > 0 {
		for len(history) > a.maxHistory {
			history = trimHistory(history)
		}
	}

	messages = append(messages, history...)
	if len(input) > 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})
	}
	if len(calls) > 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			ToolCalls: calls,
		})
		for i, call := range calls {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Content:    results[i],
			})
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    a.model,
		Messages: messages,
		Tools:    a.tool.Tools(),
	}

	return ChatStream(ctx, a.llmClient, req, onStream)
}

func (a *Agent) Call(ctx context.Context, calls []openai.ToolCall, onCall func(openai.ToolCall, string)) ([]string, error) {
	results := make([]string, 0, len(calls))
	var errs []error
	for _, call := range calls {
		if len(call.Function.Name) == 0 || len(call.Function.Arguments) == 0 {
			continue
		}
		result, err := a.tool.Call(ctx, call.Function.Name, call.Function.Arguments)
		if err != nil {
			result = "Call tool" + call.Function.Name + "return error: " + err.Error()
			errs = append(errs, err)
		}
		results = append(results, result)
		onCall(call, result)
	}
	return results, errors.Join(errs...)
}

func (a *Agent) AddMessage(input string, content string, calls []openai.ToolCall, results []string) {
	if len(input) > 0 {
		a.msgs = append(a.msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})
	}
	if len(content) > 0 {
		a.msgs = append(a.msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
	}
	if len(calls) > 0 {
		a.msgs = append(a.msgs, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			ToolCalls: calls,
		})
		for i, call := range calls {
			a.msgs = append(a.msgs, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Content:    results[i],
			})
		}
	}
}

func (a *Agent) ClearHistory() {
	a.msgs = nil
}
