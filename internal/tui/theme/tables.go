package theme

import "github.com/charmbracelet/lipgloss"

func applyTableStyles(s *Set) {
	s.HeaderRow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	s.Row = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	s.ActiveRow = lipgloss.NewStyle().Bold(true).Background(s.ActiveBG)
}
