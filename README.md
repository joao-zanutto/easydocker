# EasyDocker

EasyDocker is a terminal UI for Docker built with Charmbracelet and the Docker client SDK.

## Features

- Browse containers, images, networks, and volumes in one terminal application.
- Keep the UI live with automatic refresh every second.
- Open a focused container view with live logs plus CPU and memory monitors.
- Inspect the selected resource in a framed details pane.

## Requirements

- Go 1.25+
- Docker daemon running locally or reachable through the standard Docker environment variables
- Permission to access the Docker socket or remote API

## Run

```bash
go run ./cmd/easydocker
```

## Docker Image

Build the image:

```bash
docker build -t easydocker .
```

Run it against the local Docker daemon:

```bash
docker run --rm -it \
	-v /var/run/docker.sock:/var/run/docker.sock \
	easydocker
```

If you use a remote Docker host, pass the relevant Docker environment variables to the container as well.

## Controls

- `up` / `down`: move selection
- `tab`: switch between containers, images, networks, and volumes
- `a`: toggle between all and running containers in the containers tab
- `enter`: open live logs and CPU/memory monitors for the selected container
- `esc`: leave the live container view
- `q`: quit

## Notes

This starter focuses on read-only inspection with a live operational view for containers. It is a solid base for follow-up actions such as starting, stopping, or removing containers, and for adding logs, stats, or compose-aware views.
