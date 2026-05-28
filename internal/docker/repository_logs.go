package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

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

	return normalizeLogs(merged.String()), nil
}

func tailOption(tail int) string {
	if tail <= 0 {
		return "all"
	}
	return fmt.Sprintf("%d", tail)
}

func normalizeLogs(stdout string) []string {
	combined := strings.TrimRight(stdout, "\n")
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

