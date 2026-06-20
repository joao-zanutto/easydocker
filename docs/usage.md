---
title: Usage
---

# Usage

## Running EasyDocker

```
easydocker
```

Opens the interactive TUI for browsing Docker resources.

## Version

```
easydocker version
```

Prints the version, commit, and build date.

## CLI flags

| Flag           | Env var                 | Description                                 |
| -------------- | ----------------------- | ------------------------------------------- |
| `--config`     | `EASYDOCKER_CONFIG`     | Path to a YAML config file                  |
| `--log-enable` | `EASYDOCKER_LOG_ENABLE` | Enable file logging                         |
| `--log-level`  | `EASYDOCKER_LOG_LEVEL`  | Log level: `debug`, `info`, `warn`, `error` |
| `--log-path`   | `EASYDOCKER_LOG_PATH`   | Log file path                               |

For details on available options and config file format, see [Configuration](/configuration).

## Key bindings

### Browse mode

| Key             | Action                                              |
| --------------- | --------------------------------------------------- |
| `←` / `→`       | Switch tabs (Containers, Images, Networks, Volumes) |
| `↑` / `↓`       | Move selection                                      |
| `pgup` / `pgdn` | Jump by page                                        |
| `home` / `end`  | Go to top / bottom                                  |
| `a`             | Toggle running / all containers                     |
| `Enter`         | Open container logs                                 |
| `i`             | Inspect selected resource                           |
| `s`             | Open interactive shell in running container         |
| `/`             | Filter                                              |
| `Esc`           | Open menu                                           |

### Logs / Inspect mode

| Key             | Action                               |
| --------------- | ------------------------------------ |
| `←` / `→`       | Scroll horizontally                  |
| `↑` / `↓`       | Scroll lines                         |
| `pgup` / `pgdn` | Jump by page                         |
| `home` / `end`  | Go to top / bottom                   |
| `f`             | Toggle follow mode                   |
| `w`             | Toggle word wrap                     |
| `/`             | Filter log lines                     |
| `s`             | Open shell (running containers only) |
| `Esc`           | Go back to browse                    |

### Menu

| Key       | Action                  |
| --------- | ----------------------- |
| `↑` / `↓` | Navigate menu items     |
| `Enter`   | Select highlighted item |
| `Esc`     | Close menu              |

The Esc menu provides access to help, configuration viewer, and quit.
