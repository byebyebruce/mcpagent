package builtinserver

import (
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/server"
)

var builtinServer = map[string]*server.MCPServer{
	"calculator": CalcMCPServer(),
	"search":     DuckDuckGoMCPServer(),
	"web":        WebPageReaderMCPServer(),
}

func BuiltinClient() map[string]*client.Client {
	var clients = make(map[string]*client.Client)
	for name, srv := range builtinServer {
		client, err := client.NewInProcessClient(srv)
		if err != nil {
			panic(err)
		}
		clients[name] = client
	}
	return clients
}
