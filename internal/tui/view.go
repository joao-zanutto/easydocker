package tui

import (
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/chrome"
	"easydocker/internal/tui/ui/components"
	"easydocker/internal/tui/ui/tables"
	"easydocker/internal/tui/ui/theme"
	"easydocker/internal/tui/util"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	if m.dataDirty {
		m = m.syncBrowseData()
	}
	content := "Loading EasyDocker..."
	if m.width > 0 && m.height > 0 {
		header := m.renderHeader()
		footer := m.renderFooter()
		mainHeight := util.MainAreaHeight(m.height, header, footer)
		main := m.renderMain(mainHeight)

		baseContent := lipgloss.JoinVertical(lipgloss.Left, header, main, footer)

		styles := menu.DefaultMenuStyles(
			m.styles.Menu.Frame,
			m.styles.Menu.Selector,
			m.styles.Menu.ItemNormal,
			m.styles.Menu.ItemDescription,
			m.styles.Menu.HelpFrame,
			m.styles.Menu.HelpTitle,
			m.styles.Menu.HelpSection,
			m.styles.Menu.HelpCommand,
			m.styles.Menu.HelpKey,
			m.styles.Menu.HelpContext,
			m.styles.Menu.HelpFooter,
			m.styles.Menu.Scrollbar,
		)
		content = menu.Render(baseContent, m.menu, m.help, styles, m.width, m.height)
		content = m.styles.Chrome.Page.Render(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) renderMain(height int) string {
	totalWidth := max(1, m.width)
	totalHeight := max(1, height)

	if m.screen == shared.LogViewer && m.browse.ActiveTab == shared.TabContainers {
		container, ok := m.selectedLogsContainer()
		if !ok {
			return m.styles.Chrome.ErrorText.Render("Selected container is no longer available.")
		}
		return m.renderLogsContent(container, totalWidth, totalHeight)
	}

	if m.screen == shared.InspectViewer && m.viewer.Vp.ContentType == viewer.ContentTypeConfig {
		return m.renderConfigContent(totalWidth, totalHeight)
	}

	if m.screen == shared.InspectViewer {
		return m.renderInspectContent(totalWidth, totalHeight)
	}

	contentWidth := util.FrameContentWidth(totalWidth, m.styles.Browse.MainFrame)
	contentHeight := util.FrameContentHeight(totalHeight, m.styles.Browse.MainFrame)
	m.browse.Width = contentWidth
	m.browse.Height = contentHeight

	m.browse.RenderedList = m.renderResourceList(m.browse.Width, browse.ListHeightForContent(m.browse.Height))

	content := m.browse.View()
	return util.RenderInFrame(m.styles.Browse.MainFrame, content, totalWidth, totalHeight)
}

func (m model) renderViewer(totalWidth, totalHeight int, opts ...viewerOption) string {
	m.viewer.Width = totalWidth
	m.viewer.Height = totalHeight
	m.viewer.Styles = m.viewerStyles()
	for _, opt := range opts {
		opt(&m.viewer)
	}
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	return m.viewer.View()
}

type viewerOption func(*viewer.Model)

func viewerContentType(ct viewer.ContentType) viewerOption {
	return func(vm *viewer.Model) { vm.Vp.ContentType = ct }
}

func viewerResourceType(rt core.ResourceType) viewerOption {
	return func(vm *viewer.Model) { vm.ResourceType = rt }
}

func viewerBreadcrumb(b string) viewerOption {
	return func(vm *viewer.Model) { vm.Breadcrumb = b }
}

func viewerContainerName(name string) viewerOption {
	return func(vm *viewer.Model) { vm.ContainerName = name }
}

func viewerLoadingMsg(msg string) viewerOption {
	return func(vm *viewer.Model) { vm.LoadingMsg = msg }
}

func viewerEmptyMsg(msg string) viewerOption {
	return func(vm *viewer.Model) { vm.EmptyMsg = msg }
}

func (m model) renderLogsContent(container core.ContainerRow, totalWidth, totalHeight int) string {
	return m.renderViewer(totalWidth, totalHeight,
		viewerContainerName(container.Name),
		viewerBreadcrumb("Containers > "+container.Name+" > Logs"),
		viewerContentType(viewer.ContentTypeLogs),
		viewerResourceType(core.ResourceContainer),
		viewerLoadingMsg("Loading logs..."),
		viewerEmptyMsg("No logs found for this container."),
	)
}

func (m model) renderInspectContent(totalWidth, totalHeight int) string {
	resourceLabel := util.ResourceLabel(shared.TabToResourceType(m.browse.ActiveTab))
	containerName := m.viewer.Inspect.ResourceName
	return m.renderViewer(totalWidth, totalHeight,
		viewerContainerName(containerName),
		viewerBreadcrumb(resourceLabel+" > "+containerName+" > Inspect"),
		viewerContentType(viewer.ContentTypeInspect),
		viewerResourceType(shared.TabToResourceType(m.browse.ActiveTab)),
		viewerLoadingMsg("Loading inspect..."),
		viewerEmptyMsg("No inspect data available."),
	)
}

func (m model) renderConfigContent(totalWidth, totalHeight int) string {
	return m.renderViewer(totalWidth, totalHeight,
		viewerContainerName(""),
		viewerBreadcrumb("Configuration"),
		viewerContentType(viewer.ContentTypeConfig),
		viewerLoadingMsg("Loading..."),
		viewerEmptyMsg("No configuration data available."),
	)
}

func (m model) viewerStyles() viewer.Styles {
	return viewer.Styles{
		Breadcrumb:   m.styles.Viewer.Breadcrumb,
		FollowOn:     m.styles.Viewer.FollowOn,
		FollowOff:    m.styles.Viewer.FollowOff,
		Muted:        m.styles.Browse.Muted,
		Divider:      m.styles.Browse.Divider,
		SubpageFrame: m.styles.Viewer.SubpageFrame,
		Key:          m.styles.Chrome.Key,
		KeyText:      m.styles.Chrome.KeyText,
	}
}

func (m model) metricsLoadingIndicator() string {
	if !m.shouldAnimateMetricsLoadingIndicator() {
		return ""
	}
	return strings.TrimSpace(m.spinner.View())
}

func (m model) renderHeader() string {
	isViewer := m.screen == shared.LogViewer || m.screen == shared.InspectViewer
	return chrome.RenderHeader(chrome.HeaderInput{
		Width:            m.width,
		Title:            "EasyDocker",
		LoadingStageText: chrome.RenderLoadingStageLabel(m.loadingStage, m.metricsLoaded),
		ActiveTab:        m.browse.ActiveTab,
		ShowAll:          m.browse.ShowAll,
		HideScope:        isViewer,
		DimTabs:          isViewer,
		Err:              m.err,
		Tabs: []chrome.TabSpec{
			{Tab: shared.TabContainers, Icon: "🐳", Name: "Containers", Count: len(m.browse.Snapshot.Containers)},
			{Tab: shared.TabImages, Icon: "💿", Name: "Images", Count: len(m.browse.Snapshot.Images)},
			{Tab: shared.TabNetworks, Icon: "🔌", Name: "Networks", Count: len(m.browse.Snapshot.Networks)},
			{Tab: shared.TabVolumes, Icon: "📂", Name: "Volumes", Count: len(m.browse.Snapshot.Volumes)},
		},
		Styles: chrome.HeaderStyles{
			Header:    m.styles.Chrome.Header,
			Title:     m.styles.Chrome.Title,
			TitleMeta: m.styles.Chrome.TitleMeta,
			Badge:     m.styles.Chrome.Badge,
			ErrorText: m.styles.Chrome.ErrorText,
			Key:       m.styles.Chrome.Key,
			KeyText:   m.styles.Chrome.KeyText,
		},
		RenderTab:        m.renderChromeTab,
		Snapshot:         m.browse.Snapshot,
		LoadingStage:     m.loadingStage,
		MetricsLoaded:    m.metricsLoaded,
		LoadingIndicator: m.metricsLoadingIndicator(),
	})
}

func (m model) renderFooter() string {
	return chrome.RenderFooter(chrome.FooterInput{
		Width:  m.width,
		KeyMap: m.footerKeyMap(),
		Styles: chrome.FooterStyles{
			Footer:  m.styles.Chrome.Footer,
			Key:     m.styles.Chrome.Key,
			KeyText: m.styles.Chrome.KeyText,
		},
	})
}

func (m model) renderChromeTab(tab shared.Tab, label string) string {
	parts := strings.SplitN(label, " ", 2)
	if len(parts) == 2 {
		text := m.styles.Tabs.Tab.Render(parts[1])
		if m.browse.ActiveTab == tab {
			text = m.styles.Tabs.ActiveTab.Render(parts[1])
		}
		return parts[0] + " " + text
	}
	if m.browse.ActiveTab == tab {
		return m.styles.Tabs.ActiveTab.Render(label)
	}
	return m.styles.Tabs.Tab.Render(label)
}

func (m model) browseDetailRenderer() browse.DetailProvider {
	return browseDetailRenderer{
		labelStyle: m.styles.Browse.Label,
		valueStyle: m.styles.Browse.Value,
	}
}

func (m model) renderResourceList(width, height int) string {
	hasFilter := m.browse.Filter.Active || m.browse.Filter.Query != ""

	var filterHint string
	if !hasFilter {
		filterHint = m.styles.Chrome.KeyText.Render(" ") + m.styles.Chrome.Key.Inline(true).Render(" / ") + m.styles.Chrome.KeyText.Render(" filter")
	}

	switch m.browse.ActiveTab {
	case shared.TabContainers:
		spec := tables.BuildContainerSpec(width, m.browse.Cursors.Container, m.browse.Data.ContainerListRows, m.browse.ActiveTab == shared.TabContainers, m.browse.Data.MetricsLoadingIndicator)
		applyFilterToHeader(m, &spec)
		if !hasFilter && len(spec.Columns) > 0 {
			spec.Columns[0].Header += filterHint
		}
		return renderResourceTableFromSpec(m, width, height, spec)
	case shared.TabImages:
		return renderSimpleResourceTable(m, width, height, m.browse.Cursors.Image, m.browse.Data.FilteredImages, tables.ImageSchema, tables.ImageTableRow, "No images found.", hasFilter, filterHint)
	case shared.TabNetworks:
		return renderSimpleResourceTable(m, width, height, m.browse.Cursors.Network, m.browse.Data.FilteredNetworks, tables.NetworkSchema, tables.NetworkTableRow, "No networks found.", hasFilter, filterHint)
	default:
		return renderSimpleResourceTable(m, width, height, m.browse.Cursors.Volume, m.browse.Data.FilteredVolumes, tables.VolumeSchema, tables.VolumeTableRow, "No volumes found.", hasFilter, filterHint)
	}
}

func renderSimpleResourceTable[T any](m model, width, height int, cursor int, items []T, schema []tables.ColumnDef, rowBuilder func(T) []string, emptyMsg string, hasFilter bool, filterHint string) string {
	tw := tables.ContentWidth(width)
	spec := tables.SimpleSpec(tw, emptyMsg, cursor, items, func(w int) []tables.ColumnDef { return tables.ResolveColumns(w, schema) }, rowBuilder)
	applyFilterToHeader(m, &spec)
	if !hasFilter && len(spec.Columns) > 0 {
		spec.Columns[0].Header += filterHint
	}
	return renderResourceTableFromSpec(m, width, height, spec)
}

func applyFilterToHeader[T any](m model, spec *tables.Spec[T]) {
	if (!m.browse.Filter.Active && m.browse.Filter.Query == "") || len(spec.Columns) == 0 {
		return
	}
	input := m.browse.Filter.Input
	input.SetWidth(components.DynamicInputWidth(input.Prompt, spec.Columns[0].Width))
	spec.Columns[0].Header = components.PadVisibleWidth(input.View(), spec.Columns[0].Width)
}

func renderResourceTableFromSpec[T any](m model, width, height int, spec tables.Spec[T]) string {
	tableStyles := tables.DefaultStyles()
	tableStyles.Header = m.styles.Tables.HeaderRow.Inline(true)
	tableStyles.Cell = m.styles.Tables.Row.Inline(true)
	tableStyles.Selected = m.styles.Tables.ActiveRow.Bold(true).Inline(true)
	return tables.RenderFromSpec(width, height, spec, tableStyles)
}

type browseDetailRenderer struct {
	labelStyle lipgloss.Style
	valueStyle lipgloss.Style
}

func (r browseDetailRenderer) DetailLine(label, value string, width int) string {
	labelText := label + ": "
	if width <= 0 {
		return r.labelStyle.Render(labelText) + r.valueStyle.Render(value)
	}

	labelRendered := r.labelStyle.Render(labelText)
	labelWidth := util.DisplayWidth(labelRendered)
	if labelWidth >= width {
		return util.ConstrainLine(labelRendered, width)
	}

	valueWidth := max(1, width-labelWidth)
	return labelRendered + r.valueStyle.Render(util.ConstrainLine(value, valueWidth))
}

func (r browseDetailRenderer) RenderContainerState(container core.ContainerRow) string {
	return r.valueStyle.Foreground(theme.ContainerStateColor(container.State)).
		Render(util.ContainerStateText(container))
}
