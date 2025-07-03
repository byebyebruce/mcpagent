package openaiagent

import (
	"context"
	_ "embed"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

type Messages []openai.ChatCompletionMessage

func ChatStream(ctx context.Context, llmClient *openai.Client, req openai.ChatCompletionRequest, onStream func(text string)) (string, []openai.ToolCall, error) {
	resp, err := llmClient.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", nil, err
	}
	var (
		fullText  string
		toolCalls []openai.ToolCall
	)
	for {
		ret, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", nil, err
		}

		choice := ret.Choices[0]
		if len(choice.Delta.ToolCalls) > 0 {
			// 目前只支持1个函数
			t := choice.Delta.ToolCalls[0]
			if len(t.ID) > 0 {
				tc := openai.ToolCall{
					ID:       t.ID,
					Index:    t.Index,
					Type:     t.Type,
					Function: t.Function,
				}
				tc.Function.Arguments = ""
				toolCalls = append(toolCalls, tc)
			}
			if t.Index != nil {
				idx := *t.Index
				if idx >= 0 && idx < len(toolCalls) {
					toolCalls[idx].Function.Arguments += t.Function.Arguments
				}
			}
		} else {
			result := ret.Choices[0].Delta.Content
			fullText += result
			onStream(result)
		}
	}
	return fullText, toolCalls, nil
}
