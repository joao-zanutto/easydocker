package tui

import (
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	tea "charm.land/bubbletea/v2"
)

func (m *model) handleInspectKey(msg tea.KeyPressMsg) tea.Cmd {
	return m.handleViewerKey(msg, viewerKeyHandlers{
		visibleWidth: m.inspectVisibleWidth,
		visibleRows:  m.inspectVisibleRows,
		openFilter:   m.openInspectFilter,
		closeFilter:  m.closeInspectFilter,
		exitMode:     m.exitInspectMode,
		postTransition: func() {
			m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
		},
	})
}

func (m *model) handleInspectTransition() (tea.Model, tea.Cmd) {
	resourceType, resourceID, resourceName, ok := m.selectedInspectResource()
	if !ok {
		return m, nil
	}
	m.previousScreen = m.screen
	m.screen = shared.Inspect
	m.logs.Viewport.SetXOffset(0)
	m.logs.InitialLoad = true
	m.logs.Data = nil
	m.logs.ContainerID = resourceID
	m.logs.ResourceName = resourceName
	return m, m.loadInspectCmd(resourceType, resourceID, resourceName)
}

func (m *model) loadInspectCmd(resourceType shared.Tab, resourceID, resourceName string) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		var data []string
		var err error
		switch resourceType {
		case tabContainers:
			data, err = svc.InspectContainer(resourceID)
		case tabImages:
			data, err = svc.InspectImage(resourceID)
		case tabNetworks:
			data, err = svc.InspectNetwork(resourceID)
		case tabVolumes:
			data, err = svc.InspectVolume(resourceID)
		}
		return inspectResultMsg{
			resourceType: int(resourceType),
			resourceID:   resourceID,
			resourceName: resourceName,
			data:         data,
			err:          err,
		}
	}
}

func (m model) handleInspectResultMsg(msg inspectResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.logs.InitialLoad = false
		return m, nil
	}
	m.logs.ApplyContentInitial(msg.data)
	m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
	return m, nil
}

func (m *model) exitInspectMode() {
	m.screen = m.previousScreen
}

func (m model) selectedInspectResource() (shared.Tab, string, string, bool) {
	switch m.activeTab {
	case tabContainers:
		c, ok := m.selectedContainer()
		if !ok {
			return 0, "", "", false
		}
		return tabContainers, c.FullID, c.Name, true
	case tabImages:
		img, ok := m.selectedImage()
		if !ok {
			return 0, "", "", false
		}
		return tabImages, img.ID, img.Tags, true
	case tabNetworks:
		net, ok := m.selectedNetwork()
		if !ok {
			return 0, "", "", false
		}
		return tabNetworks, net.ID, net.Name, true
	case tabVolumes:
		vol, ok := m.selectedVolume()
		if !ok {
			return 0, "", "", false
		}
		return tabVolumes, vol.Name, vol.Name, true
	default:
		return 0, "", "", false
	}
}

func (m model) inspectVisibleWidth() int {
	totalWidth := max(1, m.width)
	return m.logsPageContentWidth(totalWidth)
}

func (m model) inspectVisibleRows() int {
	mainHeight := util.MainAreaHeight(m.height, m.renderHeader(), m.renderFooter())
	contentHeight := m.logsPageContentHeight(mainHeight)
	return viewer.VisibleRowsForContent(contentHeight, m.logs.Filter.Active)
}
