package theme

import "charm.land/lipgloss/v2"

func applyTableStyles(s *Set) {
	s.HeaderRow = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Row = lipgloss.NewStyle().Foreground(TextSecondary)
	s.ActiveRow = lipgloss.NewStyle().Bold(true).Background(s.ActiveBG)
}
