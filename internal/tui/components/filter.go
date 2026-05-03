package components

import (
	"strings"

	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

const FilterHeaderHeight = 2

type FilterState struct {
	Active bool
	Query  string
	Input  textinput.Model
}

func NewFilterState() FilterState {
	input := textinput.New()
	input.Prompt = "🔎︎ "
	input.Placeholder = ""
	input.CharLimit = 200
	return FilterState{Input: input}
}

func RenderFilterHeader(input string, width int, dividerStyle lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	inputLine := PadVisibleWidth(input, width)
	titleDivider := dividerStyle.Bold(true).Render(strings.Repeat("━", max(1, width)))
	divider := dividerStyle.Render(strings.Repeat("─", max(1, width)))
	return util.JoinSections(titleDivider, inputLine, divider)
}

func RenderTitleDivider(style lipgloss.Style, width int) string {
	return style.Bold(true).Render(strings.Repeat("━", max(1, width)))
}

func RenderFilterInputOnly(input string, width int, dividerStyle lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	inputLine := PadVisibleWidth(input, width)
	divider := dividerStyle.Render(strings.Repeat("─", max(1, width)))
	return util.JoinSections(inputLine, divider)
}

func PadVisibleWidth(line string, width int) string {
	constrained := util.ClampSingleLine(line, width)
	padding := width - util.DisplayWidth(constrained)
	if padding <= 0 {
		return constrained
	}
	return constrained + strings.Repeat(" ", padding)
}

func DynamicInputWidth(prompt string, lineWidth int) int {
	return max(1, lineWidth-util.DisplayWidth(prompt))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
