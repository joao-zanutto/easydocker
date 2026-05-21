package shared

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
	switch t {
	case TabContainers:
		return "Containers"
	case TabImages:
		return "Images"
	case TabNetworks:
		return "Networks"
	case TabVolumes:
		return "Volumes"
	default:
		return "Unknown"
	}
}
