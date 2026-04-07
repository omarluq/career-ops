package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
)

// searchApplicationsTool defines the search_applications MCP tool.
func searchApplicationsTool() mcp.Tool {
	return mcp.NewTool(
		"search_applications",
		mcp.WithDescription("Full-text search across applications (company, role, notes)"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
	)
}

// handleSearchApplications searches applications by company, role, or notes.
func (s *Server) handleSearchApplications(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = ctx
	args := getArgs(req)

	query, ok := args["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "parsing applications").Error()), nil
	}

	needle := strings.ToLower(query)
	matches := lo.Filter(apps, func(app model.CareerApplication, _ int) bool {
		hay := strings.ToLower(app.Company + " " + app.Role + " " + app.Notes + " " + app.TlDr)
		return strings.Contains(hay, needle)
	})

	return marshalToolResult(matches)
}

// pipelineStatusTool defines the pipeline_status MCP tool.
func pipelineStatusTool() mcp.Tool {
	return mcp.NewTool(
		"pipeline_status",
		mcp.WithDescription("Get pipeline metrics and summary (total, by_status, avg_score, top_score)"),
	)
}

// handlePipelineStatus returns aggregate pipeline metrics.
func (s *Server) handlePipelineStatus(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = ctx
	_ = req
	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "parsing applications").Error()), nil
	}

	metrics := tracker.ComputeMetrics(apps)
	return marshalToolResult(metrics)
}

// addToPipelineTool defines the add_to_pipeline MCP tool.
func addToPipelineTool() mcp.Tool {
	return mcp.NewTool(
		"add_to_pipeline",
		mcp.WithDescription("Add a URL to the evaluation pipeline"),
		mcp.WithString("url", mcp.Required(), mcp.Description("Job posting URL to add")),
		mcp.WithString("source", mcp.Description("Source of the URL (e.g., linkedin, indeed)")),
	)
}

// handleAddToPipeline appends a URL to data/pipeline.md.
func (s *Server) handleAddToPipeline(
	ctx context.Context,
	req mcp.CallToolRequest,
) (_ *mcp.CallToolResult, retErr error) {
	_ = ctx
	args := getArgs(req)

	url, ok := args["url"].(string)
	if !ok || url == "" {
		return mcp.NewToolResultError("url parameter is required"), nil
	}

	source, ok := args["source"].(string)
	if !ok {
		source = ""
	}

	pipePath := filepath.Join(s.careerOpsPath, "data", "pipeline.md")

	line := fmt.Sprintf("- %s", url)
	if source != "" {
		line = fmt.Sprintf("- %s (source: %s)", url, source)
	}
	line += fmt.Sprintf(" <!-- added %s -->\n", time.Now().Format("2006-01-02"))

	f, err := os.OpenFile(filepath.Clean(pipePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "opening pipeline.md").Error()), nil
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil && retErr == nil {
			retErr = oops.Wrapf(closeErr, "closing pipeline.md")
		}
	}()

	if _, err := f.WriteString(line); err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "writing to pipeline.md").Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Added %s to pipeline", url)), nil
}

// listApplicationsTool defines the list_applications MCP tool.
func listApplicationsTool() mcp.Tool {
	return mcp.NewTool(
		"list_applications",
		mcp.WithDescription("List applications with optional filters"),
		mcp.WithString("status", mcp.Description("Filter by canonical status (e.g., evaluated, applied, interview)")),
		mcp.WithNumber("min_score", mcp.Description("Minimum score threshold (0-5)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results to return")),
	)
}

// handleListApplications returns a filtered list of applications.
func (s *Server) handleListApplications(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = ctx
	args := getArgs(req)

	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "parsing applications").Error()), nil
	}

	if status, ok := args["status"].(string); ok && status != "" {
		canonical := strings.ToLower(status)
		apps = lo.Filter(apps, func(app model.CareerApplication, _ int) bool {
			return strings.EqualFold(states.Normalize(app.Status), canonical)
		})
	}

	if minScore, ok := args["min_score"].(float64); ok && minScore > 0 {
		apps = lo.Filter(apps, func(app model.CareerApplication, _ int) bool {
			return app.Score >= minScore
		})
	}

	if limit, ok := args["limit"].(float64); ok && limit > 0 {
		resultLimit := int(limit)
		if resultLimit < len(apps) {
			apps = apps[:resultLimit]
		}
	}

	return marshalToolResult(apps)
}

// getApplicationTool defines the get_application MCP tool.
func getApplicationTool() mcp.Tool {
	return mcp.NewTool(
		"get_application",
		mcp.WithDescription("Get a single application by its tracker number"),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Application number in the tracker")),
	)
}

// handleGetApplication returns a single application by number.
func (s *Server) handleGetApplication(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = ctx
	args := getArgs(req)

	numRaw, ok := args["number"].(float64)
	if !ok || numRaw < 1 {
		return mcp.NewToolResultError("number parameter is required and must be >= 1"), nil
	}
	num := int(numRaw)

	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "parsing applications").Error()), nil
	}

	app, found := lo.Find(apps, func(a model.CareerApplication) bool {
		return a.Number == num
	})
	if !found {
		return mcp.NewToolResultError(fmt.Sprintf("application #%d not found", num)), nil
	}

	return marshalToolResult(app)
}

// updateStatusTool defines the update_status MCP tool.
func updateStatusTool() mcp.Tool {
	return mcp.NewTool(
		"update_status",
		mcp.WithDescription("Update an application's status by tracker number"),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Application number in the tracker")),
		mcp.WithString("status", mcp.Required(), mcp.Description("New canonical status")),
	)
}

// handleUpdateStatus changes the status of an application.
func (s *Server) handleUpdateStatus(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = ctx
	args := getArgs(req)

	numRaw, ok := args["number"].(float64)
	if !ok || numRaw < 1 {
		return mcp.NewToolResultError("number parameter is required and must be >= 1"), nil
	}

	newStatus, ok := args["status"].(string)
	if !ok || newStatus == "" {
		return mcp.NewToolResultError("status parameter is required"), nil
	}

	if !states.IsCanonical(newStatus) {
		return mcp.NewToolResultError(
			fmt.Sprintf("invalid status %q; use one of the canonical statuses", newStatus),
		), nil
	}

	num := int(numRaw)

	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "parsing applications").Error()), nil
	}

	app, found := lo.Find(apps, func(a model.CareerApplication) bool {
		return a.Number == num
	})
	if !found {
		return mcp.NewToolResultError(fmt.Sprintf("application #%d not found", num)), nil
	}

	if err := tracker.UpdateStatus(s.careerOpsPath, &app, newStatus); err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "updating status").Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Updated application #%d status to %q", num, newStatus),
	), nil
}

// getArgs extracts the arguments map from a CallToolRequest using the v0.47+ API.
func getArgs(req mcp.CallToolRequest) map[string]any {
	return req.GetArguments()
}

// marshalToolResult serializes a value to JSON and returns it as a tool result.
func marshalToolResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "marshaling result").Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
