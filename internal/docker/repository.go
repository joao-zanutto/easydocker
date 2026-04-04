package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"

	"easydocker/internal/core"

	"github.com/charmbracelet/x/ansi"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Repository struct {
	clientOnce sync.Once
	client     *client.Client
	clientErr  error
	now        func() time.Time
}

func NewRepository() *Repository {
	return &Repository{now: time.Now}
}

func (r *Repository) LoadContainerRows(ctx context.Context) ([]core.ContainerRow, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	rows := make([]core.ContainerRow, 0, len(containers))
	for _, item := range containers {
		rows = append(rows, mapContainerRow(item))
	}
	core.SortContainers(rows)

	return rows, nil
}

func (r *Repository) LoadSupportingResources(ctx context.Context) (core.Snapshot, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return core.Snapshot{}, err
	}

	images, networks, volumes, info, err := r.loadSupportingResourcesData(ctx, cli)
	if err != nil {
		return core.Snapshot{}, err
	}

	snapshot := core.Snapshot{
		Images:    make([]core.ImageRow, 0, len(images)),
		Networks:  make([]core.NetworkRow, 0, len(networks)),
		Volumes:   make([]core.VolumeRow, 0, len(volumes.Volumes)),
		Timestamp: r.now(),
	}

	for _, item := range images {
		snapshot.Images = append(snapshot.Images, mapImageRow(item))
	}
	core.SortImages(snapshot.Images)

	for _, item := range networks {
		snapshot.Networks = append(snapshot.Networks, mapNetworkRow(item))
	}
	core.SortNetworks(snapshot.Networks)

	for _, item := range volumes.Volumes {
		snapshot.Volumes = append(snapshot.Volumes, mapVolumeRow(item))
	}
	core.SortVolumes(snapshot.Volumes)

	snapshot.TotalCPU = 0
	snapshot.TotalMem = 0
	snapshot.TotalLimit = uint64(info.MemTotal)

	return snapshot, nil
}

func (r *Repository) LoadContainerMetrics(ctx context.Context, rows []core.ContainerRow) (map[string]core.ContainerMetrics, float64, uint64, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return nil, 0, 0, err
	}

	runningRows := make([]core.ContainerRow, 0, len(rows))
	for _, row := range rows {
		if strings.EqualFold(row.State, "running") {
			runningRows = append(runningRows, row)
		}
	}
	if len(runningRows) == 0 {
		return map[string]core.ContainerMetrics{}, 0, 0, nil
	}

	metricsByID := make(map[string]core.ContainerMetrics, len(runningRows))
	workerCount := min(len(runningRows), max(2, min(runtime.NumCPU(), 6)))
	jobs := make(chan core.ContainerRow)
	var mu sync.Mutex
	var totalCPU float64
	var totalMem uint64
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for row := range jobs {
			metrics, err := r.loadSingleContainerMetrics(ctx, cli, row.FullID)
			if err != nil {
				continue
			}

			mu.Lock()
			metricsByID[row.FullID] = metrics
			totalCPU += metrics.CPUPercent
			totalMem += metrics.MemoryUsageBytes
			mu.Unlock()
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}
	for _, row := range runningRows {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return metricsByID, totalCPU, totalMem, ctx.Err()
		case jobs <- row:
		}
	}
	close(jobs)
	wg.Wait()

	return metricsByID, totalCPU, totalMem, nil
}

