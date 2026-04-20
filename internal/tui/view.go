package tui

import (
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/browse"
	"easydocker/internal/tui/chrome"
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/tables"
	"easydocker/internal/tui/util"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading EasyDocker..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	mainHeight := util.MainAreaHeight(m.height, header, footer)
	main := m.renderMain(mainHeight)
	used := lipgloss.Height(header) + lipgloss.Height(main) + lipgloss.Height(footer)
	spacer := ""
	if used < m.height {
		spacer = strings.Repeat("\n", m.height-used)
	}

	return m.styles.Page.Render(lipgloss.JoinVertical(lipgloss.Left, header, main, spacer, footer))
}

func (m model) renderMain(height int) string {
	totalWidth := max(1, m.width)
	totalHeight := max(1, height)
	if m.screen == screenModeLogs && m.activeTab == tabContainers {
		container, ok := m.selectedLogsContainer()
		if !ok {
			return m.styles.ErrorText.Render("Selected container is no longer available.")
		}
		return logs.RenderContent(logs.ViewModel{
			State:            m.logs,
			ContainerName:    container.Name,
			LoadingIndicator: m.logsLoadingIndicator(),
			Width:            totalWidth,
			Height:           totalHeight,
			Styles: logs.ViewStyles{
				Breadcrumb:   m.styles.Breadcrumb,
				FollowOn:     m.styles.FollowOn,
				FollowOff:    m.styles.FollowOff,
				Muted:        m.styles.Muted,
				Divider:      m.styles.Divider,
				SubpageFrame: m.styles.SubpageFrame,
			},
		})
	}

	layout := util.ComputeFrameLayout(totalWidth, totalHeight, m.styles.MainFrame)
	content := m.renderBrowseContent(layout.ContentWidth, layout.ContentHeight)
	return util.RenderFramedContent(m.styles.MainFrame, layout, content)
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
		pageContentWidth := m.logsPageContentWidth(totalWidth)
		return max(1, pageContentWidth-2)
	}
	innerWidth := util.FrameContentWidth(totalWidth, m.styles.MainFrame)
	return max(1, innerWidth-2)
}

func (m model) logSectionHeight() int {
	mainHeight := util.MainAreaHeight(m.height, m.renderHeader(), m.renderFooter())
	if m.screen == screenModeLogs {
		return logs.VisibleRowsForContent(m.logsPageContentHeight(mainHeight), m.logs.FilterActive)
	}
	innerHeight := util.FrameContentHeight(mainHeight, m.styles.MainFrame)
	return max(1, innerHeight-2)
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
	safeContentWidth := max(1, width-2)
	return browse.RenderContent(browse.ViewModel{
		Loading:                 m.loading,
		Snapshot:                m.snapshot,
		ActiveTab:               m.activeTab,
		MetricsLoadingIndicator: m.containerMetricsLoadingIndicator(),
		Width:                   safeContentWidth,
		Height:                  height,
		Styles: browse.ViewStyles{
			Divider: m.styles.Divider,
			Muted:   m.styles.Muted,
			Section: m.styles.Section,
		},
		Selections: m.browseSelections(),
		// Add filter state
		FilterActive: m.browseFilterActive,
		FilterQuery:  m.browseFilterQuery,
		FilterInput:  m.renderBrowseFilterInputView(safeContentWidth),
	}, m.renderResourceList(safeContentWidth, browse.ListHeightForContent(height, m.browseFilterActive)), m.browseDetailRenderer())
}

func (m model) renderBrowseFilterInputView(lineWidth int) string {
	input := m.browseFilterInput
	input.Width = max(1, lineWidth-util.DisplayWidth(input.Prompt))
	return input.View()
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
	image, hasImage := m.selectedImage()
	network, hasNetwork := m.selectedNetwork()
	volume, hasVolume := m.selectedVolume()
	return browse.SelectionSet{
		Container:    container,
		HasContainer: hasContainer,
		Image:        image,
		HasImage:     hasImage,
		Network:      network,
		HasNetwork:   hasNetwork,
		Volume:       volume,
		HasVolume:    hasVolume,
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
		spec := tables.BuildContainerSpec(width, m.containerCursor, m.filteredContainers(), m.activeTab == tabContainers, m.containerMetricsLoadingIndicator())
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
