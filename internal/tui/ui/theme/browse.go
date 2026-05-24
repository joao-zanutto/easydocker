package theme

import "charm.land/lipgloss/v2"

func applyBrowseStyles(s *Set) {
	s.Section = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	s.Label = lipgloss.NewStyle().Foreground(Teal)
	s.Value = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Muted = lipgloss.NewStyle().Foreground(TextMuted)
	s.Divider = lipgloss.NewStyle().Foreground(MutedPurple)

	s.StateRun = lipgloss.NewStyle().Foreground(Green)
	s.StateWarn = lipgloss.NewStyle().Foreground(Orange)
	s.StateStop = lipgloss.NewStyle().Foreground(Red)
	s.StateDead = lipgloss.NewStyle().Foreground(Pink)
	s.StateOther = lipgloss.NewStyle().Foreground(BlueGray)
}
