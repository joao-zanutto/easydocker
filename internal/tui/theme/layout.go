package theme

import "github.com/charmbracelet/lipgloss"

func applyFrameStyles(s *Set) {
	s.MainFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("67")).Padding(0, 1)
	s.SubpageFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("110")).Padding(0, 1)
}
