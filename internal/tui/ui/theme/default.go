package theme

import "charm.land/lipgloss/v2"

func Default() Set {
	s := Set{Chrome: ChromeStyles{Page: lipgloss.NewStyle()}, ActiveBG: Surface}

	s.Chrome.Header = lipgloss.NewStyle().Padding(0, 1, 0, 1)
	s.Chrome.Title = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SurfaceBG).Padding(0, 1)
	s.Chrome.TitleMeta = lipgloss.NewStyle().Foreground(TextSecondary).Padding(0, 1)
	s.Chrome.Badge = lipgloss.NewStyle().Foreground(LightYellow).Padding(0, 1)
	s.Chrome.ErrorText = lipgloss.NewStyle().Foreground(Red)
	s.Chrome.Footer = lipgloss.NewStyle().Padding(0, 0, 0, 0).Foreground(TextDim)
	s.Chrome.Key = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SteelBlue).Padding(0, 1)
	s.Chrome.KeyText = lipgloss.NewStyle().Foreground(TextDim)
	s.Tabs.Tab = lipgloss.NewStyle().Foreground(TextSecondary).Padding(0, 1)
	s.Tabs.ActiveTab = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SelectorPurple).Padding(0, 1)

	s.Browse.Section = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	s.Browse.Label = lipgloss.NewStyle().Foreground(Teal)
	s.Browse.Value = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Browse.Muted = lipgloss.NewStyle().Foreground(TextMuted)
	s.Browse.Divider = lipgloss.NewStyle().Foreground(MutedPurple)
	s.Browse.MainFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Periwinkle).Padding(0, 1)

	s.Viewer.Breadcrumb = lipgloss.NewStyle().Bold(true).Foreground(TextBreadcrumb)
	s.Viewer.FollowOn = lipgloss.NewStyle().Foreground(TextSecondary).Bold(true)
	s.Viewer.FollowOff = lipgloss.NewStyle().Foreground(TextMuted)
	s.Viewer.SubpageFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BlueGray).Padding(0, 1)

	s.Tables.HeaderRow = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Tables.Row = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Tables.ActiveRow = lipgloss.NewStyle().Bold(true).Background(s.ActiveBG)

	s.Menu.Frame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Periwinkle).Background(SurfaceDeep).Padding(1, 2)
	s.Menu.Selector = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Menu.ItemNormal = lipgloss.NewStyle().Bold(true).Foreground(TextMuted)
	s.Menu.ItemDescription = lipgloss.NewStyle().Bold(true).Foreground(TextSilver)
	s.Menu.HelpFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BlueGray).Background(SurfaceDeep).Padding(1, 2)
	s.Menu.HelpTitle = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	s.Menu.HelpSection = lipgloss.NewStyle().Bold(true).Foreground(Teal)
	s.Menu.HelpCommand = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Menu.HelpKey = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Menu.HelpFooter = lipgloss.NewStyle().Foreground(TextMuted)
	s.Menu.Scrollbar = lipgloss.NewStyle().Foreground(BlueGray)
	s.Menu.HelpContext = lipgloss.NewStyle().Foreground(BlueGray)

	return s
}
