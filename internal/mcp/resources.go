package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/tracker"
)

// handleApplicationsResource returns all tracked applications as JSON.
func (s *Server) handleApplicationsResource(
	ctx context.Context,
	req mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	_ = ctx
	_ = req
	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return nil, oops.Wrapf(err, "parsing applications")
	}

	data, err := json.MarshalIndent(apps, "", "  ")
	if err != nil {
		return nil, oops.Wrapf(err, "marshaling applications")
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "applications://list",
			Text:     string(data),
			MIMEType: "application/json",
		},
	}, nil
}

// handleMetricsResource returns pipeline metrics as JSON.
func (s *Server) handleMetricsResource(
	ctx context.Context,
	req mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	_ = ctx
	_ = req
	apps, err := tracker.ParseApplications(s.careerOpsPath)
	if err != nil {
		return nil, oops.Wrapf(err, "parsing applications")
	}

	metrics := tracker.ComputeMetrics(apps)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return nil, oops.Wrapf(err, "marshaling metrics")
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "pipeline://metrics",
			Text:     string(data),
			MIMEType: "application/json",
		},
	}, nil
}
