package theme

import "github.com/charmbracelet/lipgloss"

func applyBrowseStyles(s *Set) {
	s.Section = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("186"))
	s.Label = lipgloss.NewStyle().Foreground(lipgloss.Color("109"))
	s.Value = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	s.Muted = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	s.Divider = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))

	s.StateRun = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	s.StateWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	s.StateStop = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	s.StateDead = lipgloss.NewStyle().Foreground(lipgloss.Color("199"))
	s.StateOther = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
}
