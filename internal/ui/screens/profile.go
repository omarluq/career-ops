package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/model"
	"github.com/omarluq/career-ops/internal/ui/theme"
)

// profileKeyMap defines key bindings for the profile screen.
type profileKeyMap struct {
	Scroll   key.Binding
	Collapse key.Binding
	Page     key.Binding
	TopEnd   key.Binding
	Back     key.Binding
}

func newProfileKeyMap() profileKeyMap {
	return profileKeyMap{
		Scroll:   key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "scroll")),
		Collapse: key.NewBinding(key.WithKeys("tab", "enter"), key.WithHelp("Tab/Enter", "collapse")),
		Page:     key.NewBinding(key.WithKeys("pgup", "pgdown"), key.WithHelp("PgUp/Dn", "page")),
		TopEnd:   key.NewBinding(key.WithKeys("g", "G"), key.WithHelp("g/G", "top/end")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "back")),
	}
}

func (k profileKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Scroll, k.Collapse, k.Page, k.TopEnd, k.Back}
}

func (k profileKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// Key constants for profile screen.
const (
	profileKeyEsc    = "esc"
	profileKeyDown   = "down"
	profileKeyUp     = "up"
	profileKeyQ      = "q"
	profileKeyJ      = "j"
	profileKeyK      = "k"
	profileKeyPgDown = "pgdown"
	profileKeyPgUp   = "pgup"
	profileKeyCtrlD  = "ctrl+d"
	profileKeyCtrlU  = "ctrl+u"
	profileKeyHome   = "home"
	profileKeyEnd    = "end"
	profileKeyG      = "g"
	profileKeyGUpper = "G"
	profileKeyTab    = "tab"
	profileKeyEnter  = "enter"
)

// ProfileClosedMsg is emitted when the profile screen is dismissed.
type ProfileClosedMsg struct{}

// profileSection represents a collapsible section of the profile view.
type profileSection struct {
	title     string
	content   string
	collapsed bool
}

// ProfileModel implements the career profile display screen.
type ProfileModel struct {
	theme         theme.Theme
	profile       *model.UserProfile
	renderedLines []string
	sections      []profileSection
	help          help.Model
	keys          profileKeyMap
	height        int
	width         int
	scrollOffset  int
	sectionCursor int
	renderStale   bool
}

// NewProfileModel creates a new profile screen.
func NewProfileModel(t *theme.Theme, profile *model.UserProfile, width, height int) ProfileModel {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.Subtext)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.Subtext)

	m := ProfileModel{
		theme:       *t,
		profile:     profile,
		width:       width,
		height:      height,
		renderStale: true,
		help:        h,
		keys:        newProfileKeyMap(),
	}
	m.sections = m.buildSections()
	return m
}

// Init implements tea.Model.
func (m *ProfileModel) Init() tea.Cmd {
	return nil
}

// Resize updates dimensions.
func (m *ProfileModel) Resize(width, height int) {
	m.width = width
	m.height = height
	m.renderStale = true
}

// Update handles input for the profile screen.
func (m *ProfileModel) Update(msg tea.Msg) (ProfileModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeys(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.renderStale = true
	}

	return *m, nil
}

func (m *ProfileModel) handleKeys(msg tea.KeyMsg) (ProfileModel, tea.Cmd) {
	switch msg.String() {
	case profileKeyQ, profileKeyEsc:
		return *m, func() tea.Msg { return ProfileClosedMsg{} }
	case profileKeyDown, profileKeyJ:
		m.scrollDown()
	case profileKeyUp, profileKeyK:
		m.scrollUp()
	case profileKeyPgDown, profileKeyCtrlD:
		m.pageDown()
	case profileKeyPgUp, profileKeyCtrlU:
		m.pageUp()
	case profileKeyHome, profileKeyG:
		m.scrollOffset = 0
	case profileKeyEnd, profileKeyGUpper:
		m.scrollOffset = m.maxScroll()
	case profileKeyTab, profileKeyEnter:
		m.toggleSection()
	}

	return *m, nil
}

