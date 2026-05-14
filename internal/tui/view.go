package tui

import (
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/ui/chrome"
	"easydocker/internal/tui/ui/tables"
	"easydocker/internal/tui/util"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
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
	if m.screen == screenModeLogs && m.activeTab == tabContainers {
		container, ok := m.selectedLogsContainer()
		if !ok {
			return m.styles.ErrorText.Render("Selected container is no longer available.")
		}

		logList := viewer.FilterLines(m.logs.Data, m.logs.Filter.Query)
		start, end := viewer.VisibleContentRange(&m.logs, logList)

		return viewer.RenderContent(viewer.ViewModel{
			State:            &m.logs,
			ContainerName:    container.Name,
			Breadcrumb:       "Containers / " + container.Name + " / Logs",
			LineCount:        &viewer.LineCountInfo{Total: len(logList), Start: start + 1, End: max(start+1, end)},
			LoadingMessage:   "Loading logs...",
			EmptyMessage:     "No logs found for this container.",
			LoadingIndicator: m.logsLoadingIndicator(),
			Width:            totalWidth,
			Height:           totalHeight,
			ContentType:      viewer.ContentTypeLogs,
			ResourceType:     viewer.ResourceTypeContainer,
			Styles: viewer.ViewStyles{
				Breadcrumb:   m.styles.Breadcrumb,
				FollowOn:     m.styles.FollowOn,
				FollowOff:    m.styles.FollowOff,
				Muted:        m.styles.Muted,
				Divider:      m.styles.Divider,
				SubpageFrame: m.styles.SubpageFrame,
			},
		})
	}

	if m.screen == screenModeInspect {
		return m.renderInspectContent(totalWidth, totalHeight)
	}

	layout := util.ComputeFrameLayout(totalWidth, totalHeight, m.styles.MainFrame)
	content := m.renderBrowseContent(layout.ContentWidth, layout.ContentHeight)
	return util.RenderFramedContent(m.styles.MainFrame, layout, content)
}

func resourceTypeFromTab(tab int) viewer.ResourceType {
	switch tab {
	case tabContainers:
		return viewer.ResourceTypeContainer
	case tabImages:
		return viewer.ResourceTypeImage
	case tabNetworks:
		return viewer.ResourceTypeNetwork
	case tabVolumes:
		return viewer.ResourceTypeVolume
	default:
		return viewer.ResourceTypeContainer
	}
}

func (m model) renderInspectContent(totalWidth, totalHeight int) string {
	resourceLabel := viewer.GetResourceLabel(resourceTypeFromTab(m.activeTab))
	containerName := m.logs.ResourceName
	breadcrumb := resourceLabel + " / " + containerName + " / Inspect"

	logList := viewer.FilterLines(m.logs.Data, m.logs.Filter.Query)
	start, end := viewer.VisibleContentRange(&m.logs, logList)

	return viewer.RenderContent(viewer.ViewModel{
		State:            &m.logs,
		ContainerName:    containerName,
		Breadcrumb:       breadcrumb,
		LineCount:        &viewer.LineCountInfo{Total: len(logList), Start: start + 1, End: max(start+1, end)},
		LoadingMessage:   "Loading inspect...",
		EmptyMessage:     "No inspect data available.",
		LoadingIndicator: m.logsLoadingIndicator(),
		Width:            totalWidth,
		Height:           totalHeight,
		ContentType:      viewer.ContentTypeInspect,
		ResourceType:     resourceTypeFromTab(m.activeTab),
		Styles: viewer.ViewStyles{
			Breadcrumb:   m.styles.Breadcrumb,
			FollowOn:     m.styles.FollowOn,
			FollowOff:    m.styles.FollowOff,
			Muted:        m.styles.Muted,
			Divider:      m.styles.Divider,
			SubpageFrame: m.styles.SubpageFrame,
		},
	})
}

func (m model) logsLoadingIndicator() string {
	if !m.shouldAnimateLogsLoadingIndicator() {
		return ""
	}
	return strings.TrimSpace(m.logsSpinner.View())
}

func (m model) logVisibleRows() int {
	return max(1, m.logSectionHeight())
}

func (m model) logVisibleWidth() int {
	totalWidth := max(1, m.width)
	if m.screen == screenModeLogs {
		return m.logsPageContentWidth(totalWidth)
	}
	innerWidth := util.FrameContentWidth(totalWidth, m.styles.MainFrame)
	return max(1, innerWidth)
}

func (m model) logSectionHeight() int {
	mainHeight := util.MainAreaHeight(m.height, m.renderHeader(), m.renderFooter())
	if m.screen == screenModeLogs {
		return viewer.VisibleRowsForContent(m.logsPageContentHeight(mainHeight), m.logs.Filter.Active)
	}
	innerHeight := util.FrameContentHeight(mainHeight, m.styles.MainFrame)
	return max(1, innerHeight)
}

