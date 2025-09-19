package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byebyebruce/mcpagent/openaiagent"
	"github.com/byebyebruce/mcpagent/openaiagent/history"
	"github.com/byebyebruce/mcpagent/openaiagent/mcptool"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func CLI(openAIClient *openai.Client, model string) *cobra.Command {
	c := &cobra.Command{
		Use:   "mcptool",
		Short: "MCP Tool CLI",
		Long:  "A command line interface for interacting with MCP tools.",
	}
	var (
		autoAllowTool = c.Flags().Bool("auto-allow", false, "Automatically allow tool calls without confirmation")
		systemPrompt  = c.Flags().String("system-prompt", "", "System prompt to use for the agent")
		input         = c.Flags().String("i", "", "Input prompt for the agent (if not provided, will read from stdin)")
	)
	c.Run = func(cmd *cobra.Command, args []string) {
		cfgFile, _ := cmd.Flags().GetString("mcp")
		mt, err := mcptool.NewMcpToolLoadConfig(cfgFile)
		if err != nil {
			panic(err)
		}
		agent := openaiagent.NewAgent(openAIClient, *systemPrompt, mt, model)

		reader := bufio.NewReader(os.Stdin)
		interactMode := true
		if len(*input) > 0 {
			interactMode = false
			reader = bufio.NewReader(strings.NewReader(*input + "\n"))
			*autoAllowTool = true
		} else {
			fmt.Println("Enter your prompt (or type 'exit' to quit):")
		}
		his := history.NewHistory(10)
		for loop := true; loop; {
			if !interactMode {
				loop = false
			} else {
				fmt.Println()
				color.Yellow("User:")
			}
			input, err := reader.ReadString('\n')
			if err != nil || input == "exit" {
				break
			}
			color.Blue("Assistant:")
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)
			his.TrimHistory()
			_, err = openaiagent.Loop(ctx, agent, his, input, 10,
				func(call openai.ToolCall) bool {
					fmt.Println("Tool call", call.Function.Name, call.Function.Arguments)
					if !*autoAllowTool {
						confirm, _ := choose("Do you want to allow this tool call?", "Yes", "No")
						if confirm != "Yes" {
							return false
						}
					}
					return true
				},
				func(call openai.ToolCall, result string, err error) {
					fmt.Println("Tool result", call.Function.Name, result)
				},
				func(text string) {
					fmt.Print(text)
				})
			fmt.Println()
			if err != nil {
				color.Red("Error: %v", err.Error())
				continue
			}
		}
	}
	return c
}

func choose(label string, options ...string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}

	_, result, err := prompt.Run()
	return result, err
}
