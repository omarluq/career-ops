// Package screens implements Bubble Tea screen models for the TUI dashboard.
package screens

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/states"
	"github.com/omarluq/career-ops/internal/ui/theme"
)

// pipelineKeyMap defines key bindings for the pipeline screen.
type pipelineKeyMap struct {
	Nav     *key.Binding
	Tabs    *key.Binding
	Sort    *key.Binding
	Report  *key.Binding
	OpenURL *key.Binding
	Change  *key.Binding
	View    *key.Binding
	Kanban  *key.Binding
	Profile *key.Binding
	Quit    *key.Binding
}

func newPipelineKeyMap() pipelineKeyMap {
	return pipelineKeyMap{
		Nav:     lo.ToPtr(key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "nav"))),
		Tabs:    lo.ToPtr(key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←→", "tabs"))),
		Sort:    lo.ToPtr(key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort"))),
		Report:  lo.ToPtr(key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "report"))),
		OpenURL: lo.ToPtr(key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open URL"))),
		Change:  lo.ToPtr(key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "change"))),
		View:    lo.ToPtr(key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view"))),
		Kanban:  lo.ToPtr(key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "kanban"))),
		Profile: lo.ToPtr(key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "profile"))),
		Quit:    lo.ToPtr(key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "quit"))),
	}
}

// pipelineStatusPickerKeyMap is shown when the status picker is active.
type pipelineStatusPickerKeyMap struct {
	Nav     *key.Binding
	Confirm *key.Binding
	Cancel  *key.Binding
}

func newPipelineStatusPickerKeyMap() pipelineStatusPickerKeyMap {
	return pipelineStatusPickerKeyMap{
		Nav:     lo.ToPtr(key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "navigate"))),
		Confirm: lo.ToPtr(key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "confirm"))),
		Cancel:  lo.ToPtr(key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel"))),
	}
}

func (k pipelineKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		*k.Nav, *k.Tabs, *k.Sort, *k.Report, *k.OpenURL,
		*k.Change, *k.View, *k.Kanban, *k.Profile, *k.Quit,
	}
}

func (k pipelineKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (k pipelineStatusPickerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{*k.Nav, *k.Confirm, *k.Cancel}
}

func (k pipelineStatusPickerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// PipelineClosedMsg is emitted when the pipeline screen is dismissed.
type PipelineClosedMsg struct{}

// PipelineOpenReportMsg is emitted when a report should be opened in FileViewer.
type PipelineOpenReportMsg struct {
	Path   string
	Title  string
	JobURL string
}

// PipelineOpenURLMsg is emitted when a job URL should be opened in browser.
type PipelineOpenURLMsg struct {
	URL string
}

// PipelineLoadReportMsg requests lazy loading of a report summary.
type PipelineLoadReportMsg struct {
	CareerOpsPath string
	ReportPath    string
}

// PipelineUpdateStatusMsg requests a status update for an application.
type PipelineUpdateStatusMsg struct {
	CareerOpsPath string
	NewStatus     string
	App           model.CareerApplication
}

type reportSummary struct {
	archetype string
	tldr      string
	remote    string
	comp      string
}

// Sort modes.
const (
	sortScore   = "score"
	sortDate    = "date"
	sortCompany = "company"
	sortStatus  = "status"
)

// Filter modes.
const (
	filterAll       = "all"
	filterEvaluated = "evaluated"
	filterApplied   = "applied"
	filterInterview = "interview"
	filterSkip      = "skip"
	filterTop       = "top"
)

// View mode constant.
const viewGrouped = "grouped"

// Key constants for pipeline keys.
const (
	pipelineKeyDown  = "down"
	pipelineKeyUp    = "up"
	pipelineKeyEnter = "enter"
)

type pipelineTab struct {
	filter string
	label  string
}

var pipelineTabs = []pipelineTab{
	{filterAll, "ALL"},
	{filterEvaluated, "EVALUATED"},
	{filterApplied, "APPLIED"},
	{filterInterview, "INTERVIEW"},
	{filterTop, "TOP \u22654"},
	{filterSkip, "SKIP"},
}

var sortCycle = []string{sortScore, sortDate, sortCompany, sortStatus}

var statusOptions = []string{"Evaluated", "Applied", "Responded", "Interview", "Offer", "Rejected", "Discarded", "SKIP"}

// statusGroupOrder defines display order for grouped view.
var statusGroupOrder = []string{
	"interview", "offer", "responded", "applied",
	"evaluated", "skip", "rejected", "discarded",
}

// PipelineModel implements the career pipeline dashboard screen.
type PipelineModel struct {
	help             help.Model
	reportCache      map[string]reportSummary
	theme            theme.Theme
	viewMode         string
	sortMode         string
	careerOpsPath    string
	keys             pipelineKeyMap
	statusPickerKeys pipelineStatusPickerKeyMap
	filtered         []model.CareerApplication
	apps             []model.CareerApplication
	metrics          model.PipelineMetrics
	activeTab        int
	height           int
	width            int
	scrollOffset     int
	cursor           int
	statusCursor     int
	statusPicker     bool
}

// NewPipelineModel creates a new pipeline screen.
func NewPipelineModel(
	t *theme.Theme,
	apps []model.CareerApplication,
	metrics model.PipelineMetrics,
	careerOpsPath string,
	width, height int,
) PipelineModel {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.Subtext)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.Subtext)

	m := PipelineModel{
		apps:             apps,
		metrics:          metrics,
		sortMode:         sortScore,
		activeTab:        0,
		viewMode:         viewGrouped,
		width:            width,
		height:           height,
		theme:            *t,
		careerOpsPath:    careerOpsPath,
		reportCache:      make(map[string]reportSummary),
		help:             h,
		keys:             newPipelineKeyMap(),
		statusPickerKeys: newPipelineStatusPickerKeyMap(),
	}
	m.applyFilterAndSort()
	return m
}