func (m model) logsPageContentWidth(width int) int {
	return util.FrameContentWidth(width, m.styles.SubpageFrame)
}

func (m model) logsPageContentHeight(height int) int {
	return util.FrameContentHeight(height, m.styles.SubpageFrame)
}

func (m model) renderHeader() string {
	return chrome.RenderHeader(chrome.HeaderInput{
		Width:            m.width,
		Title:            "EasyDocker",
		TotalsText:       chrome.RenderTotalsLabel(m.snapshot, m.loadingStage, loadStageIdle, loadStageMetrics, m.metricsLoaded, m.metricsLoadingIndicator()),
		LoadingStageText: chrome.RenderLoadingStageLabel(m.loadingStage, loadStageContainers, loadStageResources, loadStageMetrics, m.metricsLoaded),
		ActiveTab:        m.activeTab,
		ShowAll:          m.showAll,
		Err:              m.err,
		Tabs: []chrome.TabSpec{
			{Tab: tabContainers, Icon: "🐳", Name: "Containers", Count: len(m.filteredContainers())},
			{Tab: tabImages, Icon: "💿", Name: "Images", Count: len(m.snapshot.Images)},
			{Tab: tabNetworks, Icon: "🔌", Name: "Networks", Count: len(m.snapshot.Networks)},
			{Tab: tabVolumes, Icon: "📂", Name: "Volumes", Count: len(m.snapshot.Volumes)},
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

func (m model) renderChromeTab(tab int, label string) string {
	if m.activeTab == tab {
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

func (m model) renderBrowseContent(width, height int) string {
	filterCopy := m.browseFilter
	if filterCopy.Active {
		filterCopy.Input.SetWidth(max(1, width-util.DisplayWidth(filterCopy.Input.Prompt)))
	}

	return browse.RenderContent(browse.ViewModel{
		Loading:                 m.loading,
		Snapshot:                m.snapshot,
		ActiveTab:               m.activeTab,
		MetricsLoadingIndicator: m.containerMetricsLoadingIndicator(),
		Width:                   width,
		Height:                  height,
		Styles: browse.ViewStyles{
			Divider: m.styles.Divider,
			Muted:   m.styles.Muted,
			Section: m.styles.Section,
		},
		Selections: m.browseSelections(),
		Filter:     filterCopy,
	}, m.renderResourceList(width, browse.ListHeightForContent(height, m.browseFilter.Active)), m.browseDetailRenderer())
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

func (m model) browseSelections() browse.SelectionSet {
	container, hasContainer := m.selectedContainer()
	composeProject, hasComposeProject := m.selectedComposeProject()
	image, hasImage := m.selectedImage()
	network, hasNetwork := m.selectedNetwork()
	volume, hasVolume := m.selectedVolume()
	return browse.SelectionSet{
		Container:         container,
		HasContainer:      hasContainer,
		ComposeProject:    composeProject,
		HasComposeProject: hasComposeProject,
		Image:             image,
		HasImage:          hasImage,
		Network:           network,
		HasNetwork:        hasNetwork,
		Volume:            volume,
		HasVolume:         hasVolume,
	}
}

type browseDetailRenderer struct{ model }

func (r browseDetailRenderer) DetailLine(label, value string, width int) string {
	return r.model.detailLineWithWidth(label, value, width)
}

func (r browseDetailRenderer) RenderContainerState(container core.ContainerRow) string {
	return r.model.stateStyle(container.State).Render(browse.ContainerStateText(container))
}

func (m model) browseDetailRenderer() browse.DetailProvider {
	return browseDetailRenderer{model: m}
}

func (m model) stateStyle(state string) lipgloss.Style {
	switch strings.ToLower(state) {
	case "running":
		return m.styles.StateRun
	case "paused", "restarting", "created":
		return m.styles.StateWarn
	case "exited", "stopped":
		return m.styles.StateStop
	case "dead":
		return m.styles.StateDead
	default:
		return m.styles.StateOther
	}
}

func (m model) renderResourceList(width, height int) string {
	switch m.activeTab {
	case tabContainers:
		spec := tables.BuildContainerSpec(width, m.containerCursor, m.containerListRows(), m.activeTab == tabContainers, m.containerMetricsLoadingIndicator())
		return renderResourceTableFromSpec(m, width, height, spec)
	case tabImages:
		spec := tables.BuildImageSpec(width, m.imageCursor, m.filteredImages())
		return renderResourceTableFromSpec(m, width, height, spec)
	case tabNetworks:
		spec := tables.BuildNetworkSpec(width, m.networkCursor, m.filteredNetworks())
		return renderResourceTableFromSpec(m, width, height, spec)
	default:
		spec := tables.BuildVolumeSpec(width, m.volumeCursor, m.filteredVolumes())
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
