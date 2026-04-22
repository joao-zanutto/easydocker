# EasyDocker 🐋

[![CI](https://github.com/joao-zanutto/easydocker/actions/workflows/pr.yml/badge.svg)](https://github.com/joao-zanutto/easydocker/actions/workflows/pr.yml)
[![Release](https://img.shields.io/github/v/release/joao-zanutto/easydocker?display_name=tag)](https://github.com/joao-zanutto/easydocker/releases)

EasyDocker is a TUI for Docker inspired by lazydocker and k9s while leveraging beautiful graphics from BubbleTea

### [See our Docs](https://joao-zanutto.github.io/easydocker/)

![easydocker usage](./docs/example.gif)

<div align="center">Troubleshoot your containers with style 😎</div>

## Features

This project is under development but already has the following functionalities implemented:

- Browse containers, images, networks, and volumes .
- View live container logs that loads as you scroll up.
- Individual and aggregated container resource usage metrics.
- Runs in really small terminal screens

## Install and Run

### Linux/macOS (sh):

```bash
curl -fsSL https://raw.githubusercontent.com/joao-zanutto/easydocker/main/install/install.sh | sh
```

### Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/joao-zanutto/easydocker/main/install/install.ps1 | iex
```

---

### Run

```bash
easydocker
```

### Docker

```bash
docker run --rm -it \
	-v /var/run/docker.sock:/var/run/docker.sock \
	joaozanutto2/easydocker:latest
```