func (m *ProfileModel) toggleSection() {
	if m.sectionCursor >= 0 && m.sectionCursor < len(m.sections) {
		m.sections[m.sectionCursor].collapsed = !m.sections[m.sectionCursor].collapsed
		m.renderStale = true
	}
}

func (m *ProfileModel) maxScroll() int {
	ms := len(m.getRendered()) - m.bodyHeight()
	if ms < 0 {
		return 0
	}

	return ms
}

func (m *ProfileModel) scrollDown() {
	if m.scrollOffset < m.maxScroll() {
		m.scrollOffset++

		m.updateSectionCursorFromScroll()
	}
}

func (m *ProfileModel) scrollUp() {
	if m.scrollOffset > 0 {
		m.scrollOffset--

		m.updateSectionCursorFromScroll()
	}
}

func (m *ProfileModel) pageDown() {
	jump := m.bodyHeight() / 2
	m.scrollOffset += jump

	if m.scrollOffset > m.maxScroll() {
		m.scrollOffset = m.maxScroll()
	}

	m.updateSectionCursorFromScroll()
}

func (m *ProfileModel) pageUp() {
	jump := m.bodyHeight() / 2
	m.scrollOffset -= jump

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	m.updateSectionCursorFromScroll()
}

func (m *ProfileModel) bodyHeight() int {
	h := m.height - 3 // header + footer
	if h < 3 {
		h = 3
	}

	return h
}

// updateSectionCursorFromScroll moves the section cursor to whichever section
// header is closest to the current scroll position.
func (m *ProfileModel) updateSectionCursorFromScroll() {
	lines := m.getRendered()
	viewLine := m.scrollOffset

	// Walk lines backwards from viewLine to find the nearest section marker.
	bestSection := 0

	for i, sec := range m.sections {
		// Find the line index where this section title appears.
		marker := m.sectionHeader(i, sec.title, sec.collapsed)
		for idx, line := range lines {
			if idx > viewLine+m.bodyHeight()/2 {
				break
			}

			if strings.Contains(line, strings.TrimSpace(stripAnsi(marker))) {
				if idx <= viewLine+m.bodyHeight()/2 {
					bestSection = i
				}

				break
			}
		}
	}

	m.sectionCursor = bestSection
}

// View renders the profile screen.
func (m *ProfileModel) View() string {
	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m *ProfileModel) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Mauve).
		Padding(0, 1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Subtext).
		Padding(0, 1)

	var name, headline string

	if m.profile != nil {
		name = m.profile.FullName
		headline = m.profile.Headline
	}

	if name == "" {
		name = "No profile configured"
	}

	bar := lipgloss.NewStyle().
		Background(m.theme.Surface).
		Width(m.width).
		Render(titleStyle.Render(name) + subtitleStyle.Render(headline))

	return bar
}

func (m *ProfileModel) renderBody() string {
	bh := m.bodyHeight()
	padStyle := lipgloss.NewStyle().Padding(0, 2)
	lines := m.getRendered()

	if len(lines) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext)
		return padStyle.Render(emptyStyle.Render("(no profile data)"))
	}

	end := m.scrollOffset + bh
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[m.scrollOffset:end]

	for len(visible) < bh {
		visible = append(visible, "")
	}

	return padStyle.Render(strings.Join(visible, "\n"))
}

func (m *ProfileModel) renderFooter() string {
	barStyle := lipgloss.NewStyle().
		Background(m.theme.Surface).
		Width(m.width).
		Padding(0, 1)

	return barStyle.Render(m.help.View(m.keys))
}

// getRendered returns cached rendered lines, rebuilding if stale.
func (m *ProfileModel) getRendered() []string {
	if m.renderStale {
		m.renderedLines = m.renderSections()
		m.renderStale = false
	}

	return m.renderedLines
}

