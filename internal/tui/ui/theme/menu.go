package theme

import "charm.land/lipgloss/v2"

func applyMenuStyles(s *Set) {
	s.MenuFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("67")).Background(lipgloss.Color("234")).Padding(1, 2)
	s.MenuSelector = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	s.MenuItem = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	s.MenuDesc = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	s.HelpFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("110")).Background(lipgloss.Color("234")).Padding(1, 2)
	s.HelpTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("186"))
	s.HelpSection = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("109"))
	s.HelpCommand = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	s.HelpKey = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	s.HelpFooter = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	s.Scrollbar = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	s.HelpContext = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
}