// Init implements tea.Model.
func (m *PipelineModel) Init() tea.Cmd {
	return nil
}

// Resize updates dimensions.
func (m *PipelineModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// Width returns the current width.
func (m *PipelineModel) Width() int { return m.width }

// Height returns the current height.
func (m *PipelineModel) Height() int { return m.height }

// CopyReportCache copies the report cache from another pipeline model.
func (m *PipelineModel) CopyReportCache(other *PipelineModel) {
	lo.ForEach(lo.Entries(other.reportCache), func(e lo.Entry[string, reportSummary], _ int) {
		m.reportCache[e.Key] = e.Value
	})
}

// EnrichReport caches report summary data for preview.
func (m *PipelineModel) EnrichReport(reportPath, archetype, tldr, remote, comp string) {
	m.reportCache[reportPath] = reportSummary{
		archetype: archetype,
		tldr:      tldr,
		remote:    remote,
		comp:      comp,
	}
}

// CurrentApp returns the currently selected application, if any.
func (m *PipelineModel) CurrentApp() (model.CareerApplication, bool) {
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		return model.CareerApplication{}, false
	}
	return m.filtered[m.cursor], true
}

// Update handles input for the pipeline screen.
func (m *PipelineModel) Update(msg tea.Msg) (PipelineModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.statusPicker {
			return m.handleStatusPicker(msg)
		}
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return *m, nil
	}
	return *m, nil
}

// handleKey dispatches key messages to specialized handlers.
func (m *PipelineModel) handleKey(msg tea.KeyMsg) (PipelineModel, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return *m, func() tea.Msg { return PipelineClosedMsg{} }
	case pipelineKeyDown, pipelineKeyUp, "pgdown", "ctrl+d", "pgup", "ctrl+u":
		return m.handleNavKeys(msg)
	case "s", "f", "right", "left":
		return m.handleFilterKeys(msg)
	case "v", pipelineKeyEnter, "o", "c":
		return m.handleActionKeys(msg)
	}
	return *m, nil
}

