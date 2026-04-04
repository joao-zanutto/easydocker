package core

import "strings"

func FilterContainersByScope(containers []ContainerRow, showAll bool) []ContainerRow {
	if showAll {
		return containers
	}
	filtered := make([]ContainerRow, 0, len(containers))
	for _, container := range containers {
		if strings.EqualFold(container.State, "running") {
			filtered = append(filtered, container)
		}
	}
	return filtered
}
