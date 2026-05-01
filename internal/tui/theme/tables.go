package theme

import "charm.land/lipgloss/v2"

func applyTableStyles(s *Set) {
	s.HeaderRow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	s.Row = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	s.ActiveRow = lipgloss.NewStyle().Bold(true).Background(s.ActiveBG)
}
