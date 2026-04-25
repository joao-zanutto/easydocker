package theme

import "github.com/charmbracelet/lipgloss"

func applyBrowseStyles(s *Set, p palette) {
	s.Section = lipgloss.NewStyle().Bold(true).Foreground(p.sectionFg)
	s.Label = lipgloss.NewStyle().Foreground(p.labelFg)
	s.Value = lipgloss.NewStyle().Foreground(p.valueFg)
	s.Muted = lipgloss.NewStyle().Foreground(p.mutedFg)
	s.Divider = lipgloss.NewStyle().Foreground(p.dividerFg)

	s.StateRun = lipgloss.NewStyle().Foreground(p.stateRunFg)
	s.StateWarn = lipgloss.NewStyle().Foreground(p.stateWarnFg)
	s.StateStop = lipgloss.NewStyle().Foreground(p.stateStopFg)
	s.StateDead = lipgloss.NewStyle().Foreground(p.stateDeadFg)
	s.StateOther = lipgloss.NewStyle().Foreground(p.stateOtherFg)
}
