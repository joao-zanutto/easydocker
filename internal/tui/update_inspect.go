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
	m.initViewer(shared.InspectViewer, viewer.ContentTypeInspect, resourceType,
		resourceID, resourceName, "Loading inspect...", "No inspect data available.")
	m.viewer.Vp.Follow = false
	m.viewer.Vp.GotoTop()
	m.viewer.Inspect.ResourceName = resourceName
	return m, m.loadInspectCmd(resourceType, resourceID)
}

func (m *model) loadInspectCmd(resourceType core.ResourceType, resourceID string) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		data, err := svc.InspectResource(context.Background(), resourceType, resourceID)
		return inspectResultMsg{data: data, err: err}
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
