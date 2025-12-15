package openaiagent

import (
	"context"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

type Messages []openai.ChatCompletionMessage

func ChatStream(ctx context.Context, llmClient *openai.Client, req openai.ChatCompletionRequest, onStream func(text string)) (string, []openai.ToolCall, error) {
	resp, err := llmClient.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Close()

	var (
		fullText  string
		toolCalls []openai.ToolCall
		// 使用 map 来跟踪 Index -> toolCalls 数组位置的映射
		indexToPos = make(map[int]int)
	)
	for {
		ret, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", nil, err
		}

		if len(ret.Choices) == 0 {
			continue
		}

		choice := ret.Choices[0]
		if len(choice.Delta.ToolCalls) > 0 {
			// 遍历所有 tool calls，而不是只取第一个
			for _, t := range choice.Delta.ToolCalls {
				// 如果有 ID，说明是新的 tool call
				if len(t.ID) > 0 {
					tc := openai.ToolCall{
						ID:       t.ID,
						Index:    t.Index,
						Type:     t.Type,
						Function: t.Function,
					}
					tc.Function.Arguments = ""
					pos := len(toolCalls)
					toolCalls = append(toolCalls, tc)
					// 记录 Index 到数组位置的映射
					if t.Index != nil {
						indexToPos[*t.Index] = pos
					}
				}

				// 累积 Arguments
				if t.Index != nil {
					idx := *t.Index
					if pos, ok := indexToPos[idx]; ok && pos < len(toolCalls) {
						toolCalls[pos].Function.Arguments += t.Function.Arguments
					}
				}
			}
		} else {
			result := choice.Delta.Content
			fullText += result
			if onStream != nil {
				onStream(result)
			}
		}
	}
	return fullText, toolCalls, nil
}
