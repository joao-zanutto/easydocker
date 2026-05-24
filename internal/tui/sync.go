package tui

import (
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/ui/tables"
)

func (m model) syncBrowseData() model {
	m.browse.Data = browse.BrowseData{
		ContainerListRows:       tables.BuildContainerListRows(m.filteredContainers(), m.browse.ComposeExpanded),
		FilteredImages:          m.filteredImages(),
		FilteredNetworks:        m.filteredNetworks(),
		FilteredVolumes:         m.filteredVolumes(),
		MetricsLoadingIndicator: m.containerMetricsLoadingIndicator(),
	}
	m.browse.DetailProvider = m.browseDetailRenderer()
	m.browse.ListRenderer = func(width, h int) string {
		return m.renderResourceList(width, h)
	}
	return m
}
