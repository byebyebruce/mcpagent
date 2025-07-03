package builtinserver

import (
	"context"
	"io"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ReadPage(url string) (string, error) {
	u := "https://r.jina.ai/" + url
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func WebPageReaderMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"BuiltInWebPageReader",
		"1.0.0",
	)

	// 添加工具
	{

		calculatorTool := mcp.NewTool("read_page",
			mcp.WithDescription("读取网页内容"),

			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("要读取的网页 URL"),
			),
		)

		s.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			arg := request.GetArguments()
			url := arg["url"].(string)

			content, err := ReadPage(url)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(content), nil
		})
	}
	return s
}
