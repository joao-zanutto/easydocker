package menu

import (
	"strings"

	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/viewer"

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
		Commands: buildHelpCommands(),
		Width:    width,
		Height:   height,
	}
}

func buildHelpCommands() []HelpCommand {
	browseKeys := browse.NewKeyMap()
	viewerKeys := viewer.NewKeyMap()
	menuKeys := NewKeyMap()

	return []HelpCommand{
		navCommand(joinKeyLabels("/", browseKeys.MoveUp, browseKeys.MoveDown), "move up/down", ""),
		navCommand(joinKeyLabels("/", viewerKeys.PageUp, viewerKeys.PageDown), "page up/down", ""),
		navCommand(joinKeyLabels("/", viewerKeys.Home, viewerKeys.End), "go to top/bottom", noteModeLogsInspect),
		navCommand(joinKeyLabels("/", browseKeys.TabLeft, browseKeys.TabRight), "prev/next tab", noteModeBrowse),
		navCommand(joinKeyLabels("/", viewerKeys.Left, viewerKeys.Right), "scroll left/right", noteModeLogsInspect),
		enterCommand(keyLabel(browseKeys.OpenLogs), "expand/collapse", noteRowCompose),
		enterCommand(keyLabel(browseKeys.OpenLogs), "open logs", noteRowContainer),
		actionCommand(keyLabel(browseKeys.OpenFilter), "filter", ""),
		actionCommand(keyLabel(browseKeys.ToggleScope), "toggle running/all", noteTabContainers),
		actionCommand(keyLabel(browseKeys.OpenInspect), "inspect", noteModeBrowse),
		actionCommand(keyLabel(browseKeys.OpenShell), "interactive shell", noteRowRunningContainer),
		actionCommand(keyLabel(viewerKeys.ToggleWrap), "toggle wrap", noteModeLogsInspect),
		actionCommand(keyLabel(viewerKeys.ToggleFollow), "toggle follow", noteModeLogs),
		actionCommand(keyLabel(menuKeys.Back), "back/close", ""),
	}
}

func keyLabel(binding key.Binding) string {
	return strings.TrimSpace(binding.Help().Key)
}

func joinKeyLabels(sep string, bindings ...key.Binding) string {
	labels := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		label := keyLabel(binding)
		if label == "" {
			continue
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, sep)
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
