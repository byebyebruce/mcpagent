package main

import (
	"fmt"
	"os"

	"github.com/byebyebruce/mcpagent/cmd"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	godotenv.Overload()

	token := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	model := os.Getenv("OPENAI_API_MODEL")

	cfg := openai.DefaultConfig(token)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	cli := openai.NewClientWithConfig(cfg)

	root := cmd.CLI(cli, model)
	root.PersistentFlags().String("mcp", "mcp.json", "Path to the configuration file")

	root.AddCommand(
		cmd.ListTools(),
		cmd.MCPServer(cli, model),
	)
	if err := root.Execute(); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	}
}
