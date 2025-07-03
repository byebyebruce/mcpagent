package openaiagent

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Tool interface {
	Tools() []openai.Tool
	Call(ctx context.Context, name, arg string) (string, error)
}

type EmptyTool struct{}

func (e *EmptyTool) Tools() []openai.Tool {
	return nil
}
func (e *EmptyTool) Call(ctx context.Context, name, arg string) (string, error) {
	return "", nil
}
