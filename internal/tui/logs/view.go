package logs

import (
	"fmt"
	"strings"

	"easydocker/internal/tui/util"

	"github.com/charmbracelet/lipgloss"
)

type ViewStyles struct {
	Breadcrumb   lipgloss.Style
	FollowOn     lipgloss.Style
	FollowOff    lipgloss.Style
	Muted        lipgloss.Style
	Divider      lipgloss.Style
	SubpageFrame lipgloss.Style
}

type ViewModel struct {
	State         State
	ContainerName string
	Width         int
	Height        int
	Styles        ViewStyles
}

func RenderContent(vm ViewModel) string {
	if vm.Width == 0 || vm.Height == 0 {
		return ""
	}
	layout := util.ComputeFrameLayout(vm.Width, vm.Height, vm.Styles.SubpageFrame)
	safeContentWidth := max(1, layout.ContentWidth-2)
	headerVM := vm
	headerVM.Width = safeContentWidth
	breadcrumb := util.ClampSingleLine("Containers / "+vm.ContainerName+" / Logs", safeContentWidth)
	logsHeight := max(1, layout.ContentHeight-3)
	logList := vm.State.Data.Logs
	start, end := ViewportRange(vm.State, len(logList))
	headline := RenderHeader(headerVM, breadcrumb, len(logList), start, end)
	logsPanel := RenderPanel(vm, safeContentWidth, logsHeight)
	return util.RenderFramedContent(vm.Styles.SubpageFrame, layout, util.JoinSections(headline, renderDivider(vm.Styles.Divider, safeContentWidth), logsPanel))
}

func RenderHeader(vm ViewModel, breadcrumb string, total, start, end int) string {
	follow := "off"
	if vm.State.Follow {
		follow = "on"
	}
	first := 0
	last := 0
	if total > 0 {
		first = start + 1
		last = max(first, end)
	}
	left := vm.Styles.Breadcrumb.Render(breadcrumb)
	followText := vm.Styles.FollowOff.Render(follow)
	if vm.State.Follow {
		followText = vm.Styles.FollowOn.Render(follow)
	}
	right := lipgloss.JoinHorizontal(
		lipgloss.Left,
		vm.Styles.Muted.Render("follow:"),
		followText,
		vm.Styles.Muted.Render(fmt.Sprintf("  lines:(%d-%d/%d)", first, last, total)),
	)
	return renderRightPriorityLine(left, right, vm.Width)
}

func RenderPanel(vm ViewModel, width, height int) string {
	contentWidth := max(1, width-2)
	if vm.State.InitialLoad {
		return strings.Join(util.ClipAndPadLines([]string{renderLoadingLine(vm.Styles.Muted, contentWidth)}, height, ""), "\n")
	}

	logList := vm.State.Data.Logs
	if len(logList) == 0 {
		return strings.Join(util.ClipAndPadLines([]string{util.ClampSingleLine(vm.Styles.Muted.Render("No logs found for this container."), contentWidth)}, height, ""), "\n")
	}

	lines := strings.Split(vm.State.Viewport.View(), "\n")
	if vm.State.HistoryLoad {
		lines = append([]string{renderHistoryLoadingLine(vm.Styles.Muted, contentWidth)}, lines...)
	}
	lines = util.ClipAndPadLines(lines, height, "")
	return strings.Join(lines, "\n")
}

func renderDivider(style lipgloss.Style, width int) string {
	return style.Render(strings.Repeat("─", max(1, width-2)))
}

func renderLoadingLine(style lipgloss.Style, width int) string {
	return util.ClampSingleLine(style.Render("⟳ Loading logs..."), width)
}

func renderHistoryLoadingLine(style lipgloss.Style, width int) string {
	return util.ClampSingleLine(style.Render("⇡ Loading older logs..."), width)
}

func renderRightPriorityLine(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	right = util.ClampSingleLine(right, width)
	rightWidth := util.DisplayWidth(right)
	if rightWidth >= width {
		return right
	}
	leftWidth := max(1, width-rightWidth-1)
	left = util.ClampSingleLine(left, leftWidth)
	leftRenderedWidth := util.DisplayWidth(left)
	spacing := max(0, width-leftRenderedWidth-rightWidth)
	return left + strings.Repeat(" ", spacing) + right
}
