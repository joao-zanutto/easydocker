package core

import "strings"

func FilterContainersRunningOnly(containers []ContainerRow, showAll bool) []ContainerRow {
	if showAll {
		return containers
	}
	filtered := make([]ContainerRow, 0, len(containers))
	for _, container := range containers {
		if container.State == StateRunning {
			filtered = append(filtered, container)
		}
	}
	return filtered
}

func filterByQuery[T any](items []T, query string, names func(T) []string) []T {
	if strings.TrimSpace(query) == "" {
		return items
	}
	q := strings.ToLower(query)
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		for _, name := range names(item) {
			if strings.Contains(strings.ToLower(name), q) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}

func FilterContainersByQuery(containers []ContainerRow, query string) []ContainerRow {
	return filterByQuery(containers, query, func(c ContainerRow) []string {
		return []string{c.Name, c.Image, c.ComposeProject}
	})
}

func FilterImagesByQuery(images []ImageRow, query string) []ImageRow {
	return filterByQuery(images, query, func(i ImageRow) []string { return []string{i.Tags} })
}

func FilterNetworksByQuery(networks []NetworkRow, query string) []NetworkRow {
	return filterByQuery(networks, query, func(n NetworkRow) []string { return []string{n.Name} })
}

func FilterVolumesByQuery(volumes []VolumeRow, query string) []VolumeRow {
	return filterByQuery(volumes, query, func(v VolumeRow) []string { return []string{v.Name} })
}
