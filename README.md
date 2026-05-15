# EasyDocker 🐋

[![CI](https://github.com/joao-zanutto/easydocker/actions/workflows/pr-app.yml/badge.svg)](https://github.com/joao-zanutto/easydocker/actions/workflows/pr.yml)
[![Release](https://img.shields.io/github/v/release/joao-zanutto/easydocker?display_name=tag)](https://github.com/joao-zanutto/easydocker/releases)

EasyDocker is a TUI for Docker inspired by lazydocker and k9s while leveraging beautiful graphics from BubbleTea

### [See our Docs](https://joao-zanutto.github.io/easydocker/)

![easydocker usage](./docs/public/example.gif)

<div align="center">Troubleshoot your containers with style 😎</div>

## Features

Already implemented features:

- 🧭 Browse containers, images, networks, and volumes.
- 🪵 View live container logs that dynamic loads as you scroll up.
- 📃 Inspect resources to get low level information about them.
- 🔎 Filter resources, log lines and inspect output to look for specific strings.
- 🤿 Dive into containers by entering interacting shell sessions.
- 📊 Individual and aggregated container resource usage metrics.

Roadmap:

- 📡 Control containers in remote hosts.
- ⚙️ Settings customization.
- 🎨 Style customization.

💬 Suggest your own features! We would love to hear what would be most valuable for our user community.

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
	jpberno/easydocker:latest
```
