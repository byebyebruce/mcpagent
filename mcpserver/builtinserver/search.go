package builtinserver

import (
	"context"
	"fmt"

	"github.com/acheong08/DuckDuckGo-API/duckduckgo"
	"github.com/acheong08/DuckDuckGo-API/typings"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func DuckDuckGoMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"BuiltInSearch",
		"1.0.0",
	)

	// 添加工具
	{

		calculatorTool := mcp.NewTool("search",
			mcp.WithDescription("使用 DuckDuckGo 进行搜索"),

			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("要搜索的内容"),
			),
			mcp.WithNumber("limit",
				mcp.Description("返回结果的数量，默认为 5"),
				mcp.DefaultNumber(5),
			),
		)

		s.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			arg := request.GetArguments()
			query := arg["query"].(string)
			limit := 5
			if n, ok := arg["limit"].(float64); ok {
				limit = int(n)
			}
			var search = typings.Search{
				Query: query,
				Limit: limit,
			}
			results, err := duckduckgo.Get_results(search)
			if err != nil {
				return nil, err
			}

			ret := ""
			for i, result := range results {
				tmp := fmt.Sprintf("# %d %s\nLink:%s\n%s\n\n", i+1, result.Title, result.Link, result.Snippet)
				ret += tmp
			}
			return mcp.NewToolResultText(ret), nil
		})
	}
	return s
}
