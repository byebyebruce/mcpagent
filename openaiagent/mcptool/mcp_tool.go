package mcptool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/byebyebruce/mcpagent/mcpserver/builtinserver"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/sync/errgroup"
)

type MCPConfig struct {
	MCPServers    map[string]MCPServerConfig `json:"mcpServers"`
	EnableBuiltIn *bool                      `json:"enableBuiltIn"` // Enable built-in tools
	TimeoutSecond int                        `json:"timeoutSecond,omitempty"`
}

type MCPTool struct {
	client map[string]*client.Client
	tools  []openai.Tool
	cfg    MCPConfig
}

func NewMcpToolLoadConfig(file string) (*MCPTool, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()
	var cfg MCPConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return NewMcpTool(cfg)
}

func NewMcpTool(cfg MCPConfig) (*MCPTool, error) {
	return NewMcpToolWithInprocessClients(cfg, nil)
}
func NewMcpToolWithInprocessClients(cfg MCPConfig, mcpClients map[string]*client.Client) (*MCPTool, error) {
	if cfg.TimeoutSecond == 0 {
		cfg.TimeoutSecond = 40
	}
	clients, err := CreateMCPClients(cfg.MCPServers)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP clients: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.TimeoutSecond))
	defer cancel()
	var (
		eg, egCtx = errgroup.WithContext(ctx)
		mu        = &sync.Mutex{}
		tools     []openai.Tool
	)

	if nil == cfg.EnableBuiltIn || (*cfg.EnableBuiltIn) {
		builtinClients := builtinserver.BuiltinClient()
		for name, client := range builtinClients {
			clients[name] = client
		}
	}
	for name, client := range mcpClients {
		clients[name] = client
	}

	for _name, _client := range clients {
		name := _name
		mcpClient := _client
		eg.Go(func() error {
			ts, err := InitMCPClient(egCtx, mcpClient)
			if err != nil {
				return fmt.Errorf("failed to initialize MCP client for %s: %w", name, err)
			}
			if len(ts) == 0 {
				return nil
			}

			mu.Lock()
			for _, tool := range ts {
				ot := MCPToolToOpenAITool(&tool)
				ot.Function.Name = MCPToolName2OpenAIToolName(name, ot.Function.Name)
				tools = append(tools, ot)
			}
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to create MCP clients: %w", err)
	}
	return &MCPTool{
		client: clients,
		tools:  tools,
		cfg:    cfg,
	}, nil
}

func (m *MCPTool) Tools() []openai.Tool {
	return m.tools
}

func (m *MCPTool) Call(ctx context.Context, name, args string) (string, error) {
	server, toolName := OpenAIToolName2MCPToolName(name)
	client, ok := m.client[server]
	if !ok {
		return "", fmt.Errorf("tool %s not found", name)
	}

	request := mcp.CallToolRequest{}
	request.Params.Name = toolName
	if err := json.Unmarshal([]byte(args), &request.Params.Arguments); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	ctx, _ = context.WithTimeout(ctx, time.Second*time.Duration(m.cfg.TimeoutSecond))
	result, err := client.CallTool(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to call tool %s: %w", name, err)
	}
	if result.IsError {
		return "", fmt.Errorf("tool call error: %s", result.Content)
	}
	retStr := ""
	for i, c := range result.Content {
		switch tc := c.(type) {
		case mcp.TextContent:
			retStr += tc.Text
		case mcp.ImageContent:
			retStr += tc.Data
		default:
			continue
		}
		if len(result.Content) > 1 && i < len(result.Content)-1 {
			retStr += "\n"
		}
	}
	return retStr, nil
}

func InitMCPClient(ctx context.Context, mcpClient *client.Client) ([]mcp.Tool, error) {
	if err := mcpClient.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "Example MCP Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	_, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}
	//serverInfo.Result.Meta
	/*
		fmt.Println("Server Info:", serverInfo.Result.Meta)
		rs, err := client.ListTools(context.Background(), mcp.ListToolsRequest{})
		if err != nil {
			log.Fatalf("Failed to list tools: %v", err)
		}
		ts := []openai.Tool{}
		for _, t := range rs.Tools {
			ot := MCPToolToOpenAITool(&t)
			ts = append(ts, ot)
		}
	*/
	t, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools %w", err)
	}
	return t.Tools, nil
}
