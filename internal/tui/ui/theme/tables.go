package theme

import "charm.land/lipgloss/v2"

func applyTableStyles(s *Set) {
	s.Tables.HeaderRow = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Tables.Row = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Tables.ActiveRow = lipgloss.NewStyle().Bold(true).Background(s.ActiveBG)
}
