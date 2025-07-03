package builtinserver

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func CalcMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"BuiltInCalculator",
		"1.0.0",
	)

	// 添加工具
	{
		calculatorTool := mcp.NewTool("calculate",
			mcp.WithDescription("执行基本的算术运算"),
			mcp.WithString("operation",
				mcp.Required(),
				mcp.Description("要执行的算术运算类型"),
				mcp.Enum("add", "subtract", "multiply", "divide"), // 保持英文
			),
			mcp.WithNumber("x",
				mcp.Required(),
				mcp.Description("第一个数字"),
			),
			mcp.WithNumber("y",
				mcp.Required(),
				mcp.Description("第二个数字"),
			),
		)

		s.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			op := request.GetString("operation", "")
			x := request.GetFloat("x", 0)
			y := request.GetFloat("y", 0)

			var result float64
			switch op {

			case "add":
				result = x + y
			case "subtract":
				result = x - y
			case "multiply":
				result = x * y
			case "divide":
				if y == 0 {

					return nil, errors.New("不允许除以零")
				}
				result = x / y
			default:
				return nil, errors.New("未知的运算类型: " + op)
			}

			return mcp.FormatNumberResult(result), nil
		})
	}
	return s
}
