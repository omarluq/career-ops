package screens

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/omarluq/career-ops/internal/ui/theme"
)

// footerHeight is the number of lines reserved for the help bar.
const footerHeight = 1

// ViewerClosedMsg is emitted when the viewer is dismissed.
type ViewerClosedMsg struct{}

// markdownRenderedMsg carries the rendered content back to the model.
type markdownRenderedMsg struct {
	content string
}

// ViewerModel implements an integrated file viewer screen.
type ViewerModel struct {
	theme    theme.Theme
	viewport viewport.Model
	width    int
	height   int
	loading  bool
}

// NewViewerWithPath returns a model and a Cmd that renders markdown in the background.
func NewViewerWithPath(t *theme.Theme, path string, width, height int) (ViewerModel, tea.Cmd) {
	vp := viewport.New(width, height-footerHeight)
	vp.Style = lipgloss.NewStyle().Padding(0, 2)
	vp.SetContent("Loading...")

	m := ViewerModel{
		width:    width,
		height:   height,
		theme:    *t,
		viewport: vp,
		loading:  true,
	}

	cmd := func() tea.Msg {
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return markdownRenderedMsg{content: "Error: " + err.Error()}
		}
		rendered := renderMarkdown(string(content), width)
		return markdownRenderedMsg{content: rendered}
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
func (m *ViewerModel) Init() tea.Cmd {
	return nil
}

// Resize updates dimensions.
func (m *ViewerModel) Resize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - footerHeight
}

// Update handles input for the viewer screen.
func (m *ViewerModel) Update(msg tea.Msg) (ViewerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case markdownRenderedMsg:
		m.loading = false
		m.viewport.SetContent(msg.content)
		return *m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", kanbanKeyEsc:
			return *m, func() tea.Msg { return ViewerClosedMsg{} }
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - footerHeight
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return *m, cmd
}

// View renders the viewer screen.
func (m *ViewerModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), m.renderFooter())
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
