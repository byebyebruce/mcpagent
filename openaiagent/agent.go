package openaiagent

import (
	"context"
	"errors"

	"github.com/byebyebruce/mcpagent/openaiagent/history"
	openai "github.com/sashabaranov/go-openai"
)

type Agent struct {
	llmClient    *openai.Client
	tool         Tool
	model        string
	systemPrompt string // System prompt for the agent
}

func NewAgent(llmClient *openai.Client, systemPrompt string, tool Tool, model string) *Agent {
	return &Agent{
		llmClient:    llmClient,
		tool:         tool,
		systemPrompt: systemPrompt,
		model:        model,
	}
}

func (a *Agent) Chat(ctx context.Context, input string, his *history.History, calls []openai.ToolCall, results []string, onStream func(text string)) (string, []openai.ToolCall, error) {
	messages := make([]openai.ChatCompletionMessage, 0)
	if len(a.systemPrompt) > 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.systemPrompt,
		})
	}

	messages = append(messages, his.GetMessages()...)
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

type OnCallFunc func(call openai.ToolCall) bool
type OnResultFunc func(call openai.ToolCall, result string, err error)

var ErrToolCallCanceled = errors.New("tool call canceled")

func (a *Agent) Call(ctx context.Context, calls []openai.ToolCall, onCall OnCallFunc, onResult OnResultFunc) ([]string, error) {
	results := make([]string, 0, len(calls))
	var errs []error
	for _, call := range calls {
		if len(call.Function.Name) == 0 || len(call.Function.Arguments) == 0 {
			continue
		}
		if onCall != nil {
			if !onCall(call) {
				result := "Tool call canceled"
				results = append(results, result)
				return results, ErrToolCallCanceled
			}
		}
		result, err := a.tool.Call(ctx, call.Function.Name, call.Function.Arguments)
		if err != nil {
			result = "Call tool" + call.Function.Name + "return error: " + err.Error()
			errs = append(errs, err)
		}
		results = append(results, result)
		if onResult != nil {
			onResult(call, result, err)
		}
	}
	return results, errors.Join(errs...)
}
