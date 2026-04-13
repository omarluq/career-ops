// Package screens implements Bubble Tea screen models for the TUI dashboard.
package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/ui/theme"
)

// renderStatusPicker renders the status picker overlay appended to the given body.
func renderStatusPicker(body string, t theme.Theme, statusCursor int) string {
	boardLines := strings.Split(body, "\n")

	pickerWidth := 30
	padStyle := lipgloss.NewStyle().Padding(0, 2)
	borderStyle := lipgloss.NewStyle().
		Foreground(t.Blue).
		Bold(true)

	picker := append(
		[]string{padStyle.Render(borderStyle.Render("Change status:"))},
		lo.Map(statusOptions, func(opt string, i int) string {
			style := lipgloss.NewStyle().Foreground(t.Text).Width(pickerWidth)
			if i == statusCursor {
				style = style.Background(t.Overlay).Bold(true)
			}
			prefix := lo.Ternary(i == statusCursor, "> ", "  ")
			return padStyle.Render(style.Render(prefix + opt))
		})...,
	)

	boardLines = append(boardLines, picker...)
	return strings.Join(boardLines, "\n")
}
