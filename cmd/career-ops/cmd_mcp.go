package main

import (
	"github.com/spf13/cobra"

	careermcp "github.com/omarluq/career-ops/internal/mcp"
)

var mcpPath string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI assistant integration",
	Long: "Starts a Model Context Protocol server on stdin/stdout, " +
		"exposing career-ops operations as tools.",
	RunE: runMCP,
}

func init() {
	mcpCmd.Flags().StringVar(&mcpPath, "path", ".", "path to career-ops root directory")
}

func runMCP(cmd *cobra.Command, _ []string) error {
	srv := careermcp.NewServer(mcpPath)
	return srv.Start(cmd.Context())
}
