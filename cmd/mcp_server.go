package cmd

import (
	"fmt"
	"log/slog"

	"github.com/byebyebruce/mcpagent/mcpserver/agentserver"
	"github.com/byebyebruce/mcpagent/openaiagent/mcptool"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func MCPServer(openAIClient *openai.Client, model string) *cobra.Command {
	c := &cobra.Command{
		Use:   "server",
		Short: "MCP SSE Server",
	}

	var (
		addr         = c.Flags().String("addr", ":8082", "Address to listen on for the MCP server")
		systemPrompt = c.Flags().String("system-prompt", "", "System prompt to use for the agent")
	)

	c.Run = func(cmd *cobra.Command, args []string) {
		cfgFile, _ := cmd.Flags().GetString("mcp")
		mt, err := mcptool.NewMcpToolLoadConfig(cfgFile)
		if err != nil {
			panic(err)
		}

		// Implementation of the MCP server logic goes here
		s := agentserver.AgentMCPServer(openAIClient, mt, *systemPrompt, model)
		ss := server.NewSSEServer(s)
		slog.Info("Starting MCP server", "address", *addr)
		if err := ss.Start(*addr); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}

	return c
}
