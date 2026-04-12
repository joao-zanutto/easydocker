# EasyDocker

![easydocker usage](./docs/tui.gif)

<div align="center">EasyDocker is a TUI for Docker inspired by legendary projects lazydocker and k9s, while leveraging beautiful graphics from BubbleTea</div>

## Features

This project is under development but already has the following functionalities implemented:

- Browse containers, images, networks, and volumes .
- View live container logs that loads as you scroll up.
- Individual and aggregated container resource usage metrics.
- Runs in really small terminal screens

## Installation

### Linux/macOS (sh):

```bash
curl -fsSL https://raw.githubusercontent.com/joao-zanutto/easydocker/main/install/install.sh | sh
```

### Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/joao-zanutto/easydocker/main/install/install.ps1 | iex
```

### Docker

```bash
docker run --rm -it \
	-v /var/run/docker.sock:/var/run/docker.sock \
	joaozanutto2/easydocker:latest
```
