// Package main implements the career-ops CLI.
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/oops"
	"github.com/spf13/cobra"

	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/tracker"
	"github.com/omarluq/career-ops/internal/ui/screens"
	"github.com/omarluq/career-ops/internal/ui/theme"
)

var dashboardPath string

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive TUI dashboard",
	RunE:  runDashboard,
}

func init() {
	dashboardCmd.Flags().StringVar(&dashboardPath, "path", ".", "Path to career-ops directory")
}

// viewState tracks which screen is active.
type viewState int

const (
	viewPipeline viewState = iota
	viewReport
)

// appModel is the top-level Bubble Tea model that routes between screens.
// Heavy fields are stored as pointers to keep the struct small enough for
// bubbletea's value-receiver tea.Model interface.
type appModel struct {
	viewer        *screens.ViewerModel
	pipeline      *screens.PipelineModel
	careerOpsPath string
	state         viewState
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case screens.PipelineClosedMsg:
		return m, tea.Quit
	case screens.PipelineLoadReportMsg:
		return m.handleLoadReport(msg)
	case screens.PipelineUpdateStatusMsg:
		return m.handleUpdateStatus(&msg)
	case screens.PipelineOpenReportMsg:
		return m.handleOpenReport(msg)
	case screens.ViewerClosedMsg:
		m.state = viewPipeline
		return m, nil
	case screens.PipelineOpenURLMsg:
		return m, openURLCmd(msg.URL)
	default:
		return m.handleDefault(msg)
	}
}

func (m *appModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.pipeline.Resize(msg.Width, msg.Height)
	if m.state == viewReport {
		m.viewer.Resize(msg.Width, msg.Height)
	}
	pm, cmd := m.pipeline.Update(msg)
	*m.pipeline = pm
	return *m, cmd
}

func (m *appModel) handleLoadReport(msg screens.PipelineLoadReportMsg) (tea.Model, tea.Cmd) {
	archetype, tldr, remote, comp := tracker.LoadReportSummary(msg.CareerOpsPath, msg.ReportPath)
	m.pipeline.EnrichReport(msg.ReportPath, archetype, tldr, remote, comp)
	return *m, nil
}

func (m *appModel) handleUpdateStatus(msg *screens.PipelineUpdateStatusMsg) (tea.Model, tea.Cmd) {
	err := tracker.UpdateStatus(msg.CareerOpsPath, &msg.App, msg.NewStatus)
	if err != nil {
		return *m, nil
	}
	apps, parseErr := tracker.ParseApplications(m.careerOpsPath)
	if parseErr != nil {
		return *m, nil
	}
	metrics := tracker.ComputeMetrics(apps)
	t := theme.NewTheme("catppuccin-mocha")
	newPM := screens.NewPipelineModel(
		&t, apps, metrics, m.careerOpsPath,
		m.pipeline.Width(), m.pipeline.Height(),
	)
	newPM.CopyReportCache(m.pipeline)
	m.pipeline = &newPM
	return *m, nil
}

func (m *appModel) handleOpenReport(msg screens.PipelineOpenReportMsg) (tea.Model, tea.Cmd) {
	t := theme.NewTheme("catppuccin-mocha")
	viewer, cmd := screens.NewViewerWithPath(
		&t, msg.Path,
		m.pipeline.Width(), m.pipeline.Height(),
	)
	m.viewer = &viewer
	m.state = viewReport
	return *m, cmd
}

func (m *appModel) handleDefault(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == viewReport {
		vm, cmd := m.viewer.Update(msg)
		*m.viewer = vm
		return *m, cmd
	}
	pm, cmd := m.pipeline.Update(msg)
	*m.pipeline = pm
	return *m, cmd
}

// openURLCmd returns a tea.Cmd that opens a URL in the user's default browser.
// The URL is validated before being passed to the system command.
func openURLCmd(rawURL string) tea.Cmd {
	return func() tea.Msg {
		parsed, err := url.Parse(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil
		}
		safeURL := parsed.String()

		ctx := context.Background()
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.CommandContext(ctx, "open", safeURL)
		case "linux":
			cmd = exec.CommandContext(ctx, "xdg-open", safeURL)
		default:
			cmd = exec.CommandContext(ctx, "open", safeURL)
		}
		if err := cmd.Start(); err != nil {
			return nil
		}
		return nil
	}
}

func (m appModel) View() string {
	if m.state == viewReport && m.viewer != nil {
		return m.viewer.View()
	}
	return m.pipeline.View()
}

func runDashboard(_ *cobra.Command, _ []string) error {
	careerOpsPath := dashboardPath

	// Initialize states from YAML config.
	states.Init(careerOpsPath)

	// Load applications.
	apps, err := tracker.ParseApplications(careerOpsPath)
	if err != nil {
		return oops.Wrapf(err, "could not load applications")
	}
	if len(apps) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: no applications found; the dashboard will be empty.")
	}

	// Compute metrics.
	metrics := tracker.ComputeMetrics(apps)

	// Build the pipeline screen with an initial size estimate.
	t := theme.NewTheme("catppuccin-mocha")
	pm := screens.NewPipelineModel(&t, apps, metrics, careerOpsPath, 120, 40)

	// Batch-load all report summaries so the preview is ready immediately.
	for i := range apps {
		if apps[i].ReportPath == "" {
			continue
		}
		archetype, tldr, remote, comp := tracker.LoadReportSummary(
			careerOpsPath, apps[i].ReportPath,
		)
		if archetype != "" || tldr != "" || remote != "" || comp != "" {
			pm.EnrichReport(apps[i].ReportPath, archetype, tldr, remote, comp)
		}
	}

	m := appModel{
		pipeline:      &pm,
		careerOpsPath: careerOpsPath,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return oops.Wrapf(err, "TUI error")
	}
	return nil
}
