package theme

import "github.com/charmbracelet/lipgloss"

func applyTableStyles(s *Set, p palette) {
	s.HeaderRow = lipgloss.NewStyle().Bold(true).Foreground(p.headerRowFg)
	s.Row = lipgloss.NewStyle().Foreground(p.rowFg)
	s.ActiveRow = lipgloss.NewStyle().Bold(true).Background(s.ActiveBG)
}
