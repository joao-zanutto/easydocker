package menu

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type MenuAction int

const (
	MenuActionHelp MenuAction = iota
	MenuActionQuit
)

type MenuItem struct {
	Label       string
	Description string
	Action      MenuAction
}

type MenuState struct {
	Active bool
	Cursor int
	Items  []MenuItem
}

type HelpState struct {
	Active   bool
	Cursor   int
	Commands []HelpCommand
	Width    int
	Height   int
}

type HelpCommand struct {
	Key         string
	Description string
	Note        string
	Group       string
}

type MenuKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
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

func NewMenuState() MenuState {
	return MenuState{
		Active: false,
		Cursor: 0,
		Items: []MenuItem{
			{Label: "Help", Description: "Show all commands", Action: MenuActionHelp},
			{Label: "Quit", Description: "Exit application", Action: MenuActionQuit},
		},
	}
}

func NewKeyMap() MenuKeyMap {
	return MenuKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "navigate"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "navigate"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

func NewHelpState(width, height int) HelpState {
	return HelpState{
		Active:   false,
		Cursor:   0,
		Commands: buildHelpCommands(),
		Width:    width,
		Height:   height,
	}
}

func buildHelpCommands() []HelpCommand {
	return []HelpCommand{
		{Key: "↑/↓", Description: "move up/down", Note: "", Group: "nav"},
		{Key: "pgup/pgdn", Description: "page up/down", Note: "", Group: "nav"},
		{Key: "home/end", Description: "go to top/bottom", Note: "mode: logs/inspect", Group: "nav"},
		{Key: "←/→", Description: "prev/next tab", Note: "mode: browse", Group: "nav"},
		{Key: "←/→", Description: "scroll left/right", Note: "mode: logs/inspect", Group: "nav"},
		{Key: "enter", Description: "expand/collapse", Note: "row : compose", Group: "enter"},
		{Key: "enter", Description: "open logs", Note: "row : container", Group: "enter"},
		{Key: "/", Description: "filter", Note: "", Group: "action"},
		{Key: "a", Description: "toggle running/all", Note: "tab : containers", Group: "action"},
		{Key: "i", Description: "inspect", Note: "mode: browse", Group: "action"},
		{Key: "s", Description: "interactive shell", Note: "row : running container", Group: "action"},
		{Key: "w", Description: "toggle wrap", Note: "mode: logs/inspect", Group: "action"},
		{Key: "f", Description: "toggle follow", Note: "mode: logs", Group: "action"},
		{Key: "esc", Description: "back/close", Note: "", Group: "action"},
	}
}

func DefaultMenuStyles(frameStyle, selectorStyle, itemStyle, descStyle, helpFrameStyle, helpTitleStyle, helpSectionStyle, helpCmdStyle, helpKeyStyle, helpContextStyle, helpFooterStyle, scrollbarStyle lipgloss.Style) MenuStyles {
	return MenuStyles{
		Frame:           frameStyle,
		Selector:        selectorStyle,
		ItemNormal:      itemStyle,
		ItemDescription: descStyle,
		HelpFrame:       helpFrameStyle,
		HelpTitle:       helpTitleStyle,
		HelpSection:     helpSectionStyle,
		HelpCommand:     helpCmdStyle,
		HelpKey:         helpKeyStyle,
		HelpContext:     helpContextStyle,
		HelpFooter:      helpFooterStyle,
		Scrollbar:       scrollbarStyle,
	}
}
