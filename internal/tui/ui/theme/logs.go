package theme

import "charm.land/lipgloss/v2"

func applyLogsStyles(s *Set) {
	s.Breadcrumb = lipgloss.NewStyle().Bold(true).Foreground(TextBreadcrumb)
	s.FollowOn = lipgloss.NewStyle().Foreground(TextSecondary).Bold(true)
	s.FollowOff = lipgloss.NewStyle().Foreground(TextMuted)
}