func (r *Repository) LoadContainerLiveData(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (core.ContainerLiveData, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return core.ContainerLiveData{}, err
	}

	inspected, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return core.ContainerLiveData{}, fmt.Errorf("inspect container: %w", err)
	}

	logReader, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tailOption(tail),
	})
	if err != nil {
		return core.ContainerLiveData{}, fmt.Errorf("container logs: %w", err)
	}
	defer logReader.Close()

	rawLogBytes, err := io.ReadAll(logReader)
	if err != nil {
		return core.ContainerLiveData{}, fmt.Errorf("read container logs: %w", err)
	}

	var merged bytes.Buffer
	// Docker returns multiplexed streams for non-TTY containers and raw streams
	// for TTY containers. Try stdcopy first, then fall back to raw bytes.
	if _, err := stdcopy.StdCopy(&merged, &merged, bytes.NewReader(rawLogBytes)); err != nil {
		merged.Reset()
		_, _ = merged.Write(rawLogBytes)
	}

	metrics, err := r.loadSingleContainerMetrics(ctx, cli, containerID)
	if err != nil {
		metrics = core.ContainerMetrics{MemoryUsage: "-", MemoryLimit: "-"}
	}

	return core.ContainerLiveData{
		ContainerID:   containerID,
		Logs:          normalizeLogs(merged.String(), ""),
		CPUPercent:    metrics.CPUPercent,
		MemoryPercent: metrics.MemoryPercent,
		MemoryUsage:   metrics.MemoryUsage,
		MemoryLimit:   metrics.MemoryLimit,
		MemoryBytes:   metrics.MemoryUsageBytes,
		MemoryMax:     metrics.MemoryLimitBytes,
		CPUHistory:    appendHistory(previousCPU, metrics.CPUPercent),
		MemHistory:    appendHistory(previousMem, metrics.MemoryPercent),
		State:         inspected.State.Status,
		UpdatedAt:     r.now(),
	}, nil
}

func tailOption(tail int) string {
	if tail <= 0 {
		return "all"
	}
	return fmt.Sprintf("%d", tail)
}

func (r *Repository) loadSingleContainerMetrics(ctx context.Context, cli *client.Client, containerID string) (core.ContainerMetrics, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		statsCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		statsReader, err := cli.ContainerStats(statsCtx, containerID, false)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}

		var stats container.StatsResponse
		decodeErr := json.NewDecoder(statsReader.Body).Decode(&stats)
		closeErr := statsReader.Body.Close()
		cancel()
		if decodeErr != nil {
			lastErr = decodeErr
			continue
		}
		if closeErr != nil {
			lastErr = closeErr
			continue
		}

		cpuPercent := computeCPUPercent(stats)
		memoryPercent, memoryUsage, memoryLimit, memoryBytes, memoryMax := computeMemoryUsage(stats)

		return core.ContainerMetrics{
			CPUPercent:       cpuPercent,
			MemoryPercent:    memoryPercent,
			MemoryUsage:      memoryUsage,
			MemoryLimit:      memoryLimit,
			MemoryUsageBytes: memoryBytes,
			MemoryLimitBytes: memoryMax,
		}, nil
	}

	return core.ContainerMetrics{}, lastErr
}

func (r *Repository) dockerClient() (*client.Client, error) {
	r.clientOnce.Do(func() {
		r.client, r.clientErr = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	})
	return r.client, r.clientErr
}

func (r *Repository) loadSupportingResourcesData(ctx context.Context, cli *client.Client) ([]image.Summary, []network.Inspect, volume.ListResponse, system.Info, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type imagesResult struct {
		items []image.Summary
		err   error
	}
	type networksResult struct {
		items []network.Inspect
		err   error
	}
	type volumesResult struct {
		items volume.ListResponse
		err   error
	}
	type infoResult struct {
		item system.Info
		err  error
	}

	imagesCh := make(chan imagesResult, 1)
	networksCh := make(chan networksResult, 1)
	volumesCh := make(chan volumesResult, 1)
	infoCh := make(chan infoResult, 1)

	go func() {
		items, err := cli.ImageList(ctx, image.ListOptions{})
		imagesCh <- imagesResult{items: items, err: err}
	}()
	go func() {
		items, err := cli.NetworkList(ctx, network.ListOptions{})
		networksCh <- networksResult{items: items, err: err}
	}()
	go func() {
		items, err := cli.VolumeList(ctx, volume.ListOptions{})
		volumesCh <- volumesResult{items: items, err: err}
	}()
	go func() {
		item, err := cli.Info(ctx)
		infoCh <- infoResult{item: item, err: err}
	}()

	imagesRes := <-imagesCh
	if imagesRes.err != nil {
		cancel()
		return nil, nil, volume.ListResponse{}, system.Info{}, fmt.Errorf("list images: %w", imagesRes.err)
	}

	networksRes := <-networksCh
	if networksRes.err != nil {
		cancel()
		return nil, nil, volume.ListResponse{}, system.Info{}, fmt.Errorf("list networks: %w", networksRes.err)
	}

	volumesRes := <-volumesCh
	if volumesRes.err != nil {
		cancel()
		return nil, nil, volume.ListResponse{}, system.Info{}, fmt.Errorf("list volumes: %w", volumesRes.err)
	}

	infoRes := <-infoCh
	info := infoRes.item
	if infoRes.err != nil {
		info = system.Info{}
	}

	return imagesRes.items, networksRes.items, volumesRes.items, info, nil
}

