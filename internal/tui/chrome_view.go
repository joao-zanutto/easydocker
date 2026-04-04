package tui

import (
	"fmt"
	"strings"

	"easydocker/internal/core"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m model) renderHeader() string {
	innerWidth := max(1, m.width-m.styles.header.GetHorizontalFrameSize())
	titleText := "EasyDocker"
	totalsText := m.renderTotalsLabel()
	if m.loadingStage != loadStageIdle {
		totalsText += " " + m.renderLoadingStageLabel()
	}
	tabs := m.renderHeaderTabs(max(1, innerWidth-6))
	tabsText := strings.Join(tabs, " ")
	tabsWidth := ansi.StringWidth(stripANSI(tabsText))
	leftAvail := max(1, innerWidth-tabsWidth-1)
	firstLeft := m.styles.title.Render(constrainLine(titleText, max(1, leftAvail-m.styles.title.GetHorizontalFrameSize())))
	firstLine := m.renderEdgeAlignedLine(firstLeft, tabsText, innerWidth)

	secondRight := ""
	if m.activeTab == tabContainers {
		secondRight = m.renderScopeBadge(max(1, innerWidth/3))
	}
	secondLine := m.renderPinnedHeaderLine(m.styles.titleMeta, totalsText, secondRight, innerWidth)
	line := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
	if m.err != nil {
		line = lipgloss.JoinVertical(lipgloss.Left, line, constrainLine(m.styles.errorText.Render(m.err.Error()), innerWidth))
	}
	return m.styles.header.Render(line)
}

func (m model) renderLoadingStageLabel() string {
	switch m.loadingStage {
	case loadStageContainers:
		return "loading containers"
	case loadStageResources:
		return "loading resources"
	case loadStageMetrics:
		return "loading metrics"
	default:
		return ""
	}
}

func (m model) renderTab(tab int, label string) string {
	if m.activeTab == tab {
		return m.styles.activeTab.Render(label)
	}
	return m.styles.tab.Render(label)
}

func (m model) renderFooter() string {
	items := []string{
		m.helpItem("↑/↓", "navigate"),
		m.helpItem("←/→", "switch tabs"),
		m.helpItem("esc", "quit"),
	}
	if m.activeTab == tabContainers && m.screen != screenModeLogs {
		items = append(items, m.helpItem("a", "toggle running/all"), m.helpItem("enter", "logs"))
	}
	if m.screen == screenModeLogs {
		items = []string{
			m.helpItem("← ↑ ↓ →", "navigate"),
			m.helpItem("pgup/dn", "jump up/down"),
			m.helpItem("home/end", "go to top/bottom"),
			m.helpItem("f", "toggle follow"),
			m.helpItem("esc", "back"),
		}
	}
	innerWidth := max(1, m.width-m.styles.footer.GetHorizontalFrameSize())
	line := constrainLine(strings.Join(items, "   "), innerWidth)
	return m.styles.footer.Render(m.renderCenteredLine(line, innerWidth))
}

func (m model) helpItem(key, description string) string {
	return m.styles.key.Render(key) + " " + m.styles.keyText.Render(description)
}

func (m model) renderTotalsLabel() string {
	if m.loadingStage == loadStageMetrics || (m.loadingStage != loadStageIdle && m.snapshot.TotalCPU == 0 && m.snapshot.TotalMem == 0) {
		return "CPU loading  MEM loading"
	}
	mem := core.HumanBytes(int64(m.snapshot.TotalMem))
	if m.snapshot.TotalLimit > 0 {
		return fmt.Sprintf("CPU %s  MEM %s", renderPercent(m.snapshot.TotalCPU), formatMemoryUsage(mem, (float64(m.snapshot.TotalMem)/float64(m.snapshot.TotalLimit))*100, core.HumanBytes(int64(m.snapshot.TotalLimit))))
	}
	return fmt.Sprintf("CPU %s  MEM %s", renderPercent(m.snapshot.TotalCPU), mem)
}

func (m model) renderHeaderTabs(maxWidth int) []string {
	specs := []struct {
		tab   int
		icon  string
		name  string
		count int
	}{
		{tab: tabContainers, icon: "🐳", name: "Containers", count: len(m.filteredContainers())},
		{tab: tabImages, icon: "💿", name: "Images", count: len(m.snapshot.Images)},
		{tab: tabNetworks, icon: "🔌", name: "Networks", count: len(m.snapshot.Networks)},
		{tab: tabVolumes, icon: "📂", name: "Volumes", count: len(m.snapshot.Volumes)},
	}

	variants := []func(tab int, icon, name string, count int) string{
		func(tab int, icon, name string, count int) string {
			return m.renderTab(tab, fmt.Sprintf("%s %s (%d)", icon, name, count))
		},
		func(tab int, icon, name string, count int) string {
			return m.renderTab(tab, fmt.Sprintf("%s %s %d", icon, name, count))
		},
		func(tab int, icon, name string, count int) string {
			return m.renderTab(tab, fmt.Sprintf("%s %d", icon, count))
		},
	}

	for _, variant := range variants {
		tabs := make([]string, 0, len(specs))
		joinedWidth := 0
		for i, spec := range specs {
			label := variant(spec.tab, spec.icon, spec.name, spec.count)
			tabs = append(tabs, label)
			joinedWidth += ansi.StringWidth(stripANSI(label))
			if i > 0 {
				joinedWidth++
			}
		}
		if joinedWidth <= maxWidth {
			return tabs
		}
	}

	tabs := make([]string, 0, len(specs))
	for _, spec := range specs {
		tabs = append(tabs, m.renderTab(spec.tab, spec.icon))
	}
	return tabs
}

func (m model) renderScopeBadge(maxWidth int) string {
	scope := "running"
	if m.showAll {
		scope = "all"
	}
	labels := []string{
		"container scope: " + scope,
		"scope: " + scope,
		scope,
	}
	for _, label := range labels {
		badge := m.styles.badge.Render(label)
		if ansi.StringWidth(stripANSI(badge)) <= max(1, maxWidth) {
			return badge
		}
	}
	return m.styles.badge.Render(scope)
}