// handleNavKeys handles cursor movement and scrolling keys.
func (m *PipelineModel) handleNavKeys(msg tea.KeyMsg) (PipelineModel, tea.Cmd) {
	switch msg.String() {
	case pipelineKeyDown:
		if len(m.filtered) > 0 {
			m.cursor++
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			m.adjustScroll()
			cmd := m.loadCurrentReport()
			return *m, cmd
		}

	case pipelineKeyUp:
		if len(m.filtered) > 0 {
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScroll()
			cmd := m.loadCurrentReport()
			return *m, cmd
		}

	case "pgdown", "ctrl+d":
		m.scrollOffset += m.height / 2
		return *m, nil

	case "pgup", "ctrl+u":
		m.scrollOffset -= m.height / 2
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return *m, nil
	}

	return *m, nil
}

// handleFilterKeys handles sort, filter, and tab switching keys.
func (m *PipelineModel) handleFilterKeys(msg tea.KeyMsg) (PipelineModel, tea.Cmd) {
	switch msg.String() {
	case "s":
		_, idx, _ := lo.FindIndexOf(sortCycle, func(s string) bool {
			return s == m.sortMode
		})
		m.sortMode = sortCycle[(idx+1)%len(sortCycle)]
		m.applyFilterAndSort()
		m.cursor = 0
		m.scrollOffset = 0

	case "f", "right":
		m.activeTab++
		if m.activeTab >= len(pipelineTabs) {
			m.activeTab = 0
		}
		m.applyFilterAndSort()
		m.cursor = 0
		m.scrollOffset = 0

	case "left":
		m.activeTab--
		if m.activeTab < 0 {
			m.activeTab = len(pipelineTabs) - 1
		}
		m.applyFilterAndSort()
		m.cursor = 0
		m.scrollOffset = 0
	}

	return *m, nil
}

// handleActionKeys handles view toggle, report open, URL open, and status change keys.
func (m *PipelineModel) handleActionKeys(msg tea.KeyMsg) (PipelineModel, tea.Cmd) {
	switch msg.String() {
	case "v":
		m.toggleViewMode()
	case pipelineKeyEnter:
		return m.openReport()
	case "o":
		return m.openJobURL()
	case "c":
		if len(m.filtered) > 0 {
			m.statusPicker = true
			m.statusCursor = 0
		}
	}
	return *m, nil
}

func (m *PipelineModel) toggleViewMode() {
	if m.viewMode == viewGrouped {
		m.viewMode = "flat"
	} else {
		m.viewMode = viewGrouped
	}
}

func (m *PipelineModel) openReport() (PipelineModel, tea.Cmd) {
	if app, ok := m.CurrentApp(); ok && app.ReportPath != "" {
		fullPath := filepath.Join(m.careerOpsPath, app.ReportPath)
		title := fmt.Sprintf("%s \u2014 %s", app.Company, app.Role)
		jobURL := app.JobURL
		return *m, func() tea.Msg {
			return PipelineOpenReportMsg{Path: fullPath, Title: title, JobURL: jobURL}
		}
	}
	return *m, nil
}

func (m *PipelineModel) openJobURL() (PipelineModel, tea.Cmd) {
	if app, ok := m.CurrentApp(); ok && app.JobURL != "" {
		return *m, func() tea.Msg {
			return PipelineOpenURLMsg{URL: app.JobURL}
		}
	}
	return *m, nil
}

func (m *PipelineModel) handleStatusPicker(msg tea.KeyMsg) (PipelineModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.statusPicker = false
		return *m, nil

	case pipelineKeyDown:
		m.statusCursor++
		if m.statusCursor >= len(statusOptions) {
			m.statusCursor = len(statusOptions) - 1
		}

	case pipelineKeyUp:
		m.statusCursor--
		if m.statusCursor < 0 {
			m.statusCursor = 0
		}

	case pipelineKeyEnter:
		m.statusPicker = false
		if app, ok := m.CurrentApp(); ok {
			newStatus := statusOptions[m.statusCursor]
			return *m, func() tea.Msg {
				return PipelineUpdateStatusMsg{
					CareerOpsPath: m.careerOpsPath,
					App:           app,
					NewStatus:     newStatus,
				}
			}
		}
	}
	return *m, nil
}

