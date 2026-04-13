// Package screens implements Bubble Tea screen models for the TUI dashboard.
package screens

import (
	"fmt"
	"math"
	"path/filepath"
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

// kanbanKeyMap defines key bindings for the kanban screen.
type kanbanKeyMap struct {
	Move     *key.Binding
	Scroll   *key.Binding
	Report   *key.Binding
	Open     *key.Binding
	Status   *key.Binding
	Empty    *key.Binding
	Pipeline *key.Binding
	Profile  *key.Binding
	Quit     *key.Binding
}

func newKanbanKeyMap() kanbanKeyMap {
	return kanbanKeyMap{
		Move:     lo.ToPtr(key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←→", "move"))),
		Scroll:   lo.ToPtr(key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "scroll"))),
		Report:   lo.ToPtr(key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "report"))),
		Open:     lo.ToPtr(key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open"))),
		Status:   lo.ToPtr(key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "status"))),
		Empty:    lo.ToPtr(key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "empty"))),
		Pipeline: lo.ToPtr(key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "pipeline"))),
		Profile:  lo.ToPtr(key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "profile"))),
		Quit:     lo.ToPtr(key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "quit"))),
	}
}

// kanbanStatusPickerKeyMap is shown when the status picker is active.
type kanbanStatusPickerKeyMap struct {
	Nav     *key.Binding
	Confirm *key.Binding
	Cancel  *key.Binding
}

func newKanbanStatusPickerKeyMap() kanbanStatusPickerKeyMap {
	return kanbanStatusPickerKeyMap{
		Nav:     lo.ToPtr(key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "navigate"))),
		Confirm: lo.ToPtr(key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "confirm"))),
		Cancel:  lo.ToPtr(key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel"))),
	}
}

func (k kanbanKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{*k.Move, *k.Scroll, *k.Report, *k.Open, *k.Status, *k.Empty, *k.Pipeline, *k.Profile, *k.Quit}
}

func (k kanbanKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (k kanbanStatusPickerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{*k.Nav, *k.Confirm, *k.Cancel}
}

func (k kanbanStatusPickerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// KanbanClosedMsg is emitted when the kanban screen is dismissed.
type KanbanClosedMsg struct{}

// kanbanColumn holds the state for a single status column in the kanban board.
type kanbanColumn struct {
	status string
	label  string
	cards  []model.CareerApplication
	scroll int
	cursor int
}

// KanbanModel implements a kanban board view for the career pipeline.
type KanbanModel struct {
	theme            theme.Theme
	help             help.Model
	keys             kanbanKeyMap
	statusPickerKeys kanbanStatusPickerKeyMap
	careerOpsPath    string
	columns          []kanbanColumn
	apps             []model.CareerApplication
	activeCol        int
	height           int
	width            int
	statusCursor     int
	statusPicker     bool
	showEmpty        bool
}

// card layout constants.
const (
	kanbanCardWidth   = 26
	kanbanCardHeight  = 6
	kanbanColGap      = 1
	kanbanHeaderLines = 3
	kanbanFooterLines = 2

	kanbanKeyEsc   = "esc"
	kanbanKeyLeft  = "left"
	kanbanKeyRight = "right"
)

// NewKanbanModel creates a new kanban board screen.
func NewKanbanModel(
	t *theme.Theme,
	apps []model.CareerApplication,
	careerOpsPath string,
	width, height int,
) KanbanModel {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.Subtext)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.Subtext)

	m := KanbanModel{
		theme:            *t,
		careerOpsPath:    careerOpsPath,
		apps:             apps,
		width:            width,
		height:           height,
		showEmpty:        true,
		help:             h,
		keys:             newKanbanKeyMap(),
		statusPickerKeys: newKanbanStatusPickerKeyMap(),
	}
	m.buildColumns()
	return m
}

// Init implements tea.Model.
func (m *KanbanModel) Init() tea.Cmd {
	return nil
}

