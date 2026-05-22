package tui

import (
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
	m.previousScreen = m.screen
	m.screen = shared.Inspect
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.viewer.Follow = false
	m.viewer.Viewport.SetXOffset(0)
	m.viewer.Viewport.GotoTop()
	m.viewer.InitialLoad = true
	m.viewer.Data = nil
	m.viewer.ContainerID = resourceID
	m.viewer.ResourceName = resourceName
	m.viewer.ResourceType = resourceTypeFromTab(m.browse.ActiveTab)
	m.viewer.ContentType = viewer.ContentTypeInspect
	m.viewer.LoadingMsg = "Loading inspect..."
	m.viewer.EmptyMsg = "No inspect data available."
	m.viewer.Breadcrumb = ""
	m.viewer.ContainerName = resourceName
	m.viewer.Styles = viewer.ViewStyles{
		Breadcrumb:   m.styles.Breadcrumb,
		FollowOn:     m.styles.FollowOn,
		FollowOff:    m.styles.FollowOff,
		Muted:        m.styles.Muted,
		Divider:      m.styles.Divider,
		SubpageFrame: m.styles.SubpageFrame,
	}
	return m, m.loadInspectCmd(resourceType, resourceID, resourceName)
}

func (m *model) loadInspectCmd(resourceType shared.Tab, resourceID, resourceName string) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		data, err := svc.InspectResource(tabToResourceType(resourceType), resourceID)
		return inspectResultMsg{
			resourceType: int(resourceType),
			resourceID:   resourceID,
			resourceName: resourceName,
			data:         data,
			err:          err,
		}
	}
}

func tabToResourceType(tab shared.Tab) core.ResourceType {
	switch tab {
	case tabContainers:
		return core.ResourceContainer
	case tabImages:
		return core.ResourceImage
	case tabNetworks:
		return core.ResourceNetwork
	case tabVolumes:
		return core.ResourceVolume
	default:
		return core.ResourceContainer
	}
}

func (m model) selectedInspectResource() (shared.Tab, string, string, bool) {
	switch m.browse.ActiveTab {
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
