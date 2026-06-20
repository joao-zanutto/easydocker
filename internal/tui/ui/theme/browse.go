package theme

import "charm.land/lipgloss/v2"

func applyBrowseStyles(s *Set) {
	s.Browse.Section = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	s.Browse.Label = lipgloss.NewStyle().Foreground(Teal)
	s.Browse.Value = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Browse.Muted = lipgloss.NewStyle().Foreground(TextMuted)
	s.Browse.Divider = lipgloss.NewStyle().Foreground(MutedPurple)

	s.States.StateRun = lipgloss.NewStyle().Foreground(Green)
	s.States.StateWarn = lipgloss.NewStyle().Foreground(Orange)
	s.States.StateStop = lipgloss.NewStyle().Foreground(Red)
	s.States.StateDead = lipgloss.NewStyle().Foreground(Pink)
	s.States.StateOther = lipgloss.NewStyle().Foreground(BlueGray)
}
