package docker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"easydocker/internal/core"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
)

func mapContainerRow(item types.Container) core.ContainerRow {
	memoryUsage := "-"
	cpuPercent := float64(0)
	if strings.EqualFold(item.State, "running") {
		cpuPercent = -1
	}

	return core.ContainerRow{
		ID:                     shortID(item.ID),
		FullID:                 item.ID,
		Name:                   primaryName(item.Names),
		ComposeProject:         strings.TrimSpace(item.Labels["com.docker.compose.project"]),
		ComposeService:         strings.TrimSpace(item.Labels["com.docker.compose.service"]),
		ComposeWorkingDir:      strings.TrimSpace(item.Labels["com.docker.compose.project.working_dir"]),
		ComposeConfigFiles:     strings.TrimSpace(item.Labels["com.docker.compose.project.config_files"]),
		ComposeOneOff:          strings.EqualFold(item.Labels["com.docker.compose.oneoff"], "true"),
		ComposeContainerNumber: parseContainerNumber(item.Labels["com.docker.compose.container-number"]),
		Image:                  item.Image,
		State:                  item.State,
		Status:                 item.Status,
		Ports:                  formatPorts(item.Ports),
		Command:                cleanCommand(item.Command),
		CreatedUnix:            item.Created,
		CPUPercent:             cpuPercent,
		Healthy:                strings.Contains(strings.ToLower(item.Status), "healthy"),
		MemoryUsage:            memoryUsage,
		MemoryLimit:            "-",
	}
}

func mapImageRow(item image.Summary) core.ImageRow {
	return core.ImageRow{
		ID:          shortID(item.ID),
		Tags:        formatTags(item.RepoTags),
		Size:        core.HumanBytes(item.Size),
		Created:     humanAge(time.Unix(item.Created, 0)),
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
		Created:    humanAge(item.Created),
		CreatedAt:  item.Created,
	}
}

func mapVolumeRow(item *volume.Volume) core.VolumeRow {
	refCount, size := int64(-1), int64(-1)
	if item.UsageData != nil {
		refCount = item.UsageData.RefCount
		size = item.UsageData.Size
	}
	return core.VolumeRow{
		Name:       item.Name,
		Driver:     item.Driver,
		Scope:      item.Scope,
		Mountpoint: item.Mountpoint,
		Size:       humanBytesUnknown(size),
		RefCount:   refCount,
		Created:    humanTimestamp(item.CreatedAt),
		CreatedAt:  item.CreatedAt,
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

func parseContainerNumber(value string) int {
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
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
	return core.HumanBytes(size)
}

func humanTimestamp(value string) string {
	if value == "" {
		return "-"
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return humanAge(parsed)
}

func humanAge(then time.Time) string {
	delta := time.Since(then)
	switch {
	case delta < time.Minute:
		return "just now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	case delta < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	case delta < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(delta.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(delta.Hours()/(24*365)))
	}
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
