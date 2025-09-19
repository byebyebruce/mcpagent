package agentserver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/byebyebruce/mcpagent/openaiagent"
	"github.com/byebyebruce/mcpagent/openaiagent/history"
	"github.com/byebyebruce/mcpagent/openaiagent/mcptool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sashabaranov/go-openai"
)

func AgentMCPServer(openAIClient *openai.Client, mt *mcptool.MCPTool, systemPrompt, model string) *server.MCPServer {
	s := server.NewMCPServer(
		"AgentServer",
		"1.0.0",
	)

	_desc := `根据任务描述，调用多个工具来完成复杂任务。
可调用工具:
%s`
	var tools []string
	for _, tool := range mt.Tools() {
		tools = append(tools, tool.Function.Name+" - "+tool.Function.Description)
	}
	desc := fmt.Sprintf(_desc, strings.Join(tools, "\n"))

	// 添加工具
	{
		runTool := mcp.NewTool("run",
			mcp.WithDescription(desc),
			mcp.WithString("task",
				mcp.Required(),
				mcp.Description("任务描述"),
			),
			mcp.WithNumber("max_steps",
				mcp.Description("最大执行步骤数(max:10)"),
				mcp.DefaultNumber(5),
			),
		)

		s.AddTool(runTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			task := request.GetString("task", "") // Ensure task is present in the request
			maxStep := request.GetInt("max_steps", 5)

			slog.Info("Running agent with task", "task", task, "maxStep", maxStep)

			agent := openaiagent.NewAgent(openAIClient, systemPrompt, mt, model)

			his := history.NewHistory()
			resp, calls, err := agent.Chat(ctx, task, his, nil, nil, func(text string) {
				//slog.Info(text)
			})
			if err != nil {
				return nil, fmt.Errorf("Error in agent chat: %w", err)
			}
			his.AddMessage(task, "", nil, nil)
			for i := 0; i < maxStep; i++ {
				if len(calls) == 0 {
					slog.Info("No tool calls, returning response", "response", resp)
					return mcp.NewToolResultText(resp), nil
				}
				results, err := agent.Call(ctx, calls, func(call openai.ToolCall) {
					//fmt.Println("Tool call", call.Function.Name, result)
					slog.Info("Calling", "tool", call.Function.Name, "arguments", call.Function.Arguments)
				}, nil)
				if err != nil {
					slog.Info("Error in agent call", "error", err)
				}
				his.AddMessage("", "", calls, results)

				resp, calls, err = agent.Chat(ctx, "", his, nil, nil, func(text string) {
					fmt.Print(text)
				})
				if err != nil {
					slog.Error("Error in agent chat", "error", err)
					return nil, fmt.Errorf("Error in agent chat: %w", err)
				}
				if len(calls) == 0 {
					slog.Info("No tool calls, returning response", "response", resp)
					return mcp.NewToolResultText(resp), nil
				}
			}

			return nil, fmt.Errorf("Too much steps")
		})
	}
	return s
}