// Resize updates the kanban dimensions.
func (m *KanbanModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles input for the kanban screen.
func (m *KanbanModel) Update(msg tea.Msg) (KanbanModel, tea.Cmd) {
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

// handleKey dispatches key messages for the kanban board.
func (m *KanbanModel) handleKey(msg tea.KeyMsg) (KanbanModel, tea.Cmd) {
	switch msg.String() {
	case "q", kanbanKeyEsc:
		return *m, func() tea.Msg { return KanbanClosedMsg{} }
	case "h", kanbanKeyLeft:
		return *m.moveColumn(-1), nil
	case "l", kanbanKeyRight:
		return *m.moveColumn(1), nil
	case "j", "down":
		return *m.moveCursor(1), nil
	case "k", "up":
		return *m.moveCursor(-1), nil
	default:
		return m.handleKeyAction(msg)
	}
}

// handleKeyAction handles action keys (enter, open, status change) for the kanban board.
func (m *KanbanModel) handleKeyAction(msg tea.KeyMsg) (KanbanModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.openReport()
	case "o":
		return m.openJobURL()
	case "s":
		if len(m.columns) > 0 && len(m.columns[m.activeCol].cards) > 0 {
			m.statusPicker = true
			m.statusCursor = 0
		}
		return *m, nil
	case "e":
		m.showEmpty = !m.showEmpty
		m.buildColumns()
		if m.activeCol >= len(m.columns) {
			m.activeCol = max(0, len(m.columns)-1)
		}
		return *m, nil
	}
	return *m, nil
}

// moveColumn navigates horizontally between columns, preserving the cursor row
// so left/right feels like moving across a grid.
func (m *KanbanModel) moveColumn(delta int) *KanbanModel {
	if len(m.columns) == 0 {
		return m
	}
	prevCursor := m.columns[m.activeCol].cursor

	m.activeCol += delta
	if m.activeCol < 0 {
		m.activeCol = 0
	}
	if m.activeCol >= len(m.columns) {
		m.activeCol = len(m.columns) - 1
	}

	// Carry cursor row position to the new column (clamped to its card count).
	col := &m.columns[m.activeCol]
	col.cursor = min(prevCursor, max(0, len(col.cards)-1))
	m.adjustColumnScroll(col)
	return m
}

// moveCursor navigates vertically within the active column.
func (m *KanbanModel) moveCursor(delta int) *KanbanModel {
	if len(m.columns) == 0 {
		return m
	}
	col := &m.columns[m.activeCol]
	if len(col.cards) == 0 {
		return m
	}
	col.cursor += delta
	if col.cursor < 0 {
		col.cursor = 0
	}
	if col.cursor >= len(col.cards) {
		col.cursor = len(col.cards) - 1
	}
	m.adjustColumnScroll(col)
	return m
}

// adjustColumnScroll ensures the cursor is visible within the column viewport.
func (m *KanbanModel) adjustColumnScroll(col *kanbanColumn) {
	visibleCards := m.visibleCardCount()
	if visibleCards < 1 {
		visibleCards = 1
	}
	if col.cursor < col.scroll {
		col.scroll = col.cursor
	}
	if col.cursor >= col.scroll+visibleCards {
		col.scroll = col.cursor - visibleCards + 1
	}
}

// visibleCardCount returns how many cards fit vertically in the viewport.
func (m *KanbanModel) visibleCardCount() int {
	available := m.height - kanbanHeaderLines - kanbanFooterLines
	if available < kanbanCardHeight {
		return 1
	}
	return available / kanbanCardHeight
}

// currentApp returns the currently selected application, if any.
func (m *KanbanModel) currentApp() (model.CareerApplication, bool) {
	if len(m.columns) == 0 {
		return model.CareerApplication{}, false
	}
	col := m.columns[m.activeCol]
	if col.cursor < 0 || col.cursor >= len(col.cards) {
		return model.CareerApplication{}, false
	}
	return col.cards[col.cursor], true
}

func (m *KanbanModel) openReport() (KanbanModel, tea.Cmd) {
	app, ok := m.currentApp()
	if !ok || app.ReportPath == "" {
		return *m, nil
	}
	fullPath := filepath.Join(m.careerOpsPath, app.ReportPath)
	title := fmt.Sprintf("%s \u2014 %s", app.Company, app.Role)
	jobURL := app.JobURL
	return *m, func() tea.Msg {
		return PipelineOpenReportMsg{Path: fullPath, Title: title, JobURL: jobURL}
	}
}

func (m *KanbanModel) openJobURL() (KanbanModel, tea.Cmd) {
	app, ok := m.currentApp()
	if !ok || app.JobURL == "" {
		return *m, nil
	}
	return *m, func() tea.Msg {
		return PipelineOpenURLMsg{URL: app.JobURL}
	}
}

func (m *KanbanModel) handleStatusPicker(msg tea.KeyMsg) (KanbanModel, tea.Cmd) {
	switch msg.String() {
	case kanbanKeyEsc, "q":
		m.statusPicker = false
		return *m, nil
	case "down", "j":
		m.statusCursor++
		if m.statusCursor >= len(statusOptions) {
			m.statusCursor = len(statusOptions) - 1
		}
	case "up", "k":
		m.statusCursor--
		if m.statusCursor < 0 {
			m.statusCursor = 0
		}
	case "enter":
		m.statusPicker = false
		app, ok := m.currentApp()
		if ok {
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

// buildColumns groups applications by status into kanban columns.
func (m *KanbanModel) buildColumns() {
	grouped := lo.GroupBy(m.apps, func(app model.CareerApplication) string {
		return states.Normalize(app.Status)
	})

	m.columns = lo.FilterMap(statusGroupOrder, func(status string, _ int) (kanbanColumn, bool) {
		cards := grouped[status]
		if !m.showEmpty && len(cards) == 0 {
			return kanbanColumn{}, false
		}
		return kanbanColumn{
			status: status,
			label:  statusLabel(status),
			cards:  cards,
		}, true
	})
}

// RebuildColumns rebuilds columns from updated apps.
func (m *KanbanModel) RebuildColumns(apps []model.CareerApplication) {
	m.apps = apps
	m.buildColumns()
	if m.activeCol >= len(m.columns) {
		m.activeCol = max(0, len(m.columns)-1)
	}
}

// -- View --

// View renders the kanban board.
func (m *KanbanModel) View() string {
	header := m.renderHeader()
	board := m.renderBoard()
	helpBar := m.renderHelp()

	if m.statusPicker {
		board = m.overlayStatusPicker(board)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, board, helpBar)
}

func (m *KanbanModel) renderHeader() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text).
		Background(m.theme.Surface).
		Width(m.width).
		Padding(0, 2)

	title := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Blue).Render("KANBAN BOARD")
	count := lipgloss.NewStyle().Foreground(m.theme.Subtext).
		Render(fmt.Sprintf("%d offers | %d columns", len(m.apps), len(m.columns)))

	gap := m.width - lipgloss.Width(title) - lipgloss.Width(count) - 4
	if gap < 1 {
		gap = 1
	}

	return style.Render(title + strings.Repeat(" ", gap) + count)
}

func (m *KanbanModel) renderBoard() string {
	if len(m.columns) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(m.theme.Subtext).
			Padding(1, 2)
		return emptyStyle.Render("No applications to display")
	}

	colWidth := m.columnWidth()
	visibleCards := m.visibleCardCount()

	rendered := lo.Map(m.columns, func(col kanbanColumn, i int) string {
		return m.renderColumn(&col, i, colWidth, visibleCards)
	})

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// columnWidth calculates the width available for each column.
func (m *KanbanModel) columnWidth() int {
	if len(m.columns) == 0 {
		return kanbanCardWidth
	}
	totalGaps := (len(m.columns) - 1) * kanbanColGap
	available := m.width - totalGaps - 2
	w := available / len(m.columns)
	if w < kanbanCardWidth {
		w = kanbanCardWidth
	}
	return w
}

func (m *KanbanModel) renderColumn(col *kanbanColumn, colIdx, colWidth, visibleCards int) string {
	isActive := colIdx == m.activeCol
	statusColors := m.statusColorMap()
	colColor := statusColors[col.status]

	// Column header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colColor).
		Width(colWidth).
		Align(lipgloss.Center)

	headerText := fmt.Sprintf("%s (%d)", col.label, len(col.cards))
	header := headerStyle.Render(headerText)

	dividerColor := lo.Ternary(isActive, colColor, m.theme.Overlay)
	divider := lipgloss.NewStyle().
		Foreground(dividerColor).
		Width(colWidth).
		Align(lipgloss.Center).
		Render(strings.Repeat("\u2500", colWidth-2))

	// Render visible cards
	endIdx := col.scroll + visibleCards
	if endIdx > len(col.cards) {
		endIdx = len(col.cards)
	}
	startIdx := col.scroll

	cardSlice := col.cards[startIdx:endIdx]
	cards := lo.Map(cardSlice, func(app model.CareerApplication, i int) string {
		actualIdx := startIdx + i
		selected := isActive && actualIdx == col.cursor
		return m.renderCard(&app, selected, colWidth)
	})

	// Scroll indicators
	scrollUp := lo.Ternary(col.scroll > 0,
		lipgloss.NewStyle().Foreground(m.theme.Subtext).Width(colWidth).Align(lipgloss.Center).Render("\u25b2"),
		strings.Repeat(" ", colWidth))
	scrollDown := lo.Ternary(endIdx < len(col.cards),
		lipgloss.NewStyle().Foreground(m.theme.Subtext).Width(colWidth).Align(lipgloss.Center).Render("\u25bc"),
		strings.Repeat(" ", colWidth))

	parts := make([]string, 0, 3+len(cards)+1)
	parts = append(parts, header, divider, scrollUp)
	parts = append(parts, cards...)
	parts = append(parts, scrollDown)

	colStyle := lipgloss.NewStyle().
		Width(colWidth).
		MarginRight(kanbanColGap)

	return colStyle.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m *KanbanModel) renderCard(app *model.CareerApplication, selected bool, colWidth int) string {
	cardInner := colWidth - 4
	if cardInner < 10 {
		cardInner = 10
	}

	// Company name
	company := app.Company
	if len(company) > cardInner {
		company = company[:cardInner-3] + "..."
	}
	companyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text).
		Width(cardInner)

	// Role
	role := app.Role
	if len(role) > cardInner {
		role = role[:cardInner-3] + "..."
	}
	roleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Subtext).
		Width(cardInner)

	// Score with stars
	scoreStyle := m.scoreStyle(app.Score)
	stars := m.renderStars(app.Score)
	scoreLine := scoreStyle.Render(fmt.Sprintf("%.1f/5", app.Score)) + " " + stars

	// Date
	dateStyle := lipgloss.NewStyle().Foreground(m.theme.Overlay)
	dateLine := dateStyle.Render(app.Date)

	content := lipgloss.JoinVertical(lipgloss.Left,
		companyStyle.Render(company),
		roleStyle.Render(role),
		scoreLine,
		dateLine,
	)

	borderColor := m.cardBorderColor(app.Score, selected)
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(colWidth-2).
		Padding(0, 1)

	if selected {
		cardStyle = cardStyle.
			Background(m.theme.Surface)
	}

	return cardStyle.Render(content)
}

