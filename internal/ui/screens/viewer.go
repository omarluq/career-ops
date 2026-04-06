package screens

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/omarluq/career-ops/internal/ui/theme"
)

// Key constants to avoid magic strings.
const (
	viewerKeyEsc    = "esc"
	viewerKeyDown   = "down"
	viewerKeyUp     = "up"
	viewerKeyQ      = "q"
	viewerKeyJ      = "j"
	viewerKeyK      = "k"
	viewerKeyPgDown = "pgdown"
	viewerKeyPgUp   = "pgup"
	viewerKeyCtrlD  = "ctrl+d"
	viewerKeyCtrlU  = "ctrl+u"
	viewerKeyHome   = "home"
	viewerKeyEnd    = "end"
	viewerKeyG      = "g"
	viewerKeyGUpper = "G"
)

// ViewerClosedMsg is emitted when the viewer is dismissed.
type ViewerClosedMsg struct{}

// markdownRenderedMsg carries the rendered lines back to the model.
type markdownRenderedMsg struct {
	lines []string
}

// ViewerModel implements an integrated file viewer screen.
type ViewerModel struct {
	theme        theme.Theme
	lines        []string
	scrollOffset int
	width        int
	height       int
	loading      bool
}

// NewViewerWithPath returns a model and a Cmd that renders markdown in the background.
func NewViewerWithPath(t *theme.Theme, path string, width, height int) (ViewerModel, tea.Cmd) {
	m := ViewerModel{
		width:   width,
		height:  height,
		theme:   *t,
		loading: true,
		lines:   []string{"Loading..."},
	}
	cmd := func() tea.Msg {
		content, err := os.ReadFile(path) //nolint:gosec // G304: path from trusted internal source
		if err != nil {
			return markdownRenderedMsg{lines: []string{"Error: " + err.Error()}}
		}
		rendered := renderMarkdown(string(content), width)
		return markdownRenderedMsg{lines: strings.Split(rendered, "\n")}
	}
	return m, cmd
}

func renderMarkdown(src string, width int) string {
	w := width - 4
	if w < 40 {
		w = 40
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(w),
	)
	if err != nil {
		return src
	}
	out, err := r.Render(src)
	if err != nil {
		return src
	}
	return strings.TrimRight(out, "\n")
}

// Init implements tea.Model.
//
//nolint:gocritic // hugeParam: required by tea.Model interface
func (m ViewerModel) Init() tea.Cmd {
	return nil
}

// Resize updates dimensions.
func (m *ViewerModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles input for the viewer screen.
//
//nolint:gocritic // hugeParam: required by tea.Model interface
func (m ViewerModel) Update(msg tea.Msg) (ViewerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case markdownRenderedMsg:
		m.lines = msg.lines
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		return m.handleViewerKeys(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

//nolint:gocritic // hugeParam: required by tea.Model interface (called from Update)
func (m ViewerModel) handleViewerKeys(msg tea.KeyMsg) (ViewerModel, tea.Cmd) {
	switch msg.String() {
	case viewerKeyQ, viewerKeyEsc:
		return m, func() tea.Msg { return ViewerClosedMsg{} }
	case viewerKeyDown, viewerKeyJ:
		m.scrollDown()
	case viewerKeyUp, viewerKeyK:
		m.scrollUp()
	case viewerKeyPgDown, viewerKeyCtrlD:
		m.pageDown()
	case viewerKeyPgUp, viewerKeyCtrlU:
		m.pageUp()
	case viewerKeyHome, viewerKeyG:
		m.scrollOffset = 0
	case viewerKeyEnd, viewerKeyGUpper:
		m.scrollOffset = m.maxScroll()
	}
	return m, nil
}

func (m *ViewerModel) maxScroll() int {
	ms := len(m.lines) - m.bodyHeight()
	if ms < 0 {
		return 0
	}
	return ms
}

func (m *ViewerModel) scrollDown() {
	if m.scrollOffset < m.maxScroll() {
		m.scrollOffset++
	}
}

func (m *ViewerModel) scrollUp() {
	if m.scrollOffset > 0 {
		m.scrollOffset--
	}
}

func (m *ViewerModel) pageDown() {
	jump := m.bodyHeight() / 2
	m.scrollOffset += jump
	if m.scrollOffset > m.maxScroll() {
		m.scrollOffset = m.maxScroll()
	}
}

func (m *ViewerModel) pageUp() {
	jump := m.bodyHeight() / 2
	m.scrollOffset -= jump
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *ViewerModel) bodyHeight() int {
	h := m.height - 2 // footer only, no header
	if h < 3 {
		h = 3
	}
	return h
}

// View renders the viewer screen.
//
//nolint:gocritic // hugeParam: required by tea.Model interface
func (m ViewerModel) View() string {
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

func (m *ViewerModel) renderBody() string {
	bh := m.bodyHeight()
	padStyle := lipgloss.NewStyle().Padding(0, 2)

	if len(m.lines) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext)
		return padStyle.Render(emptyStyle.Render("(empty file)"))
	}

	end := m.scrollOffset + bh
	if end > len(m.lines) {
		end = len(m.lines)
	}
	visible := m.lines[m.scrollOffset:end]

	for len(visible) < bh {
		visible = append(visible, "")
	}

	return padStyle.Render(strings.Join(visible, "\n"))
}

func (m *ViewerModel) renderFooter() string {
	style := lipgloss.NewStyle().
		Foreground(m.theme.Subtext).
		Background(m.theme.Surface).
		Width(m.width).
		Padding(0, 1)

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Text)
	descStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext)

	return style.Render(
		keyStyle.Render("\u2191\u2193") + descStyle.Render(" scroll  ") +
			keyStyle.Render("PgUp/Dn") + descStyle.Render(" page  ") +
			keyStyle.Render("g/G") + descStyle.Render(" top/end  ") +
			keyStyle.Render("Esc") + descStyle.Render(" back"))
}
