package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types/container"
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
		return fmt.Errorf("create exec: %w", err)
	}

	resp, err := cli.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{Tty: true})
	if err != nil {
		return fmt.Errorf("attach exec: %w", err)
	}
	defer resp.Close()

	// Put stdin into raw mode so the remote shell gets raw keystrokes.
	if f, ok := stdin.(*os.File); ok {
		fd := f.Fd()
		oldState, rawErr := mobyterm.MakeRaw(fd)
		if rawErr == nil {
			defer mobyterm.RestoreTerminal(fd, oldState) //nolint:errcheck
		}
	}

	// Sync initial terminal size to the exec session.
	if f, ok := stdout.(*os.File); ok {
		fd := f.Fd()
		if ws, sizeErr := mobyterm.GetWinsize(fd); sizeErr == nil {
			_ = cli.ContainerExecResize(ctx, execResp.ID, container.ResizeOptions{
				Height: uint(ws.Height),
				Width:  uint(ws.Width),
			})
		}

		// Forward subsequent SIGWINCH signals so the shell reacts to resizes.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		defer signal.Stop(sigCh)
		go func() {
			for range sigCh {
				if ws, err := mobyterm.GetWinsize(fd); err == nil {
					_ = cli.ContainerExecResize(ctx, execResp.ID, container.ResizeOptions{
						Height: uint(ws.Height),
						Width:  uint(ws.Width),
					})
				}
			}
		}()
	}

	// Pump I/O between the terminal and the hijacked exec connection.
	go func() {
		_, _ = io.Copy(resp.Conn, stdin)
		_ = resp.CloseWrite()
	}()

	outDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(stdout, resp.Reader)
		outDone <- err
	}()

	return <-outDone
}