func (m *PipelineModel) loadCurrentReport() tea.Cmd {
	app, ok := m.CurrentApp()
	if !ok || app.ReportPath == "" {
		return nil
	}
	if _, cached := m.reportCache[app.ReportPath]; cached {
		return nil
	}
	path := m.careerOpsPath
	report := app.ReportPath
	return func() tea.Msg {
		return PipelineLoadReportMsg{CareerOpsPath: path, ReportPath: report}
	}
}

// applyFilterAndSort rebuilds the filtered list from apps.
func (m *PipelineModel) applyFilterAndSort() {
	m.filtered = m.filterApps()
	m.sortApps()
}

// filterApps returns the subset of apps matching the active tab filter.
func (m *PipelineModel) filterApps() []model.CareerApplication {
	currentFilter := pipelineTabs[m.activeTab].filter
	return lo.Filter(m.apps, func(app model.CareerApplication, _ int) bool {
		norm := states.Normalize(app.Status)
		switch currentFilter {
		case filterAll:
			return true
		case filterTop:
			return app.Score >= 4.0 && norm != "no_aplicar"
		default:
			return norm == currentFilter
		}
	})
}

// sortApps sorts m.filtered in place according to the current sort and view mode.
func (m *PipelineModel) sortApps() {
	m.sortByMode()

	if m.viewMode == viewGrouped {
		m.sortGrouped()
	}
}

// sortByMode applies a flat sort based on the current sort mode.
func (m *PipelineModel) sortByMode() {
	switch m.sortMode {
	case sortScore:
		sort.SliceStable(m.filtered, func(i, j int) bool {
			return m.filtered[i].Score > m.filtered[j].Score
		})
	case sortDate:
		sort.SliceStable(m.filtered, func(i, j int) bool {
			return m.filtered[i].Date > m.filtered[j].Date
		})
	case sortCompany:
		sort.SliceStable(m.filtered, func(i, j int) bool {
			return strings.ToLower(m.filtered[i].Company) < strings.ToLower(m.filtered[j].Company)
		})
	case sortStatus:
		sort.SliceStable(m.filtered, func(i, j int) bool {
			return states.Priority(m.filtered[i].Status) < states.Priority(m.filtered[j].Status)
		})
	}
}

// sortGrouped re-sorts by status priority first, then by the selected sort within groups.
func (m *PipelineModel) sortGrouped() {
	sort.SliceStable(m.filtered, func(i, j int) bool {
		pi := states.Priority(m.filtered[i].Status)
		pj := states.Priority(m.filtered[j].Status)
		if pi != pj {
			return pi < pj
		}
		return m.compareByMode(i, j)
	})
}

// compareByMode compares two filtered apps using the current sort mode.
func (m *PipelineModel) compareByMode(i, j int) bool {
	switch m.sortMode {
	case sortDate:
		return m.filtered[i].Date > m.filtered[j].Date
	case sortCompany:
		return strings.ToLower(m.filtered[i].Company) < strings.ToLower(m.filtered[j].Company)
	default:
		return m.filtered[i].Score > m.filtered[j].Score
	}
}

