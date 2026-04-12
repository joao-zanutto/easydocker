package theme

import "github.com/charmbracelet/lipgloss"

func applyChromeStyles(s *Set) {
	s.Header = lipgloss.NewStyle().Padding(1, 1, 0, 1)
	s.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1)
	s.TitleMeta = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("24")).Padding(0, 1)
	s.Tab = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1)
	s.ActiveTab = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Underline(true).Padding(0, 1)
	s.Badge = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("60")).Padding(0, 1)
	s.Footer = lipgloss.NewStyle().Padding(0, 1, 1, 1).Foreground(lipgloss.Color("248"))
	s.Key = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("31")).Padding(0, 1)
	s.KeyText = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	s.ErrorText = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
}
