package theme

import "github.com/charmbracelet/lipgloss"

func applyLogsStyles(s *Set, p palette) {
	s.Breadcrumb = lipgloss.NewStyle().Bold(true).Foreground(p.breadcrumbFg)
	s.FollowOn = lipgloss.NewStyle().Foreground(p.followOnFg).Bold(true)
	s.FollowOff = lipgloss.NewStyle().Foreground(p.followOffFg)
	s.MonitorBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(p.monitorBorderFg).Padding(0, 1)
}
