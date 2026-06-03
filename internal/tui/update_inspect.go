package tui

import (
	"context"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

func (m *model) handleInspectTransition() (tea.Model, tea.Cmd) {
	resourceType, resourceID, resourceName, ok := m.selectedInspectResource()
	if !ok {
		return m, nil
	}
	m.pushScreen(m.screen)
	m.screen = shared.InspectViewer
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.viewer.Vp.Follow = false
	m.viewer.Vp.SetXOffset(0)
	m.viewer.Vp.GotoTop()
	m.viewer.Vp.InitialLoad = true
	m.viewer.Vp.Data = nil
	m.viewer.ContainerID = resourceID
	m.viewer.Inspect.ResourceName = resourceName
	m.viewer.ResourceType = resourceType
	m.viewer.Vp.ContentType = viewer.ContentTypeInspect
	m.viewer.LoadingMsg = "Loading inspect..."
	m.viewer.EmptyMsg = "No inspect data available."
	m.viewer.Breadcrumb = ""
	m.viewer.ContainerName = resourceName
	m.viewer.Styles = viewer.Styles{
		Breadcrumb:   m.styles.Viewer.Breadcrumb,
		FollowOn:     m.styles.Viewer.FollowOn,
		FollowOff:    m.styles.Viewer.FollowOff,
		Muted:        m.styles.Browse.Muted,
		Divider:      m.styles.Browse.Divider,
		SubpageFrame: m.styles.Viewer.SubpageFrame,
		Key:          m.styles.Chrome.Key,
		KeyText:      m.styles.Chrome.KeyText,
	}
	return m, m.loadInspectCmd(resourceType, resourceID, resourceName)
}

func (m *model) loadInspectCmd(resourceType core.ResourceType, resourceID, resourceName string) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		data, err := svc.InspectResource(context.Background(), resourceType, resourceID)
		return inspectResultMsg{
			resourceType: resourceType,
			resourceID:   resourceID,
			resourceName: resourceName,
			data:         data,
			err:          err,
		}
	}
}

func (m *model) selectedInspectResource() (core.ResourceType, string, string, bool) {
	switch m.browse.ActiveTab {
	case tabContainers:
		c, ok := m.selectedContainer()
		if !ok {
			return 0, "", "", false
		}
		return core.ResourceContainer, c.FullID, c.Name, true
	case tabImages:
		img, ok := m.selectedImage()
		if !ok {
			return 0, "", "", false
		}
		return core.ResourceImage, img.ID, img.Tags, true
	case tabNetworks:
		net, ok := m.selectedNetwork()
		if !ok {
			return 0, "", "", false
		}
		return core.ResourceNetwork, net.ID, net.Name, true
	case tabVolumes:
		vol, ok := m.selectedVolume()
		if !ok {
			return 0, "", "", false
		}
		return core.ResourceVolume, vol.Name, vol.Name, true
	default:
		return 0, "", "", false
	}
}
