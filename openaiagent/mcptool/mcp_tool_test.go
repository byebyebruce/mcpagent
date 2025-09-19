package mcptool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestMCPTool(t *testing.T) {
	// Initialize the MCPTool
	f, err := os.Open("config.json")
	if err != nil {
		t.Fatalf("Failed to open config file: %v", err)
	}
	defer f.Close()
	var cfg MCPConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		t.Fatalf("Failed to decode config file: %v", err)
	}
	mcpTool, err := NewMcpTool(cfg)
	if err != nil {
		t.Fatalf("Failed to create MCPTool: %v", err)
	}

	fmt.Println(mcpTool.Tools())
	// Check if the tool is initialized correctly
	if mcpTool == nil {
		t.Fatal("Failed to initialize MCPTool")
	}

	// Call the dummy tool
	ctx := context.Background()
	result, err := mcpTool.Call(ctx, "fetch", `{"url": "https://www.theverge.com/tech"}`)
	if err != nil {
		t.Fatalf("Failed to call dummy tool: %v", err)
	}

	expected := "foo bar"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
