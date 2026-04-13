package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
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
	args := getArgs(req)

	query, ok := args["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	matches, err := s.repo.SearchApplications(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "searching applications").Error()), nil
	}

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
	_ = req
	metrics, err := s.repo.ComputeMetrics(ctx)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "computing metrics").Error()), nil
	}

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

// handleAddToPipeline inserts a URL into the pipeline table.
func (s *Server) handleAddToPipeline(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	url, ok := args["url"].(string)
	if !ok || url == "" {
		return mcp.NewToolResultError("url parameter is required"), nil
	}

	source, ok := args["source"].(string)
	if !ok {
		source = ""
	}

	if err := s.repo.AddToPipeline(ctx, url, source); err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "adding to pipeline").Error()), nil
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
	args := getArgs(req)

	apps, err := s.repo.ListApplications(ctx)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "listing applications").Error()), nil
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
	args := getArgs(req)

	numRaw, ok := args["number"].(float64)
	if !ok || numRaw < 1 {
		return mcp.NewToolResultError("number parameter is required and must be >= 1"), nil
	}
	num := int(numRaw)

	app, err := s.repo.GetApplication(ctx, num)
	if err != nil {
		return mcp.NewToolResultError(
			oops.Wrapf(err, "getting application %d", num).Error(),
		), nil
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

	app, err := s.repo.GetApplication(ctx, num)
	if err != nil {
		return mcp.NewToolResultError(
			oops.Wrapf(err, "getting application %d", num).Error(),
		), nil
	}

	app.Status = newStatus
	if err := s.repo.UpsertApplication(ctx, app); err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "updating status").Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Updated application #%d status to %q", num, newStatus),
	), nil
}

// --- Profile tools ---

// profileGetTool defines the profile_get MCP tool.
func profileGetTool() mcp.Tool {
	return mcp.NewTool(
		"profile_get",
		mcp.WithDescription("Returns the user's career profile as formatted text"),
	)
}

// handleProfileGet returns the user's profile.
func (s *Server) handleProfileGet(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = req
	profile, err := s.repo.GetProfile(ctx)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "getting profile").Error()), nil
	}

	return marshalToolResult(profile)
}

// profileUpdateTool defines the profile_update MCP tool.
func profileUpdateTool() mcp.Tool {
	return mcp.NewTool(
		"profile_update",
		mcp.WithDescription("Update a single field on the user's profile"),
		mcp.WithString("field", mcp.Required(), mcp.Description("Profile field name to update")),
		mcp.WithString("value", mcp.Required(), mcp.Description("New value for the field")),
	)
}

// handleProfileUpdate updates a single profile field.
func (s *Server) handleProfileUpdate(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	field, ok := args["field"].(string)
	if !ok || field == "" {
		return mcp.NewToolResultError("field parameter is required"), nil
	}

	value, ok := args["value"].(string)
	if !ok {
		return mcp.NewToolResultError("value parameter is required"), nil
	}

	if err := s.repo.UpdateProfileField(ctx, field, value); err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "updating profile field %q", field).Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Updated profile field %q", field)), nil
}

// profileEnrichTool defines the profile_enrich MCP tool.
func profileEnrichTool() mcp.Tool {
	return mcp.NewTool(
		"profile_enrich",
		mcp.WithDescription("Record a new profile enrichment from an external source"),
		mcp.WithString("source_type", mcp.Required(),
			mcp.Description("Source type: github, linkedin, blog, conversation, article, portfolio")),
		mcp.WithString("source_url", mcp.Required(), mcp.Description("URL of the source")),
		mcp.WithString("source_title", mcp.Required(), mcp.Description("Title or label for the source")),
		mcp.WithString("extracted_data", mcp.Required(), mcp.Description("Extracted data as a JSON object")),
		mcp.WithString("confidence", mcp.Required(), mcp.Description("Confidence level: low, medium, high")),
	)
}

// parseEnrichmentArgs extracts and validates enrichment parameters from a tool request.
func parseEnrichmentArgs(args map[string]any) (enrichment *model.ProfileEnrichment, errMsg string) {
	required := []string{"source_type", "source_url", "source_title", "extracted_data", "confidence"}
	vals := make(map[string]string, len(required))
	for _, key := range required {
		v, ok := args[key].(string)
		if !ok || v == "" {
			return nil, key + " parameter is required"
		}
		vals[key] = v
	}
	return &model.ProfileEnrichment{
		SourceType:    vals["source_type"],
		SourceURL:     vals["source_url"],
		SourceTitle:   vals["source_title"],
		ExtractedData: vals["extracted_data"],
		Confidence:    vals["confidence"],
	}, ""
}

// handleProfileEnrich records a new enrichment event.
func (s *Server) handleProfileEnrich(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	enrichment, errMsg := parseEnrichmentArgs(getArgs(req))
	if errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	if err := s.repo.RecordEnrichment(ctx, enrichment); err != nil {
		return mcp.NewToolResultError(
			oops.Wrapf(err, "recording enrichment").Error(),
		), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(
		"Recorded enrichment from %s: %s",
		enrichment.SourceType, enrichment.SourceTitle,
	)), nil
}

// profileEnrichmentsTool defines the profile_enrichments MCP tool.
func profileEnrichmentsTool() mcp.Tool {
	return mcp.NewTool(
		"profile_enrichments",
		mcp.WithDescription("List pending (unapplied) profile enrichments"),
	)
}

// handleProfileEnrichments returns all pending enrichments.
func (s *Server) handleProfileEnrichments(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	_ = req
	enrichments, err := s.repo.ListPendingEnrichments(ctx)
	if err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "listing pending enrichments").Error()), nil
	}

	return marshalToolResult(enrichments)
}

// profileApplyEnrichmentTool defines the profile_apply_enrichment MCP tool.
func profileApplyEnrichmentTool() mcp.Tool {
	return mcp.NewTool(
		"profile_apply_enrichment",
		mcp.WithDescription("Apply a pending enrichment to the user's profile"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Enrichment ID to apply")),
		mcp.WithString("applied_fields", mcp.Required(),
			mcp.Description("JSON array of field names that were applied")),
	)
}

// handleProfileApplyEnrichment marks an enrichment as applied.
func (s *Server) handleProfileApplyEnrichment(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	idRaw, ok := args["id"].(float64)
	if !ok || idRaw < 1 {
		return mcp.NewToolResultError("id parameter is required and must be >= 1"), nil
	}

	appliedFields, ok := args["applied_fields"].(string)
	if !ok || appliedFields == "" {
		return mcp.NewToolResultError("applied_fields parameter is required"), nil
	}

	id := int(idRaw)
	if err := s.repo.ApplyEnrichment(ctx, id, appliedFields); err != nil {
		return mcp.NewToolResultError(oops.Wrapf(err, "applying enrichment %d", id).Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Applied enrichment #%d", id)), nil
}

// registerProfileTools adds all profile management tools to the MCP server.
func (s *Server) registerProfileTools() {
	s.mcpServer.AddTool(profileGetTool(), s.handleProfileGet)
	s.mcpServer.AddTool(profileUpdateTool(), s.handleProfileUpdate)
	s.mcpServer.AddTool(profileEnrichTool(), s.handleProfileEnrich)
	s.mcpServer.AddTool(profileEnrichmentsTool(), s.handleProfileEnrichments)
	s.mcpServer.AddTool(profileApplyEnrichmentTool(), s.handleProfileApplyEnrichment)
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
