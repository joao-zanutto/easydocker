package menu

import "charm.land/lipgloss/v2"

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

const (
	helpGroupNav    = "nav"
	helpGroupEnter  = "enter"
	helpGroupAction = "action"
)

const (
	noteModeLogsInspect     = "mode: logs/inspect"
	noteModeBrowse          = "mode: browse"
	noteModeLogs            = "mode: logs"
	noteTabContainers       = "tab : containers"
	noteRowCompose          = "row : compose"
	noteRowContainer        = "row : container"
	noteRowRunningContainer = "row : running container"
)

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

func NewHelpState(width, height int) HelpState {
	return HelpState{
		Active:   false,
		Cursor:   0,
		Commands: nil,
		Width:    width,
		Height:   height,
	}
}

type HelpKeyLabels struct {
	MoveUpDown   string
	PageUpDown   string
	HomeEnd      string
	TabLeftRight string
	LeftRight    string
	OpenLogs     string
	OpenFilter   string
	ToggleScope  string
	OpenInspect  string
	OpenShell    string
	ToggleWrap   string
	ToggleFollow string
	BackClose    string
}

func BuildHelpCommands(labels HelpKeyLabels) []HelpCommand {
	return []HelpCommand{
		navCommand(labels.MoveUpDown, "move up/down", ""),
		navCommand(labels.PageUpDown, "page up/down", ""),
		navCommand(labels.HomeEnd, "go to top/bottom", noteModeLogsInspect),
		navCommand(labels.TabLeftRight, "prev/next tab", noteModeBrowse),
		navCommand(labels.LeftRight, "scroll left/right", noteModeLogsInspect),
		enterCommand(labels.OpenLogs, "expand/collapse", noteRowCompose),
		enterCommand(labels.OpenLogs, "open logs", noteRowContainer),
		actionCommand(labels.OpenFilter, "filter", ""),
		actionCommand(labels.ToggleScope, "toggle running/all", noteTabContainers),
		actionCommand(labels.OpenInspect, "inspect", noteModeBrowse),
		actionCommand(labels.OpenShell, "interactive shell", noteRowRunningContainer),
		actionCommand(labels.ToggleWrap, "toggle wrap", noteModeLogsInspect),
		actionCommand(labels.ToggleFollow, "toggle follow", noteModeLogs),
		actionCommand(labels.BackClose, "back/close", ""),
	}
}

func navCommand(keyLabel, description, note string) HelpCommand {
	return HelpCommand{Key: keyLabel, Description: description, Note: note, Group: helpGroupNav}
}

func enterCommand(keyLabel, description, note string) HelpCommand {
	return HelpCommand{Key: keyLabel, Description: description, Note: note, Group: helpGroupEnter}
}

func actionCommand(keyLabel, description, note string) HelpCommand {
	return HelpCommand{Key: keyLabel, Description: description, Note: note, Group: helpGroupAction}
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
