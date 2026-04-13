// Package mcp exposes career-ops operations as a Model Context Protocol server.
package mcp

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/omarluq/career-ops/internal/db"
)

// Server wraps career-ops functionality as an MCP tool server.
type Server struct {
	mcpServer     *server.MCPServer
	repo          db.Repository
	careerOpsPath string
}

// NewServer creates a new MCP server backed by the given repository.
// careerOpsPath is retained for report file access.
func NewServer(r db.Repository, careerOpsPath string) *Server {
	s := &Server{repo: r, careerOpsPath: careerOpsPath}

	s.mcpServer = server.NewMCPServer(
		"career-ops",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	s.registerTools()
	s.registerResources()

	return s
}

// Start begins serving MCP requests on stdin/stdout (stdio transport).
func (s *Server) Start(ctx context.Context) error {
	stdio := server.NewStdioServer(s.mcpServer)
	return stdio.Listen(ctx, os.Stdin, os.Stdout)
}

// registerTools adds all career-ops tools to the MCP server.
func (s *Server) registerTools() {
	s.mcpServer.AddTool(searchApplicationsTool(), s.handleSearchApplications)
	s.mcpServer.AddTool(pipelineStatusTool(), s.handlePipelineStatus)
	s.mcpServer.AddTool(addToPipelineTool(), s.handleAddToPipeline)
	s.mcpServer.AddTool(listApplicationsTool(), s.handleListApplications)
	s.mcpServer.AddTool(getApplicationTool(), s.handleGetApplication)
	s.mcpServer.AddTool(updateStatusTool(), s.handleUpdateStatus)
	s.registerProfileTools()
}

// registerResources adds all career-ops resources to the MCP server.
func (s *Server) registerResources() {
	s.mcpServer.AddResource(
		mcp.NewResource(
			"applications://list",
			"All tracked applications",
			mcp.WithResourceDescription("Full list of all job applications in the tracker"),
			mcp.WithMIMEType("application/json"),
		),
		s.handleApplicationsResource,
	)

	s.mcpServer.AddResource(
		mcp.NewResource(
			"pipeline://metrics",
			"Pipeline metrics summary",
			mcp.WithResourceDescription("Aggregate stats: totals, averages, breakdowns by status"),
			mcp.WithMIMEType("application/json"),
		),
		s.handleMetricsResource,
	)
}