// adjustScroll updates scrollOffset so the cursor stays visible.
func (m *PipelineModel) adjustScroll() {
	availHeight := m.height - 12
	if availHeight < 5 {
		availHeight = 5
	}
	line := m.cursorLineEstimate()
	margin := 3

	if line >= m.scrollOffset+availHeight-margin {
		m.scrollOffset = line - availHeight + margin + 1
	}
	if line < m.scrollOffset+margin {
		m.scrollOffset = line - margin
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *PipelineModel) cursorLineEstimate() int {
	if m.viewMode != viewGrouped {
		return m.cursor
	}
	type lineState struct {
		prevStatus string
		line       int
		found      bool
	}
	result := lo.Reduce(m.filtered, func(acc lineState, app model.CareerApplication, i int) lineState {
		if acc.found {
			return acc
		}
		norm := states.Normalize(app.Status)
		if norm != acc.prevStatus {
			acc.line++
			acc.prevStatus = norm
		}
		if i == m.cursor {
			acc.found = true
			return acc
		}
		acc.line++
		return acc
	}, lineState{})
	return result.line
}

// -- View --

// View renders the pipeline screen.
func (m *PipelineModel) View() string {
	header := m.renderHeader()
	tabs := m.renderTabs()
	metricsBar := m.renderMetrics()
	sortBar := m.renderSortBar()
	body := m.renderBody()
	preview := m.renderPreview()
	helpBar := m.renderHelp()

	bodyLines := strings.Split(body, "\n")
	if m.scrollOffset > 0 && m.scrollOffset < len(bodyLines) {
		bodyLines = bodyLines[m.scrollOffset:]
	}

	previewLines := strings.Count(preview, "\n") + 1
	availHeight := m.height - 7 - previewLines
	if availHeight < 3 {
		availHeight = 3
	}
	if len(bodyLines) > availHeight {
		bodyLines = bodyLines[:availHeight]
	}
	body = strings.Join(bodyLines, "\n")

	if m.statusPicker {
		body = m.overlayStatusPicker(body)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabs,
		metricsBar,
		sortBar,
		body,
		preview,
		helpBar,
	)
}

func (m *PipelineModel) renderHeader() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text).
		Background(m.theme.Surface).
		Width(m.width).
		Padding(0, 2)

	right := lipgloss.NewStyle().Foreground(m.theme.Subtext)
	avg := fmt.Sprintf("%.1f", m.metrics.AvgScore)
	info := right.Render(fmt.Sprintf("%d offers | Avg %s/5", m.metrics.Total, avg))

	title := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Blue).Render("CAREER PIPELINE")
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(info) - 4
	if gap < 1 {
		gap = 1
	}

	return style.Render(title + strings.Repeat(" ", gap) + info)
}

func (m *PipelineModel) renderTabs() string {
	type tabResult struct {
		tab   string
		under string
	}
	results := lo.Map(pipelineTabs, func(tab pipelineTab, i int) tabResult {
		count := m.countForFilter(tab.filter)
		label := fmt.Sprintf(" %s (%d) ", tab.label, count)

		if i == m.activeTab {
			style := lipgloss.NewStyle().
				Bold(true).
				Foreground(m.theme.Blue).
				Padding(0, 0)
			return tabResult{
				tab:   style.Render(label),
				under: strings.Repeat("\u2501", lipgloss.Width(label)),
			}
		}
		style := lipgloss.NewStyle().
			Foreground(m.theme.Subtext).
			Padding(0, 0)
		return tabResult{
			tab:   style.Render(label),
			under: strings.Repeat("\u2500", lipgloss.Width(label)),
		}
	})
	tabs := lo.Map(results, func(r tabResult, _ int) string { return r.tab })
	underParts := lo.Map(results, func(r tabResult, _ int) string { return r.under })

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	underline := lipgloss.NewStyle().Foreground(m.theme.Overlay).Render(strings.Join(underParts, ""))

	padStyle := lipgloss.NewStyle().Padding(0, 1)
	return padStyle.Render(row) + "\n" + padStyle.Render(underline)
}

func (m *PipelineModel) countForFilter(filter string) int {
	return lo.CountBy(m.apps, func(app model.CareerApplication) bool {
		norm := states.Normalize(app.Status)
		switch filter {
		case filterAll:
			return true
		case filterTop:
			return app.Score >= 4.0 && norm != "no_aplicar"
		default:
			return norm == filter
		}
	})
}

func (m *PipelineModel) renderMetrics() string {
	style := lipgloss.NewStyle().
		Background(m.theme.Surface).
		Width(m.width).
		Padding(0, 2)

	statusColors := m.statusColorMap()

	parts := lo.FilterMap(statusGroupOrder, func(status string, _ int) (string, bool) {
		count, ok := m.metrics.ByStatus[status]
		if !ok || count == 0 {
			return "", false
		}
		color := statusColors[status]
		s := lipgloss.NewStyle().Foreground(color)
		return s.Render(fmt.Sprintf("%s:%d", statusLabel(status), count)), true
	})

	return style.Render(strings.Join(parts, "  "))
}

