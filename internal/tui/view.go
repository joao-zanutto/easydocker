package tui

import (
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/chrome"
	"easydocker/internal/tui/ui/tables"
	"easydocker/internal/tui/util"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	m = m.syncBrowseData()
	content := "Loading EasyDocker..."
	if m.width > 0 && m.height > 0 {
		header := m.renderHeader()
		footer := m.renderFooter()
		mainHeight := util.MainAreaHeight(m.height, header, footer)
		main := m.renderMain(mainHeight)

		baseContent := lipgloss.JoinVertical(lipgloss.Left, header, main, footer)

		styles := menu.DefaultMenuStyles(
			m.styles.MenuFrame,
			m.styles.MenuSelector,
			m.styles.MenuItem,
			m.styles.MenuDesc,
			m.styles.HelpFrame,
			m.styles.HelpTitle,
			m.styles.HelpSection,
			m.styles.HelpCommand,
			m.styles.HelpKey,
			m.styles.HelpContext,
			m.styles.HelpFooter,
			m.styles.Scrollbar,
		)
		content = menu.Render(baseContent, m.menu, m.help, styles, m.width, m.height)
		content = m.styles.Page.Render(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) renderMain(height int) string {
	totalWidth := max(1, m.width)
	totalHeight := max(1, height)

	if m.screen == shared.LogViewer && m.browse.ActiveTab == tabContainers {
		container, ok := m.selectedLogsContainer()
		if !ok {
			return m.styles.ErrorText.Render("Selected container is no longer available.")
		}
		return m.renderLogsContent(container, totalWidth, totalHeight)
	}

	if m.screen == shared.InspectViewer {
		return m.renderInspectContent(totalWidth, totalHeight)
	}

	layout := util.ComputeFrameLayout(totalWidth, totalHeight, m.styles.MainFrame)
	m.browse.Width = layout.ContentWidth
	m.browse.Height = layout.ContentHeight

	content := m.browse.View()
	return util.RenderFramedContent(m.styles.MainFrame, layout, content)
}

func (m model) renderLogsContent(container core.ContainerRow, totalWidth, totalHeight int) string {
	m.viewer.Width = totalWidth
	m.viewer.Height = totalHeight
	m.viewer.ContainerName = container.Name
	m.viewer.Breadcrumb = "Containers / " + container.Name + " / Logs"
	m.viewer.ContentType = viewer.ContentTypeLogs
	m.viewer.ResourceType = core.ResourceContainer
	m.viewer.LoadingMsg = "Loading logs..."
	m.viewer.EmptyMsg = "No logs found for this container."
	m.viewer.Styles = viewer.Styles{
		Breadcrumb:   m.styles.Breadcrumb,
		FollowOn:     m.styles.FollowOn,
		FollowOff:    m.styles.FollowOff,
		Muted:        m.styles.Muted,
		Divider:      m.styles.Divider,
		SubpageFrame: m.styles.SubpageFrame,
	}
	return m.viewer.View()
}

func (m model) renderInspectContent(totalWidth, totalHeight int) string {
	resourceLabel := core.ResourceLabel(shared.TabToResourceType(m.browse.ActiveTab))
	containerName := m.viewer.ResourceName
	breadcrumb := resourceLabel + " / " + containerName + " / Inspect"

	m.viewer.Width = totalWidth
	m.viewer.Height = totalHeight
	m.viewer.ContainerName = containerName
	m.viewer.Breadcrumb = breadcrumb
	m.viewer.ContentType = viewer.ContentTypeInspect
	m.viewer.ResourceType = shared.TabToResourceType(m.browse.ActiveTab)
	m.viewer.LoadingMsg = "Loading inspect..."
	m.viewer.EmptyMsg = "No inspect data available."
	m.viewer.Styles = viewer.Styles{
		Breadcrumb:   m.styles.Breadcrumb,
		FollowOn:     m.styles.FollowOn,
		FollowOff:    m.styles.FollowOff,
		Muted:        m.styles.Muted,
		Divider:      m.styles.Divider,
		SubpageFrame: m.styles.SubpageFrame,
	}
	return m.viewer.View()
}

func (m model) metricsLoadingIndicator() string {
	if !m.shouldAnimateMetricsLoadingIndicator() {
		return ""
	}
	return strings.TrimSpace(m.metricsSpinner.View())
}

func (m model) containerMetricsLoadingIndicator() string {
	if !m.shouldAnimateMetricsLoadingIndicator() {
		return ""
	}
	return strings.TrimSpace(m.containerSpinner.View())
}

func (m model) renderHeader() string {
	return chrome.RenderHeader(chrome.HeaderInput{
		Width:            m.width,
		Title:            "EasyDocker",
		TotalsText:       chrome.RenderTotalsLabel(m.browse.Snapshot, m.loadingStage, m.metricsLoaded, m.metricsLoadingIndicator()),
		LoadingStageText: chrome.RenderLoadingStageLabel(m.loadingStage, m.metricsLoaded),
		ActiveTab:        m.browse.ActiveTab,
		ShowAll:          m.browse.ShowAll,
		Err:              m.err,
		Tabs: []chrome.TabSpec{
			{Tab: tabContainers, Icon: "🐳", Name: "Containers", Count: len(m.filteredContainers())},
			{Tab: tabImages, Icon: "💿", Name: "Images", Count: len(m.browse.Snapshot.Images)},
			{Tab: tabNetworks, Icon: "🔌", Name: "Networks", Count: len(m.browse.Snapshot.Networks)},
			{Tab: tabVolumes, Icon: "📂", Name: "Volumes", Count: len(m.browse.Snapshot.Volumes)},
		},
		Styles: chrome.HeaderStyles{
			Header:    m.styles.Header,
			Title:     m.styles.Title,
			TitleMeta: m.styles.TitleMeta,
			Badge:     m.styles.Badge,
			ErrorText: m.styles.ErrorText,
		},
		RenderTab: m.renderChromeTab,
	})
}

func (m model) renderFooter() string {
	return chrome.RenderFooter(chrome.FooterInput{
		Width:  m.width,
		KeyMap: m.footerKeyMap(),
		Styles: chrome.FooterStyles{
			Footer:  m.styles.Footer,
			Key:     m.styles.Key,
			KeyText: m.styles.KeyText,
		},
	})
}

func (m model) renderChromeTab(tab shared.Tab, label string) string {
	if m.browse.ActiveTab == tab {
		return m.styles.ActiveTab.Render(label)
	}
	return m.styles.Tab.Render(label)
}

func (m model) detailLineWithWidth(label, value string, width int) string {
	labelText := label + ": "
	if width <= 0 {
		return m.styles.Label.Render(labelText) + m.styles.Value.Render(value)
	}

	labelRendered := m.styles.Label.Render(labelText)
	labelWidth := util.DisplayWidth(labelRendered)
	if labelWidth >= width {
		return util.ConstrainLine(labelRendered, width)
	}

	valueWidth := max(1, width-labelWidth)
	return labelRendered + m.styles.Value.Render(util.ConstrainLine(value, valueWidth))
}

func (m model) browseDetailRenderer() browse.DetailProvider {
	return browseDetailRenderer{
		detailLine:       m.detailLineWithWidth,
		containerStateFn: m.renderContainerState,
	}
}

func (m model) renderContainerState(container core.ContainerRow) string {
	return m.stateStyle(container.State).Render(core.ContainerStateText(container))
}

func (m model) stateStyle(state core.ContainerState) lipgloss.Style {
	switch state {
	case core.StateRunning:
		return m.styles.StateRun
	case core.StatePaused, core.StateRestarting, core.StateCreated:
		return m.styles.StateWarn
	case core.StateExited:
		return m.styles.StateStop
	case core.StateDead:
		return m.styles.StateDead
	default:
		return m.styles.StateOther
	}
}

func (m model) renderResourceList(width, height int) string {
	switch m.browse.ActiveTab {
	case tabContainers:
		spec := tables.BuildContainerSpec(width, m.browse.ContainerCursor, m.browse.Data.ContainerListRows, m.browse.ActiveTab == tabContainers, m.browse.Data.MetricsLoadingIndicator)
		return renderResourceTableFromSpec(m, width, height, spec)
	case tabImages:
		spec := tables.BuildImageSpec(width, m.browse.ImageCursor, m.browse.Data.FilteredImages)
		return renderResourceTableFromSpec(m, width, height, spec)
	case tabNetworks:
		spec := tables.BuildNetworkSpec(width, m.browse.NetworkCursor, m.browse.Data.FilteredNetworks)
		return renderResourceTableFromSpec(m, width, height, spec)
	default:
		spec := tables.BuildVolumeSpec(width, m.browse.VolumeCursor, m.browse.Data.FilteredVolumes)
		return renderResourceTableFromSpec(m, width, height, spec)
	}
}

func renderResourceTableFromSpec[T any](m model, width, height int, spec tables.Spec[T]) string {
	tableStyles := tables.DefaultStyles()
	tableStyles.Header = m.styles.HeaderRow.Inline(true)
	tableStyles.Cell = m.styles.Row.Inline(true)
	tableStyles.Selected = m.styles.ActiveRow.Bold(true).Inline(true)
	return tables.RenderFromSpec(width, height, spec, tableStyles)
}

type browseDetailRenderer struct {
	detailLine       func(label, value string, width int) string
	containerStateFn func(container core.ContainerRow) string
}

func (r browseDetailRenderer) DetailLine(label, value string, width int) string {
	return r.detailLine(label, value, width)
}

func (r browseDetailRenderer) RenderContainerState(container core.ContainerRow) string {
	return r.containerStateFn(container)
}
