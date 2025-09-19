package openaiagent

import (
	"context"
	"log/slog"

	"github.com/byebyebruce/mcpagent/openaiagent/history"
)

func Loop(ctx context.Context, agent *Agent, his *history.History, input string, maxRounds int, onCall OnCallFunc, onResult OnResultFunc, onStream func(text string)) (string, error) {
	resp, calls, err := agent.Chat(ctx, input, his, nil, nil, onStream)
	if err != nil {
		return "", err
	}
	his.AddMessage(input, "", nil, nil)
	if len(calls) == 0 {
		his.AddMessage("", resp, nil, nil)
		return resp, nil
	}
	for i := 0; i < maxRounds && len(calls) > 0; i++ {
		results, err := agent.Call(ctx, calls, onCall, onResult)
		if err != nil {
			slog.Error("Error in agent call", "error", err)
		}
		his.AddMessage("", "", calls, results)
		resp, calls, err = agent.Chat(ctx, "", his, nil, nil, onStream)
		if err != nil {
			return "", err
		}
	}
	his.AddMessage("", resp, nil, nil)
	return resp, nil
}
