package docker

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"easydocker/internal/core"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
)

func mapContainerRow(item types.Container) core.ContainerRow {
	memoryUsage := core.MetricsNA
	cpuPercent := float64(0)
	if core.ContainerState(item.State) == core.StateRunning {
		cpuPercent = -1
	}

	return core.ContainerRow{
		ID:                 shortID(item.ID),
		FullID:             item.ID,
		Name:               primaryName(item.Names),
		ComposeProject:     strings.TrimSpace(item.Labels["com.docker.compose.project"]),
		ComposeService:     strings.TrimSpace(item.Labels["com.docker.compose.service"]),
		ComposeWorkingDir:  strings.TrimSpace(item.Labels["com.docker.compose.project.working_dir"]),
		ComposeConfigFiles: strings.TrimSpace(item.Labels["com.docker.compose.project.config_files"]),
		Image:              item.Image,
		State:              core.ContainerState(item.State),
		Status:             item.Status,
		Ports:              formatPorts(item.Ports),
		Command:            cleanCommand(item.Command),
		CreatedUnix:        item.Created,
		CPUPercent:         cpuPercent,
		Healthy:            strings.Contains(strings.ToLower(item.Status), "healthy"),
		MemoryUsage:        memoryUsage,
		MemoryLimit:        core.MetricsNA,
	}
}

func mapImageRow(item image.Summary) core.ImageRow {
	return core.ImageRow{
		ID:          shortID(item.ID),
		Tags:        formatTags(item.RepoTags),
		Size:        humanBytesUnknown(item.Size),
		Created:     core.HumanAge(time.Unix(item.Created, 0)),
		CreatedUnix: item.Created,
		Containers:  item.Containers,
	}
}

func mapNetworkRow(item network.Inspect) core.NetworkRow {
	return core.NetworkRow{
		ID:         shortID(item.ID),
		Name:       item.Name,
		Driver:     item.Driver,
		Scope:      item.Scope,
		Internal:   yesNo(item.Internal),
		Attachable: yesNo(item.Attachable),
		Endpoints:  len(item.Containers),
		Created:    core.HumanAge(item.Created),
		CreatedAt:  item.Created,
	}
}

func mapVolumeRow(item *volume.Volume) core.VolumeRow {
	refCount, size := int64(-1), int64(-1)
	if item.UsageData != nil {
		refCount = item.UsageData.RefCount
		size = item.UsageData.Size
	}
	createdAt, parseErr := time.Parse(time.RFC3339Nano, item.CreatedAt)
	if parseErr != nil {
		createdAt = time.Time{}
	}
	return core.VolumeRow{
		Name:       item.Name,
		Driver:     item.Driver,
		Scope:      item.Scope,
		Mountpoint: item.Mountpoint,
		Size:       humanBytesUnknown(size),
		RefCount:   refCount,
		Created:    humanTimestamp(item.CreatedAt),
		CreatedAt:  createdAt,
	}
}

func primaryName(names []string) string {
	if len(names) == 0 {
		return "-"
	}
	return strings.TrimPrefix(names[0], "/")
}

func formatPorts(ports []types.Port) string {
	if len(ports) == 0 {
		return "-"
	}

	sorted := make([]types.Port, len(ports))
	copy(sorted, ports)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].PrivatePort == sorted[j].PrivatePort {
			return sorted[i].PublicPort < sorted[j].PublicPort
		}
		return sorted[i].PrivatePort < sorted[j].PrivatePort
	})

	formatted := make([]string, 0, len(sorted))
	for _, port := range sorted {
		if port.PublicPort > 0 {
			formatted = append(formatted, fmt.Sprintf("%d->%d/%s", port.PublicPort, port.PrivatePort, port.Type))
			continue
		}
		formatted = append(formatted, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
	}

	return strings.Join(formatted, ", ")
}

func cleanCommand(command string) string {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return "-"
	}
	if len(trimmed) <= 64 {
		return trimmed
	}
	return trimmed[:61] + "..."
}

func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "<none>:<none>"
	}
	return strings.Join(tags, ", ")
}

func shortID(value string) string {
	trimmed := strings.TrimPrefix(value, "sha256:")
	if len(trimmed) <= 12 {
		return trimmed
	}
	return trimmed[:12]
}

func humanBytesUnknown(size int64) string {
	if size < 0 {
		return "-"
	}
	return core.HumanBytes(uint64(size))
}

func humanTimestamp(value string) string {
	if value == "" {
		return "-"
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return core.HumanAge(parsed)
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
