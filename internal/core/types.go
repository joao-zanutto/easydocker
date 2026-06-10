package core

import (
	"time"
)

type ContainerState string

const (
	StateRunning    ContainerState = "running"
	StateExited     ContainerState = "exited"
	StatePaused     ContainerState = "paused"
	StateCreated    ContainerState = "created"
	StateRemoving   ContainerState = "removing"
	StateDead       ContainerState = "dead"
	StateRestarting ContainerState = "restarting"
)

const (
	MetricsLoading = "loading"
	MetricsNA      = "-"
)

type ResourceType int

const (
	ResourceContainer ResourceType = iota
	ResourceImage
	ResourceNetwork
	ResourceVolume
)

func (r ResourceType) String() string {
	switch r {
	case ResourceContainer:
		return "Containers"
	case ResourceImage:
		return "Images"
	case ResourceNetwork:
		return "Networks"
	case ResourceVolume:
		return "Volumes"
	default:
		return "Unknown"
	}
}

type Snapshot struct {
	Containers      []ContainerRow
	ComposeProjects []ComposeProject
	Images          []ImageRow
	Networks        []NetworkRow
	Volumes         []VolumeRow
	TotalCPU        float64
	TotalMem        uint64
	TotalLimit      uint64
	Timestamp       time.Time
}

type ContainerRow struct {
	ID                 string
	FullID             string
	Name               string
	ComposeProject     string
	ComposeService     string
	ComposeWorkingDir  string
	ComposeConfigFiles string
	Image              string
	State              ContainerState
	Status             string
	Ports              string
	Command            string
	CreatedUnix        int64
	CPUPercent         float64
	MemoryPercent      float64
	MemoryUsage        string
	MemoryLimit        string
	MemoryUsageBytes   uint64
	MemoryLimitBytes   uint64
	Healthy            bool
	Networks           []string
}

type ComposeProject struct {
	Name             string
	Containers       []ContainerRow
	ContainerCount   int
	RunningCount     int
	HealthyCount     int
	CreatedUnix      int64
	Created          string
	Network          string
	WorkingDir       string
	ConfigFiles      string
	Services         []string
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsage      string
	MemoryLimit      string
	MemoryUsageBytes uint64
	MemoryLimitBytes uint64
}

type ImageRow struct {
	ID          string
	Tags        string
	Size        string
	Created     string
	CreatedUnix int64
	Containers  int64
}

type NetworkRow struct {
	ID         string
	Name       string
	Driver     string
	Scope      string
	Internal   string
	Attachable string
	Endpoints  int
	Created    string
	CreatedAt  time.Time
}

type VolumeRow struct {
	Name       string
	Driver     string
	Scope      string
	Mountpoint string
	Size       string
	RefCount   int64
	Created    string
	CreatedAt  time.Time
}

type ContainerMetrics struct {
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsage      string
	MemoryLimit      string
	MemoryUsageBytes uint64
	MemoryLimitBytes uint64
}

func ApplyMetricsToContainers(rows []ContainerRow, metricsByID map[string]ContainerMetrics) []ContainerRow {
	idx := make(map[string]int, len(rows))
	for i, row := range rows {
		idx[row.FullID] = i
	}
	for id, metrics := range metricsByID {
		i, ok := idx[id]
		if !ok {
			continue
		}
		rows[i].CPUPercent = metrics.CPUPercent
		rows[i].MemoryPercent = metrics.MemoryPercent
		rows[i].MemoryUsage = metrics.MemoryUsage
		rows[i].MemoryLimit = metrics.MemoryLimit
		rows[i].MemoryUsageBytes = metrics.MemoryUsageBytes
		rows[i].MemoryLimitBytes = metrics.MemoryLimitBytes
	}
	return rows
}

func PreserveRunningContainerMetrics(currentRows, previousRows []ContainerRow) []ContainerRow {
	if len(currentRows) == 0 || len(previousRows) == 0 {
		return currentRows
	}

	previousByID := make(map[string]ContainerRow, len(previousRows))
	for _, row := range previousRows {
		previousByID[row.FullID] = row
	}

	for index := range currentRows {
		row := &currentRows[index]
		if row.State != StateRunning {
			continue
		}
		if row.CPUPercent >= 0 && row.MemoryUsage != MetricsNA && row.MemoryUsage != MetricsLoading {
			continue
		}
		previous, ok := previousByID[row.FullID]
		if !ok || previous.MemoryUsage == MetricsNA || previous.MemoryUsage == MetricsLoading {
			continue
		}
		row.CPUPercent = previous.CPUPercent
		row.MemoryPercent = previous.MemoryPercent
		row.MemoryUsage = previous.MemoryUsage
		row.MemoryLimit = previous.MemoryLimit
		row.MemoryUsageBytes = previous.MemoryUsageBytes
		row.MemoryLimitBytes = previous.MemoryLimitBytes
	}

	return currentRows
}
