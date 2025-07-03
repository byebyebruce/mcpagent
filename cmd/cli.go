package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/byebyebruce/mcpagent/openaiagent"
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
		agent := openaiagent.NewAgent(openAIClient, *systemPrompt, mt, model, 10)

		ctx := context.Background()
		reader := bufio.NewReader(os.Stdin)
		interactMode := true
		if len(*input) > 0 {
			interactMode = false
			reader = bufio.NewReader(strings.NewReader(*input + "\n"))
			*autoAllowTool = true
		} else {
			fmt.Println("Enter your prompt (or type 'exit' to quit):")
		}
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
			resp, calls, err := agent.Chat(ctx, input, nil, nil, func(text string) {
				fmt.Print(text)
			})
			if err != nil {
				color.Red("Error:", err.Error())
				continue
			}
			if len(calls) == 0 {
				fmt.Println()
				agent.AddMessage(input, resp, nil, nil)
				continue
			}
			agent.AddMessage(input, "", nil, nil)
			maxRounds := 10
		LOOP:
			for i := 0; i < maxRounds && len(calls) > 0; i++ {
				//fmt.Println("Tool calls detected:")
				for _, tool := range calls {
					fmt.Println("Tool:", tool.Function.Name, "Arguments:", tool.Function.Arguments)
				}
				if !*autoAllowTool {
					confirm, _ := choose("Do you want to allow these tool calls?", "Yes", "No")
					if confirm != "Yes" {
						//fmt.Println("Tool calls not allowed, exiting.")
						break LOOP
					}
				}

				//fmt.Println("Result:")
				results, err := agent.Call(ctx, calls, func(call openai.ToolCall, result string) {
					//fmt.Println("Tool call", call.Function.Name, result)
					color.Green("Calling %s", call.Function.Name)
				})
				if err != nil {
					color.Red("Error: %s", err.Error())
				}
				agent.AddMessage("", "", calls, results)

				resp, calls, err = agent.Chat(ctx, "", nil, nil, func(text string) {
					fmt.Print(text)
				})
				if err != nil {
					color.Red("Error:", err.Error())
					break LOOP
				}
			}
			agent.AddMessage("", resp, nil, nil)
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
