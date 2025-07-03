package mcptool

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

func MCPToolToOpenAITool(mcpTool *mcp.Tool) openai.Tool {
	return openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			Parameters:  mcpTool.InputSchema,
		},
	}
}

func MCPToolName2OpenAIToolName(server, name string) string {
	return fmt.Sprintf("%s.%s", server, name)
}

func OpenAIToolName2MCPToolName(toolName string) (string, string) {
	parts := strings.SplitN(toolName, ".", 2)
	if len(parts) != 2 {
		return "", toolName
	}
	return parts[0], parts[1]
}
