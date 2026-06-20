package theme

import "charm.land/lipgloss/v2"

func applyLogsStyles(s *Set) {
	s.Viewer.Breadcrumb = lipgloss.NewStyle().Bold(true).Foreground(TextBreadcrumb)
	s.Viewer.FollowOn = lipgloss.NewStyle().Foreground(TextSecondary).Bold(true)
	s.Viewer.FollowOff = lipgloss.NewStyle().Foreground(TextMuted)
}
