package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type ChromeStyles struct {
	Page      lipgloss.Style
	Header    lipgloss.Style
	Title     lipgloss.Style
	TitleMeta lipgloss.Style
	Badge     lipgloss.Style
	ErrorText lipgloss.Style
	Footer    lipgloss.Style
	Key       lipgloss.Style
	KeyText   lipgloss.Style
}

type TabStyles struct {
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style
}

type TableStyles struct {
	HeaderRow lipgloss.Style
	Row       lipgloss.Style
	ActiveRow lipgloss.Style
}

type BrowseStyles struct {
	MainFrame lipgloss.Style
	Divider   lipgloss.Style
	Muted     lipgloss.Style
	Section   lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
}

type ViewerStyles struct {
	Breadcrumb   lipgloss.Style
	FollowOn     lipgloss.Style
	FollowOff    lipgloss.Style
	SubpageFrame lipgloss.Style
}

type MenuStyles struct {
	Frame           lipgloss.Style
	Selector        lipgloss.Style
	ItemNormal      lipgloss.Style
	ItemDescription lipgloss.Style
	HelpFrame       lipgloss.Style
	HelpTitle       lipgloss.Style
	HelpSection     lipgloss.Style
	HelpCommand     lipgloss.Style
	HelpKey         lipgloss.Style
	HelpContext     lipgloss.Style
	HelpFooter      lipgloss.Style
	Scrollbar       lipgloss.Style
}

type StateStyles struct {
	StateRun   lipgloss.Style
	StateWarn  lipgloss.Style
	StateStop  lipgloss.Style
	StateDead  lipgloss.Style
	StateOther lipgloss.Style
}

type Set struct {
	Chrome   ChromeStyles
	Tabs     TabStyles
	Tables   TableStyles
	Browse   BrowseStyles
	Viewer   ViewerStyles
	Menu     MenuStyles
	States   StateStyles
	ActiveBG color.Color
}