// renderSections builds the full list of rendered lines from all sections.
func (m *ProfileModel) renderSections() []string {
	var lines []string

	for i, sec := range m.sections {
		header := m.sectionHeader(i, sec.title, sec.collapsed)
		lines = append(lines, header)

		if !sec.collapsed && sec.content != "" {
			contentLines := strings.Split(sec.content, "\n")
			lines = append(lines, contentLines...)
		}

		lines = append(lines, "") // blank separator
	}

	return lines
}

func (m *ProfileModel) sectionHeader(index int, title string, collapsed bool) string {
	chevron := "\u25BC" // down
	if collapsed {
		chevron = "\u25B6" // right
	}

	style := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Blue)

	if index == m.sectionCursor {
		style = style.Foreground(m.theme.Mauve).Underline(true)
	}

	return style.Render(fmt.Sprintf(" %s %s", chevron, title))
}

// buildSections constructs profile sections from the user profile data.
func (m *ProfileModel) buildSections() []profileSection {
	p := *m.profile

	sections := []profileSection{
		{title: "Contact", content: m.buildContactSection(p)},
		{title: "Location", content: m.buildLocationSection(p)},
		{title: "Links", content: m.buildLinksSection(p)},
		{title: "Target Roles", content: m.buildTargetRolesSection(p)},
		{title: "Archetypes", content: m.buildArchetypesSection(p)},
		{title: "Proof Points", content: m.buildProofPointsSection(p)},
		{title: "Superpowers", content: m.buildListSection(p.Superpowers)},
		{title: "Deal Breakers", content: m.buildListSection(p.DealBreakers)},
		{title: "Compensation", content: m.buildCompSection(p)},
		{title: "Exit Story", content: m.buildExitStorySection(p)},
	}

	// Filter out empty sections.
	return lo.Filter(sections, func(s profileSection, _ int) bool {
		return strings.TrimSpace(s.content) != ""
	})
}

// profileField is a label-value pair for key-value display rows.
type profileField struct {
	label string
	value string
}

// renderKeyValueRows builds styled label-value rows, skipping empty values.
func (m *ProfileModel) renderKeyValueRows(fields []profileField, labelWidth int) string {
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.Sky).Width(labelWidth)
	valStyle := lipgloss.NewStyle().Foreground(m.theme.Text)

	var rows []string

	lo.ForEach(fields, func(f profileField, _ int) {
		if f.value != "" {
			rows = append(rows, "  "+labelStyle.Render(f.label)+valStyle.Render(f.value))
		}
	})

	return strings.Join(rows, "\n")
}

func (m *ProfileModel) buildContactSection(p model.UserProfile) string {
	return m.renderKeyValueRows([]profileField{
		{"Name", p.FullName},
		{"Email", p.Email},
		{"Phone", p.Phone},
		{"Headline", p.Headline},
	}, 12)
}

func (m *ProfileModel) buildLocationSection(p model.UserProfile) string {
	return m.renderKeyValueRows([]profileField{
		{"City", p.City},
		{"Country", p.Country},
		{"Timezone", p.Timezone},
		{"Visa", p.VisaStatus},
	}, 12)
}

func (m *ProfileModel) buildLinksSection(p model.UserProfile) string {
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.Sky).Width(12)
	linkStyle := lipgloss.NewStyle().Foreground(m.theme.Blue).Underline(true)

	links := []profileField{
		{"LinkedIn", p.LinkedInURL},
		{"GitHub", p.GitHubURL},
		{"Portfolio", p.PortfolioURL},
		{"Twitter", p.TwitterURL},
	}

	rows := lo.FilterMap(links, func(l profileField, _ int) (string, bool) {
		if l.value == "" {
			return "", false
		}

		return "  " + labelStyle.Render(l.label) + linkStyle.Render(l.value), true
	})

	return strings.Join(rows, "\n")
}