func (m *PipelineModel) renderSortBar() string {
	style := lipgloss.NewStyle().
		Foreground(m.theme.Subtext).
		Width(m.width).
		Padding(0, 2)

	sortLabel := fmt.Sprintf("[Sort: %s]", m.sortMode)
	viewLabel := fmt.Sprintf("[View: %s]", m.viewMode)
	count := fmt.Sprintf("%d shown", len(m.filtered))

	return style.Render(fmt.Sprintf("%s  %s  %s", sortLabel, viewLabel, count))
}

func (m *PipelineModel) renderBody() string {
	if len(m.filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(m.theme.Subtext).
			Padding(1, 2)
		return emptyStyle.Render("No offers match this filter")
	}

	padStyle := lipgloss.NewStyle().Padding(0, 2)

	type bodyState struct {
		prevStatus string
		lines      []string
	}
	result := lo.Reduce(m.filtered, func(acc bodyState, app model.CareerApplication, i int) bodyState {
		norm := states.Normalize(app.Status)

		if m.viewMode == viewGrouped && norm != acc.prevStatus {
			count := m.countByNormStatus(norm)
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(m.theme.Subtext)
			acc.lines = append(acc.lines, padStyle.Render(
				headerStyle.Render(fmt.Sprintf("\u2500\u2500 %s (%d) %s",
					strings.ToUpper(statusLabel(norm)), count,
					strings.Repeat("\u2500", max(0, m.width-30-len(statusLabel(norm)))))),
			))
			acc.prevStatus = norm
		}

		selected := i == m.cursor
		line := m.renderAppLine(&app, selected)
		acc.lines = append(acc.lines, line)
		return acc
	}, bodyState{})
	lines := result.lines

	return strings.Join(lines, "\n")
}

func (m *PipelineModel) renderAppLine(app *model.CareerApplication, selected bool) string {
	padStyle := lipgloss.NewStyle().Padding(0, 2)

	scoreW := 5
	companyW := 20
	statusW := 12
	compW := 14
	roleW := m.width - scoreW - companyW - statusW - compW - 10
	if roleW < 15 {
		roleW = 15
	}

	scoreStyle := m.scoreStyle(app.Score)
	score := scoreStyle.Render(fmt.Sprintf("%.1f", app.Score))

	company := app.Company
	if len(company) > companyW {
		company = company[:companyW-3] + "..."
	}
	companyStyle := lipgloss.NewStyle().Foreground(m.theme.Text).Width(companyW)

	role := app.Role
	if len(role) > roleW {
		role = role[:roleW-3] + "..."
	}
	roleStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext).Width(roleW)

	norm := states.Normalize(app.Status)
	statusColor := m.statusColorMap()[norm]
	sStyle := lipgloss.NewStyle().Foreground(statusColor).Width(statusW)
	statusText := sStyle.Render(statusLabel(norm))

	compText := ""
	if summary, ok := m.reportCache[app.ReportPath]; ok && summary.comp != "" {
		comp := summary.comp
		if len(comp) > compW-1 {
			comp = comp[:compW-4] + "..."
		}
		compStyle := lipgloss.NewStyle().Foreground(m.theme.Yellow)
		compText = compStyle.Render(comp)
	}

	line := fmt.Sprintf(" %s %s %s %s %s",
		score,
		companyStyle.Render(company),
		roleStyle.Render(role),
		statusText,
		compText,
	)

	if selected {
		selStyle := lipgloss.NewStyle().
			Background(m.theme.Overlay).
			Width(m.width - 4)
		return padStyle.Render(selStyle.Render(line))
	}
	return padStyle.Render(line)
}

