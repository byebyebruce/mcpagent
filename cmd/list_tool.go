package cmd

import (
	"fmt"

	"github.com/byebyebruce/mcpagent/openaiagent/mcptool"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ListTools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available tools",
		Long:  `List all available tools in the MCP agent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile, _ := cmd.Flags().GetString("mcp")
			mt, err := mcptool.NewMcpToolLoadConfig(cfgFile)
			if err != nil {
				panic(err)
			}
			tools := mt.Tools()
			if len(tools) == 0 {
				cmd.Println("No tools available.")
				return nil
			}
			data := [][]string{
				{"#", "name", "description"},
			}
			for i, tool := range tools {
				data = append(data, []string{
					fmt.Sprintf("%d", i+1),
					tool.Function.Name,
					tool.Function.Description,
				})
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.Header(data[0])
			table.Bulk(data[1:])
			table.Render()

			return nil
		},
	}
	return cmd
}