func (m *ProfileModel) buildTargetRolesSection(p model.UserProfile) string {
	if len(p.TargetRoles) == 0 {
		return ""
	}

	bulletStyle := lipgloss.NewStyle().Foreground(m.theme.Green)
	valStyle := lipgloss.NewStyle().Foreground(m.theme.Text)

	roles := lo.Map(p.TargetRoles, func(r string, _ int) string {
		return "  " + bulletStyle.Render("\u2022") + " " + valStyle.Render(r)
	})

	return strings.Join(roles, "\n")
}

func (m *ProfileModel) buildArchetypesSection(p model.UserProfile) string {
	if len(p.Archetypes) == 0 {
		return ""
	}

	nameStyle := lipgloss.NewStyle().Foreground(m.theme.Text).Width(28)
	levelStyle := lipgloss.NewStyle().Foreground(m.theme.Yellow).Width(14)
	fitStyle := lipgloss.NewStyle().Foreground(m.theme.Green)

	// Header row.
	headerStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext).Bold(true)
	header := "  " + headerStyle.Render(
		lipgloss.NewStyle().Width(28).Render("Name")+
			lipgloss.NewStyle().Width(14).Render("Level")+
			"Fit",
	)

	rows := lo.Map(p.Archetypes, func(a model.ArchetypeEntry, _ int) string {
		fit := fitStyle
		switch strings.ToLower(a.Fit) {
		case "primary":
			fit = fit.Foreground(m.theme.Green)
		case "secondary":
			fit = fit.Foreground(m.theme.Yellow)
		case "adjacent":
			fit = fit.Foreground(m.theme.Peach)
		}

		return "  " + nameStyle.Render(a.Name) + levelStyle.Render(a.Level) + fit.Render(a.Fit)
	})

	return header + "\n" + strings.Join(rows, "\n")
}

func (m *ProfileModel) buildProofPointsSection(p model.UserProfile) string {
	if len(p.ProofPoints) == 0 {
		return ""
	}

	nameStyle := lipgloss.NewStyle().Foreground(m.theme.Text).Width(24)
	metricStyle := lipgloss.NewStyle().Foreground(m.theme.Peach).Width(20)
	urlStyle := lipgloss.NewStyle().Foreground(m.theme.Blue).Underline(true)

	headerStyle := lipgloss.NewStyle().Foreground(m.theme.Subtext).Bold(true)
	header := "  " + headerStyle.Render(
		lipgloss.NewStyle().Width(24).Render("Name")+
			lipgloss.NewStyle().Width(20).Render("Hero Metric")+
			"URL",
	)

	rows := lo.Map(p.ProofPoints, func(pp model.ProofPoint, _ int) string {
		return "  " + nameStyle.Render(pp.Name) + metricStyle.Render(pp.HeroMetric) + urlStyle.Render(pp.URL)
	})

	return header + "\n" + strings.Join(rows, "\n")
}

func (m *ProfileModel) buildListSection(items []string) string {
	if len(items) == 0 {
		return ""
	}

	bulletStyle := lipgloss.NewStyle().Foreground(m.theme.Green)
	valStyle := lipgloss.NewStyle().Foreground(m.theme.Text)

	rows := lo.Map(items, func(item string, _ int) string {
		return "  " + bulletStyle.Render("\u2022") + " " + valStyle.Render(item)
	})

	return strings.Join(rows, "\n")
}

func (m *ProfileModel) buildCompSection(p model.UserProfile) string {
	return m.renderKeyValueRows([]profileField{
		{"Target", p.CompTarget},
		{"Minimum", p.CompMinimum},
		{"Currency", p.CompCurrency},
		{"Location Flex", p.CompLocationFlex},
	}, 16)
}

func (m *ProfileModel) buildExitStorySection(p model.UserProfile) string {
	if p.ExitStory == "" {
		return ""
	}

	storyStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Width(m.width - 8).
		PaddingLeft(2)

	return storyStyle.Render(p.ExitStory)
}

// stripAnsi removes ANSI escape codes for plain-text comparison.
func stripAnsi(s string) string {
	var out strings.Builder

	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true

			continue
		}

		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}

			continue
		}

		out.WriteRune(r)
	}

	return out.String()
}
