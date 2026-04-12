package docker

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"time"

	"easydocker/internal/core"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

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
