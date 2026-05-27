package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"

	"easydocker/internal/core"

	"github.com/charmbracelet/x/ansi"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const maxHistoryPoints = 32

func (r *Repository) LoadContainerLiveData(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (core.ContainerLiveData, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return core.ContainerLiveData{}, err
	}

	inspected, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return core.ContainerLiveData{}, wrapDockerError("inspect container", err)
	}

	logs, err := fetchRawLogs(ctx, cli, containerID, tail)
	if err != nil {
		return core.ContainerLiveData{}, err
	}

	metrics, err := r.loadSingleContainerMetrics(ctx, cli, containerID)
	if err != nil {
		metrics = core.ContainerMetrics{MemoryUsage: "-", MemoryLimit: "-"}
	}

	return core.ContainerLiveData{
		ContainerID:   containerID,
		Logs:          logs,
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

func (r *Repository) LoadContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error) {
	return withClientResult(r, func(cli *client.Client) ([]string, error) {
		return fetchRawLogs(ctx, cli, containerID, tail)
	})
}

func fetchRawLogs(ctx context.Context, cli *client.Client, containerID string, tail int) ([]string, error) {
	logReader, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tailOption(tail),
	})
	if err != nil {
		return nil, wrapDockerError("container logs", err)
	}
	defer func() { _ = logReader.Close() }()

	rawLogBytes, err := io.ReadAll(logReader)
	if err != nil {
		return nil, wrapDockerError("read container logs", err)
	}

	var merged bytes.Buffer
	// Docker returns multiplexed streams for non-TTY containers and raw streams
	// for TTY containers. Try stdcopy first, then fall back to raw bytes.
	if _, err := stdcopy.StdCopy(&merged, &merged, bytes.NewReader(rawLogBytes)); err != nil {
		merged.Reset()
		_, _ = merged.Write(rawLogBytes)
	}

	return normalizeLogs(merged.String(), ""), nil
}

func tailOption(tail int) string {
	if tail <= 0 {
		return "all"
	}
	return fmt.Sprintf("%d", tail)
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
	if len(history) > maxHistoryPoints {
		history = history[len(history)-maxHistoryPoints:]
	}
	return history
}