// renderStars returns a star rating string for a score.
func (m *KanbanModel) renderStars(score float64) string {
	full := int(math.Floor(score))
	if full > 5 {
		full = 5
	}
	empty := 5 - full
	starStyle := lipgloss.NewStyle().Foreground(m.theme.Yellow)
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Overlay)
	return starStyle.Render(strings.Repeat("\u2605", full)) + dimStyle.Render(strings.Repeat("\u2606", empty))
}

// cardBorderColor returns the border color based on score and selection state.
func (m *KanbanModel) cardBorderColor(score float64, selected bool) lipgloss.Color {
	if selected {
		return m.theme.Blue
	}
	switch {
	case score >= 4.0:
		return m.theme.Green
	case score >= 3.0:
		return m.theme.Yellow
	default:
		return m.theme.Red
	}
}

func (m *KanbanModel) scoreStyle(score float64) lipgloss.Style {
	switch {
	case score >= 4.0:
		return lipgloss.NewStyle().Foreground(m.theme.Green).Bold(true)
	case score >= 3.0:
		return lipgloss.NewStyle().Foreground(m.theme.Yellow)
	default:
		return lipgloss.NewStyle().Foreground(m.theme.Red)
	}
}

func (m *KanbanModel) statusColorMap() map[string]lipgloss.Color {
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

func (m *KanbanModel) renderHelp() string {
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

func (m *KanbanModel) overlayStatusPicker(board string) string {
	return renderStatusPicker(board, m.theme, m.statusCursor)
}
