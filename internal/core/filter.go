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

// FilterContainersByQuery filters containers by name or image using case-insensitive contains matching.
// Empty query returns the full input slice unchanged.
func FilterContainersByQuery(containers []ContainerRow, query string) []ContainerRow {
	if strings.TrimSpace(query) == "" {
		return containers
	}
	query = strings.ToLower(query)
	filtered := make([]ContainerRow, 0, len(containers))
	for _, container := range containers {
		if strings.Contains(strings.ToLower(container.Name), query) ||
			strings.Contains(strings.ToLower(container.Image), query) {
			filtered = append(filtered, container)
		}
	}
	return filtered
}

// FilterImagesByQuery filters images by name (tags) using case-insensitive contains matching.
// Empty query returns the full input slice unchanged.
func FilterImagesByQuery(images []ImageRow, query string) []ImageRow {
	if strings.TrimSpace(query) == "" {
		return images
	}
	query = strings.ToLower(query)
	filtered := make([]ImageRow, 0, len(images))
	for _, image := range images {
		if strings.Contains(strings.ToLower(image.Tags), query) {
			filtered = append(filtered, image)
		}
	}
	return filtered
}

// FilterNetworksByQuery filters networks by name using case-insensitive contains matching.
// Empty query returns the full input slice unchanged.
func FilterNetworksByQuery(networks []NetworkRow, query string) []NetworkRow {
	if strings.TrimSpace(query) == "" {
		return networks
	}
	query = strings.ToLower(query)
	filtered := make([]NetworkRow, 0, len(networks))
	for _, network := range networks {
		if strings.Contains(strings.ToLower(network.Name), query) {
			filtered = append(filtered, network)
		}
	}
	return filtered
}

// FilterVolumesByQuery filters volumes by name using case-insensitive contains matching.
// Empty query returns the full input slice unchanged.
func FilterVolumesByQuery(volumes []VolumeRow, query string) []VolumeRow {
	if strings.TrimSpace(query) == "" {
		return volumes
	}
	query = strings.ToLower(query)
	filtered := make([]VolumeRow, 0, len(volumes))
	for _, volume := range volumes {
		if strings.Contains(strings.ToLower(volume.Name), query) {
			filtered = append(filtered, volume)
		}
	}
	return filtered
}

