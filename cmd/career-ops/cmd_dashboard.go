package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
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
type appModel struct {
	pipeline      screens.PipelineModel
	viewer        screens.ViewerModel
	state         viewState
	careerOpsPath string
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.pipeline.Resize(msg.Width, msg.Height)
		if m.state == viewReport {
			m.viewer.Resize(msg.Width, msg.Height)
		}
		pm, cmd := m.pipeline.Update(msg)
		m.pipeline = pm
		return m, cmd

	case screens.PipelineClosedMsg:
		return m, tea.Quit

	case screens.PipelineLoadReportMsg:
		archetype, tldr, remote, comp := tracker.LoadReportSummary(msg.CareerOpsPath, msg.ReportPath)
		m.pipeline.EnrichReport(msg.ReportPath, archetype, tldr, remote, comp)
		return m, nil

	case screens.PipelineUpdateStatusMsg:
		err := tracker.UpdateStatus(msg.CareerOpsPath, msg.App, msg.NewStatus)
		if err != nil {
			return m, nil
		}
		apps, parseErr := tracker.ParseApplications(m.careerOpsPath)
		if parseErr != nil {
			return m, nil
		}
		metrics := tracker.ComputeMetrics(apps)
		t := theme.NewTheme("catppuccin-mocha")
		old := m.pipeline
		m.pipeline = screens.NewPipelineModel(
			t, apps, metrics, m.careerOpsPath,
			old.Width(), old.Height(),
		)
		m.pipeline.CopyReportCache(&old)
		return m, nil

	case screens.PipelineOpenReportMsg:
		t := theme.NewTheme("catppuccin-mocha")
		m.viewer = screens.NewViewerModel(
			t, msg.Path, msg.Title,
			m.pipeline.Width(), m.pipeline.Height(),
		)
		m.state = viewReport
		return m, nil

	case screens.ViewerClosedMsg:
		m.state = viewPipeline
		return m, nil

	case screens.PipelineOpenURLMsg:
		url := msg.URL
		return m, func() tea.Msg {
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", url)
			case "linux":
				cmd = exec.Command("xdg-open", url)
			default:
				cmd = exec.Command("open", url)
			}
			_ = cmd.Start()
			return nil
		}

	default:
		if m.state == viewReport {
			vm, cmd := m.viewer.Update(msg)
			m.viewer = vm
			return m, cmd
		}
		pm, cmd := m.pipeline.Update(msg)
		m.pipeline = pm
		return m, cmd
	}
}

func (m appModel) View() string {
	if m.state == viewReport {
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
		return fmt.Errorf("could not load applications: %w", err)
	}
	if len(apps) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: no applications found; the dashboard will be empty.")
	}

	// Compute metrics.
	metrics := tracker.ComputeMetrics(apps)

	// Build the pipeline screen with an initial size estimate.
	t := theme.NewTheme("catppuccin-mocha")
	pm := screens.NewPipelineModel(t, apps, metrics, careerOpsPath, 120, 40)

	// Batch-load all report summaries so the preview is ready immediately.
	for _, app := range apps {
		if app.ReportPath == "" {
			continue
		}
		archetype, tldr, remote, comp := tracker.LoadReportSummary(careerOpsPath, app.ReportPath)
		if archetype != "" || tldr != "" || remote != "" || comp != "" {
			pm.EnrichReport(app.ReportPath, archetype, tldr, remote, comp)
		}
	}

	m := appModel{
		pipeline:      pm,
		careerOpsPath: careerOpsPath,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
