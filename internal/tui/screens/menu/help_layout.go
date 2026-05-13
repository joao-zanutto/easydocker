package menu

import (
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

const helpHeaderLines = 2 // title + column header

func HelpBodyLineCount(commands []HelpCommand) int {
	return len(helpBodyLines(commands, func(HelpCommand) string { return "" }))
}

func HelpBodyHeight(containerHeight int, frame lipgloss.Style) int {
	contentHeight := util.FrameContentHeight(containerHeight, frame)
	bodyHeight := contentHeight - helpHeaderLines
	if bodyHeight < 0 {
		return 0
	}
	return bodyHeight
}

func helpBodyLines(commands []HelpCommand, format func(HelpCommand) string) []string {
	if len(commands) == 0 {
		commands = buildHelpCommands()
	}

	lines := make([]string, 0, len(commands))
	var prevGroup string
	for _, cmd := range commands {
		if cmd.Group != "" && cmd.Group != prevGroup {
			lines = append(lines, "")
			prevGroup = cmd.Group
		}
		lines = append(lines, format(cmd))
	}
	return lines
}
