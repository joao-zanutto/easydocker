package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Set contains all style primitives used by the root TUI orchestrator.
// Fields are grouped by the screen/component that primarily uses them:
//   - Chrome: Page, Header, Title, TitleMeta, Badge, Tab, ActiveTab, Footer, Key, KeyText
//   - Browse: Divider, Muted, Section, Label, Value, HeaderRow, Row, ActiveRow
//   - Viewer (Logs/Inspect): Breadcrumb, FollowOn, FollowOff, SubpageFrame
//   - Menu/Help: MenuFrame, MenuSelector, MenuItem, MenuDesc, HelpFrame, HelpTitle,
//     HelpSection, HelpCommand, HelpKey, HelpContext, HelpFooter, Scrollbar
//   - States: StateRun, StateWarn, StateStop, StateDead, StateOther
//   - Misc: ErrorText, MonitorBox, ActiveBG
type Set struct {
	Page         lipgloss.Style // Chrome: page background
	Header       lipgloss.Style // Chrome: header frame
	Title        lipgloss.Style // Chrome: title text
	TitleMeta    lipgloss.Style // Chrome: totals/loading metadata
	Breadcrumb   lipgloss.Style // Viewer: breadcrumb text
	Tab          lipgloss.Style // Chrome: inactive tab
	ActiveTab    lipgloss.Style // Chrome: active tab
	Badge        lipgloss.Style // Chrome: scope badge
	Muted        lipgloss.Style // Browse/Viewer: muted text
	MainFrame    lipgloss.Style // Browse: main content frame
	SubpageFrame lipgloss.Style // Viewer: logs/inspect subpage frame
	Divider      lipgloss.Style // Browse/Viewer: horizontal dividers
	HeaderRow    lipgloss.Style // Browse: table header row
	Row          lipgloss.Style // Browse: table rows
	ActiveRow    lipgloss.Style // Browse: selected table row
	Section      lipgloss.Style // Browse: detail section header
	Label        lipgloss.Style // Browse: detail labels
	Value        lipgloss.Style // Browse: detail values
	ErrorText    lipgloss.Style // Chrome: error messages
	Footer       lipgloss.Style // Chrome: footer frame
	Key          lipgloss.Style // Chrome: footer key bindings
	KeyText      lipgloss.Style // Chrome: footer key descriptions
	FollowOn     lipgloss.Style // Viewer: follow mode enabled
	FollowOff    lipgloss.Style // Viewer: follow mode disabled
	MonitorBox   lipgloss.Style // (unused)
	StateRun     lipgloss.Style // Browse: running state color
	StateWarn    lipgloss.Style // Browse: warning state color
	StateStop    lipgloss.Style // Browse: stopped state color
	StateDead    lipgloss.Style // Browse: dead state color
	StateOther   lipgloss.Style // Browse: other state color
	MenuFrame    lipgloss.Style // Menu: menu overlay frame
	MenuSelector lipgloss.Style // Menu: menu selection indicator
	MenuItem     lipgloss.Style // Menu: menu item text
	MenuDesc     lipgloss.Style // Menu: menu item description
	HelpFrame    lipgloss.Style // Help: help overlay frame
	HelpTitle    lipgloss.Style // Help: help title
	HelpSection  lipgloss.Style // Help: help section header
	HelpCommand  lipgloss.Style // Help: help command text
	HelpKey      lipgloss.Style // Help: help key binding
	HelpContext  lipgloss.Style // Help: help context text
	HelpFooter   lipgloss.Style // Help: help footer
	Scrollbar    lipgloss.Style // Menu/Help: scrollbar styling
	ActiveBG     color.Color    // Shared: active background color
}
