---
title: Configuration
---

# Configuration

EasyDocker supports configuration through config file, environment variables or CLI options

## Precedence

Values are resolved in this order (last wins):

1. Default defaults
2. Config file values
3. Environment variable overrides
4. CLI flag overrides

## Config file

EasyDocker looks for a YAML config file in the following locations (in order):

1. `--config` flag or `EASYDOCKER_CONFIG` environment variable (explicit path)
2. `$XDG_CONFIG_HOME/easydocker/config.yaml`
3. `~/.config/easydocker/config.yaml`

```yaml
log:
  enable: true
  level: debug
  path: /tmp/easydocker.log
viewer:
  log:
    lines: 2000
```

If no config file is found, all options use their defaults.

## Options

| CLI flag             | Environment variable          | Config key         | Default                   | Description                                  |
| -------------------- | ----------------------------- | ------------------ | ------------------------- | -------------------------------------------- |
| `--config`           | `EASYDOCKER_CONFIG`           | —                  | auto                      | Path to YAML config file                     |
| `--log-enable`       | `EASYDOCKER_LOG_ENABLE`       | `log.enable`       | `false`                   | Enable file logging                          |
| `--log-level`        | `EASYDOCKER_LOG_LEVEL`        | `log.level`        | `warn`                    | Log level: `debug`, `info`, `warn`, `error`  |
| `--log-path`         | `EASYDOCKER_LOG_PATH`         | `log.path`         | `/var/log/easydocker.log` | Log file path                                |
| `--viewer-log-lines` | `EASYDOCKER_VIEWER_LOG_LINES` | `viewer.log.lines` | `2000`                    | Amount of log lines to fetch for a container |

## Examples

### Enable debug logging to default path

```bash
easydocker --log-enable --log-level=debug
```

### Use environment variables

```bash
EASYDOCKER_LOG_ENABLE=1 EASYDOCKER_LOG_LEVEL=debug easydocker
```

### Custom config file

```bash
easydocker --config /etc/easydocker/production.yaml
```

### Print version

```bash
easydocker version
```
