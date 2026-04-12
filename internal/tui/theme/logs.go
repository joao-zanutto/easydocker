package theme

import "github.com/charmbracelet/lipgloss"

func applyLogsStyles(s *Set) {
	s.Breadcrumb = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("247"))
	s.FollowOn = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	s.FollowOff = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	s.MonitorBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("60")).Padding(0, 1)
}