func computeCPUPercent(stats container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if onlineCPUs == 0 {
		onlineCPUs = 1
	}
	return (cpuDelta / systemDelta) * onlineCPUs * 100
}

func computeMemoryUsage(stats container.StatsResponse) (float64, string, string, uint64, uint64) {
	used := effectiveMemoryUsage(stats)
	limit := stats.MemoryStats.Limit
	if limit == 0 {
		return 0, core.HumanBytes(int64(used)), "-", used, 0
	}
	percent := (float64(used) / float64(limit)) * 100
	return percent, core.HumanBytes(int64(used)), core.HumanBytes(int64(limit)), used, limit
}

func normalizeLogs(stdout, stderr string) []string {
	combined := strings.TrimRight(stdout, "\n")
	if strings.TrimSpace(stderr) != "" {
		if combined != "" {
			combined += "\n"
		}
		combined += strings.TrimRight(stderr, "\n")
	}
	combined = normalizeTerminalBoundaries(combined)
	// Normalize CRLF, but keep bare carriage returns for per-line compaction
	// so progress-style output does not explode into artificial new lines.
	combined = strings.ReplaceAll(combined, "\r\n", "\n")
	combined = strings.TrimRight(combined, "\n")
	if combined == "" {
		return []string{}
	}
	parts := strings.Split(combined, "\n")
	logs := make([]string, 0, len(parts))
	for _, line := range parts {
		normalized := collapseCarriageReturns(line)
		normalized = applyBackspaces(normalized)
		if isControlOnlyLogLine(normalized) {
			continue
		}
		logs = append(logs, normalized)
	}
	return logs
}

func collapseCarriageReturns(line string) string {
	segments := strings.Split(line, "\r")
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "" {
			return segments[i]
		}
	}
	return ""
}

func applyBackspaces(line string) string {
	if line == "" {
		return ""
	}
	stack := make([]rune, 0, len(line))
	for _, r := range line {
		if r == '\b' || r == 0x7f {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		stack = append(stack, r)
	}
	return string(stack)
}

func normalizeTerminalBoundaries(text string) string {
	// apt/dpkg progress uses ESC 7 (save cursor) / ESC 8 (restore cursor).
	// Treat them as virtual line boundaries so progress rows don't concatenate
	// with subsequent normal log lines.
	text = strings.ReplaceAll(text, "\x1b7", "")
	text = strings.ReplaceAll(text, "\x1b8", "\n")
	return text
}

func isControlOnlyLogLine(line string) bool {
	plain := ansi.Strip(line)
	visible := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, plain)
	return strings.TrimSpace(visible) == ""
}

func appendHistory(history []float64, value float64) []float64 {
	history = append(history, value)
	if len(history) > 32 {
		history = history[len(history)-32:]
	}
	return history
}

func effectiveMemoryUsage(stats container.StatsResponse) uint64 {
	used := stats.MemoryStats.Usage
	for _, key := range []string{"total_inactive_file", "inactive_file", "cache"} {
		if cached, ok := stats.MemoryStats.Stats[key]; ok && used >= cached {
			used -= cached
			break
		}
	}
	return used
}
