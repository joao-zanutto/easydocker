package shared

import "easydocker/internal/core"

// Tab identifies a resource tab in the browse screen.
type Tab int

const (
	TabContainers Tab = iota
	TabImages
	TabNetworks
	TabVolumes
	TabCount
)

func (t Tab) String() string {
	return TabToResourceType(t).String()
}

func TabToResourceType(tab Tab) core.ResourceType {
	switch tab {
	case TabContainers:
		return core.ResourceContainer
	case TabImages:
		return core.ResourceImage
	case TabNetworks:
		return core.ResourceNetwork
	case TabVolumes:
		return core.ResourceVolume
	default:
		return core.ResourceContainer
	}
}
