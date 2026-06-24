package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	mobyterm "github.com/moby/term"
)

func shellExecOptions() container.ExecOptions {
	return container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"bash"},
	}
}

func shellExecOptionsFallback() container.ExecOptions {
	return container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"sh"},
	}
}

func (r *Repository) ExecShell(ctx context.Context, containerID string, stdin io.Reader, stdout, stderr io.Writer) error {
	cli, err := r.dockerClient()
	if err != nil {
		return err
	}

	execResp, err := cli.ContainerExecCreate(ctx, containerID, shellExecOptions())
	// If bash is not available, fallback to sh
	if err != nil {
		execResp, err = cli.ContainerExecCreate(ctx, containerID, shellExecOptionsFallback())
	}
	if err != nil {
		return fmt.Errorf("repository.create exec: %w", err)
	}

	resp, err := cli.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{Tty: true})
	if err != nil {
		return fmt.Errorf("repository.attach exec: %w", err)
	}
	defer resp.Close()

	if restore, ok := setupTerminalRaw(stdin); ok {
		defer restore()
	}
	forwardResizeSignals(ctx, cli, execResp.ID, stdout)

	// Pump I/O between the terminal and the hijacked exec connection.
	go func() {
		_, _ = io.Copy(resp.Conn, stdin)
		// Best-effort close of the write side of the exec connection.
		_ = resp.CloseWrite()
	}()

	outDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(stdout, resp.Reader)
		outDone <- err
	}()

	return <-outDone
}

// setupTerminalRaw puts stdin into raw mode and returns a restore function.
func setupTerminalRaw(stdin io.Reader) (func(), bool) {
	if f, ok := stdin.(*os.File); ok {
		fd := f.Fd()
		oldState, rawErr := mobyterm.MakeRaw(fd)
		if rawErr == nil {
			return func() { _ = mobyterm.RestoreTerminal(fd, oldState) }, true
		}
	}
	return nil, false
}

// forwardResizeSignals syncs the initial terminal size and forwards SIGWINCH.
// Resize errors are best-effort — the terminal will correct itself on the next resize event.
func forwardResizeSignals(ctx context.Context, cli *client.Client, execID string, stdout io.Writer) {
	if f, ok := stdout.(*os.File); ok {
		fd := f.Fd()
		if ws, sizeErr := mobyterm.GetWinsize(fd); sizeErr == nil {
			_ = cli.ContainerExecResize(ctx, execID, container.ResizeOptions{
				Height: uint(ws.Height),
				Width:  uint(ws.Width),
			})
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		defer func() { signal.Stop(sigCh); close(sigCh) }()
		go func() {
			for range sigCh {
				if ws, err := mobyterm.GetWinsize(fd); err == nil {
					_ = cli.ContainerExecResize(ctx, execID, container.ResizeOptions{
						Height: uint(ws.Height),
						Width:  uint(ws.Width),
					})
				}
			}
		}()
	}
}
