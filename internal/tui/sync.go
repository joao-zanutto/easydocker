package tui

import (
	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/ui/tables"
)

func (m model) syncBrowseData() model {
	filteredContainers := m.filteredContainers()
	activeProjects := activeComposeProjectNames(filteredContainers)
	pruneComposeExpanded(m.browse.ComposeExpanded, activeProjects)
	m.browse.Data = browse.BrowseData{
		ContainerListRows:       tables.BuildContainerListRows(filteredContainers, m.browse.ComposeExpanded),
		FilteredImages:          m.filteredImages(),
		FilteredNetworks:        m.filteredNetworks(),
		FilteredVolumes:         m.filteredVolumes(),
		MetricsLoadingIndicator: m.metricsLoadingIndicator(),
	}
	m.browse = m.browse.ClampCursors()
	m.browse.DetailProvider = m.browseDetailRenderer()
	return m
}

func activeComposeProjectNames(containers []core.ContainerRow) map[string]struct{} {
	names := make(map[string]struct{})
	for _, c := range containers {
		if c.ComposeProject != "" {
			names[c.ComposeProject] = struct{}{}
		}
	}
	return names
}

func pruneComposeExpanded(expanded map[string]bool, active map[string]struct{}) {
	if len(expanded) == 0 {
		return
	}
	for name := range expanded {
		if _, ok := active[name]; !ok {
			delete(expanded, name)
		}
	}
}