func (m *PipelineModel) renderPreview() string {
	app, ok := m.CurrentApp()
	if !ok {
		return ""
	}

	padStyle := lipgloss.NewStyle().Padding(0, 2)
	divider := lipgloss.NewStyle().Foreground(m.theme.Overlay)

	var lines []string
	lines = append(lines, padStyle.Render(divider.Render(strings.Repeat("\u2500", m.width-4))))

	labelStyle := lipgloss.NewStyle().Foreground(m.theme.Sky).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(m.theme.Text)
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext)

	if summary, ok := m.reportCache[app.ReportPath]; ok {
		if summary.archetype != "" {
			lines = append(lines, padStyle.Render(
				labelStyle.Render("Arquetipo: ")+valueStyle.Render(summary.archetype)))
		}
		if summary.tldr != "" {
			lines = append(lines, padStyle.Render(
				labelStyle.Render("TL;DR: ")+valueStyle.Render(summary.tldr)))
		}
		if summary.comp != "" {
			lines = append(lines, padStyle.Render(
				labelStyle.Render("Comp: ")+valueStyle.Render(summary.comp)))
		}
		if summary.remote != "" {
			lines = append(lines, padStyle.Render(
				labelStyle.Render("Remote: ")+valueStyle.Render(summary.remote)))
		}
	} else if app.Notes != "" {
		notes := app.Notes
		if len(notes) > m.width-10 {
			notes = notes[:m.width-13] + "..."
		}
		lines = append(lines, padStyle.Render(dimStyle.Render(notes)))
	} else {
		lines = append(lines, padStyle.Render(dimStyle.Render("Loading preview...")))
	}

	return strings.Join(lines, "\n")
}

func (m *PipelineModel) renderHelp() string {
	barStyle := lipgloss.NewStyle().
		Background(m.theme.Surface).
		Width(m.width).
		Padding(0, 1)

	if m.statusPicker {
		return barStyle.Render(m.help.View(m.statusPickerKeys))
	}

	brand := lipgloss.NewStyle().Foreground(m.theme.Overlay).Render("career-ops")
	keys := m.help.View(m.keys)

	gap := m.width - lipgloss.Width(keys) - lipgloss.Width(brand) - 2
	if gap < 1 {
		gap = 1
	}

	return barStyle.Render(keys + strings.Repeat(" ", gap) + brand)
}

func (m *PipelineModel) overlayStatusPicker(body string) string {
	return renderStatusPicker(body, m.theme, m.statusCursor)
}

// -- Helpers --

func (m *PipelineModel) scoreStyle(score float64) lipgloss.Style {
	switch {
	case score >= 4.2:
		return lipgloss.NewStyle().Foreground(m.theme.Green).Bold(true)
	case score >= 3.8:
		return lipgloss.NewStyle().Foreground(m.theme.Yellow)
	case score >= 3.0:
		return lipgloss.NewStyle().Foreground(m.theme.Text)
	default:
		return lipgloss.NewStyle().Foreground(m.theme.Red)
	}
}

func (m *PipelineModel) statusColorMap() map[string]lipgloss.Color {
	return map[string]lipgloss.Color{
		"interview": m.theme.Green,
		"offer":     m.theme.Green,
		"applied":   m.theme.Sky,
		"responded": m.theme.Blue,
		"evaluated": m.theme.Text,
		"skip":      m.theme.Red,
		"rejected":  m.theme.Subtext,
		"discarded": m.theme.Subtext,
	}
}

func (m *PipelineModel) countByNormStatus(status string) int {
	return lo.CountBy(m.filtered, func(app model.CareerApplication) bool {
		return states.Normalize(app.Status) == status
	})
}

func statusLabel(norm string) string {
	switch norm {
	case "interview":
		return "Interview"
	case "offer":
		return "Offer"
	case "responded":
		return "Responded"
	case "applied":
		return "Applied"
	case "evaluated":
		return "Evaluated"
	case "skip":
		return "Skip"
	case "rejected":
		return "Rejected"
	case "discarded":
		return "Discarded"
	default:
		return norm
	}
}
