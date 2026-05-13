package menu

import (
	"strings"

	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

func Render(baseContent string, menuState MenuState, helpState HelpState, styles MenuStyles, width, height int) string {
	overlay := ""
	if helpState.Active {
		helpWidth := width * 7 / 10
		helpHeight := height * 9 / 10
		overlay = RenderHelp(helpState, styles, helpWidth, helpHeight)
	} else if menuState.Active {
		overlay = RenderMenu(menuState, styles, width, height)
	}

	if overlay == "" {
		return baseContent
	}

	menuWidth := lipgloss.Width(overlay)
	menuHeight := lipgloss.Height(overlay)
	centerX := (width - menuWidth) / 2
	centerY := (height - menuHeight) / 2

	if centerX < 0 {
		centerX = 0
	}
	if centerY < 0 {
		centerY = 0
	}

	bgLayer := lipgloss.NewLayer(baseContent)
	fgLayer := lipgloss.NewLayer(overlay).X(centerX).Y(centerY).Z(1)
	comp := lipgloss.NewCompositor(bgLayer, fgLayer)
	cbounds := comp.Bounds()
	canvas := lipgloss.NewCanvas(cbounds.Dx(), cbounds.Dy()).Compose(comp)

	fg := lipgloss.Color("244")
	bg := lipgloss.Color("233")

	for y := 0; y < cbounds.Dy(); y++ {
		for x := 0; x < cbounds.Dx(); x++ {
			if x >= centerX && x < centerX+menuWidth && y >= centerY && y < centerY+menuHeight {
				continue
			}

			cell := canvas.CellAt(x, y)
			if cell == nil || cell.IsZero() {
				continue
			}

			cell = cell.Clone()
			cell.Style.Fg = fg
			cell.Style.Bg = bg
			canvas.SetCell(x, y, cell)
		}
	}

	return canvas.Render()
}

func RenderMenu(state MenuState, styles MenuStyles, width, height int) string {
	if !state.Active || len(state.Items) == 0 {
		return ""
	}

	contentLines := make([]string, len(state.Items))
	for i, item := range state.Items {
		selector := " "
		if i == state.Cursor {
			selector = styles.Selector.Render(">")
		}
		label := item.Label
		if i == state.Cursor {
			label = styles.Selector.Render(item.Label)
		} else {
			label = styles.ItemNormal.Render(item.Label)
		}
		desc := styles.ItemDescription.Render(item.Description)
		contentLines[i] = selector + " " + label + "  " + desc
	}

	content := strings.Join(contentLines, "\n")

	frameWidth := 30
	innerWidth := max(1, frameWidth-styles.Frame.GetHorizontalFrameSize())
	for _, line := range contentLines {
		lineWidth := util.DisplayWidth(line)
		if lineWidth > innerWidth {
			innerWidth = lineWidth
		}
	}
	frameWidth = innerWidth + styles.Frame.GetHorizontalFrameSize()

	styledContent := lipgloss.NewStyle().
		Width(innerWidth).
		Render(content)

	return styles.Frame.Width(frameWidth).Render(styledContent)
}

func RenderHelp(state HelpState, styles MenuStyles, containerWidth, containerHeight int) string {
	if !state.Active {
		return ""
	}

	keyColWidth := 12
	descColWidth := 20
	contextColWidth := 23

	headerLine := styles.HelpKey.Width(keyColWidth).Render("COMMAND") +
		styles.HelpCommand.Width(descColWidth).Render("DESCRIPTION") +
		styles.HelpContext.Width(contextColWidth).Render("CONTEXT")

	bodyLines := helpBodyLines(state.Commands, func(cmd HelpCommand) string {
		keyStr := styles.HelpKey.Width(keyColWidth).Render(cmd.Key)
		descStr := styles.HelpCommand.Width(descColWidth).Render(cmd.Description)
		contextStr := styles.HelpContext.Width(contextColWidth).Render(cmd.Note)
		return keyStr + descStr + contextStr
	})

	innerHeight := util.FrameContentHeight(containerHeight, styles.HelpFrame)
	bodyHeight := max(0, innerHeight-helpHeaderLines)
	maxScroll := max(0, len(bodyLines)-bodyHeight)
	scrollStart := min(max(state.Cursor, 0), maxScroll)

	visibleBody := ""
	if bodyHeight > 0 && len(bodyLines) > 0 {
		end := min(scrollStart+bodyHeight, len(bodyLines))
		visibleBody = strings.Join(bodyLines[scrollStart:end], "\n")
	}

	contentLines := []string{
		styles.HelpTitle.Render("EasyDocker Help"),
		headerLine,
	}
	if visibleBody != "" {
		contentLines = append(contentLines, visibleBody)
	}
	content := strings.Join(contentLines, "\n")

	styledContent := lipgloss.NewStyle().
		Width(keyColWidth + descColWidth + contextColWidth).
		Render(content)

	frame := styles.HelpFrame
	if containerHeight > 0 {
		frame = frame.Height(innerHeight).MaxHeight(containerHeight)
	}
	return frame.Render(styledContent)
}
