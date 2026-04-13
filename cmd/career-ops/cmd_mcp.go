package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
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

func runMCP(cmd *cobra.Command, _ []string) (err error) {
	ctx := cmd.Context()
	dbPath := viper.GetString("db")

	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	srv := careermcp.NewServer(r, mcpPath)
	return srv.Start(ctx)
}
