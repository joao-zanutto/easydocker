---
title: Install
---

# Install

::: warning
Under construction... 🚧 🏗️
:::

### Linux and macOS

```bash
curl -fsSL https://raw.githubusercontent.com/joao-zanutto/easydocker/main/install/install.sh | sh
```

### Windows

```powershell
irm https://raw.githubusercontent.com/joao-zanutto/easydocker/main/install/install.ps1 | iex
```

## Run

### Local

```bash
easydocker
```

### Docker

```bash
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  joaozanutto2/easydocker:latest
```
