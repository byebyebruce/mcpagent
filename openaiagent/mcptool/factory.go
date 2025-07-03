package mcptool

import (
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

type MCPServerConfig struct {
	Disabled bool              `json:"disabled"`
	Type     string            `json:"type"` // "stdio", "sse", "streamable"
	Command  string            `json:"command,omitempty"`
	Args     []string          `json:"args,omitempty"` // Arguments for the command
	Env      []string          `json:"env,omitempty"`
	URL      string            `json:"url,omitempty"`
	Header   map[string]string `json:"header,omitempty"`
}

func CreateMCPClient(cfg MCPServerConfig) (*client.Client, error) {
	var mcpClient *client.Client
	var err error

	switch cfg.Type {
	case "stdio":
		mcpClient, err = client.NewStdioMCPClient(cfg.Command, cfg.Env, cfg.Args...)
	case "sse":
		opt := []transport.ClientOption{}
		if len(cfg.Header) > 0 {
			opt = append(opt, transport.WithHeaders(cfg.Header))
		}
		mcpClient, err = client.NewSSEMCPClient(cfg.URL, opt...)
	case "streamable":
		opt := []transport.StreamableHTTPCOption{}
		if len(cfg.Header) > 0 {
			opt = append(opt, transport.WithHTTPHeaders(cfg.Header))
		}
		mcpClient, err = client.NewStreamableHttpClient(cfg.URL, opt...)
	default:
		return nil, fmt.Errorf("unsupported MCP server type: %s", cfg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	return mcpClient, nil
}

func CreateMCPClients(cfg map[string]MCPServerConfig) (map[string]*client.Client, error) {
	ret := make(map[string]*client.Client, len(cfg))
	for name, serverCfg := range cfg {
		if serverCfg.Disabled {
			continue
		}
		mcpClient, err := CreateMCPClient(serverCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP client for %s: %w", name, err)
		}
		ret[name] = mcpClient
	}
	return ret, nil
}
