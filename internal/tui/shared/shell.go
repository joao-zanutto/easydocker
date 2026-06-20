package shared

import (
	"context"
	"io"

	"easydocker/internal/core"

	tea "charm.land/bubbletea/v2"
)

type ShellCommand struct {
	service     core.ServiceInterface
	containerID string
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
}

func (e *ShellCommand) SetStdin(r io.Reader)  { e.stdin = r }
func (e *ShellCommand) SetStdout(w io.Writer) { e.stdout = w }
func (e *ShellCommand) SetStderr(w io.Writer) { e.stderr = w }

func (e *ShellCommand) Run() error {
	// Enter/exit alternate screen buffer. Errors are intentionally discarded
	// because terminal escape sequences are best-effort.
	_, _ = io.WriteString(e.stdout, "\033[?1049h\033[H")
	defer func() {
		_, _ = io.WriteString(e.stdout, "\033[?1049l")
	}()
	return e.service.ExecShell(context.Background(), e.containerID, e.stdin, e.stdout, e.stderr)
}

func ShellCmd(service core.ServiceInterface, containerID string, doneMsg tea.Msg) tea.Cmd {
	return tea.Exec(
		&ShellCommand{service: service, containerID: containerID},
		func(err error) tea.Msg {
			return doneMsg
		},
	)
}
