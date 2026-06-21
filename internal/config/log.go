package config

import (
	"os"
	"path/filepath"
)

type LoggingConfig struct {
	Enable bool   `yaml:"enable" mapstructure:"enable"     flag:"enable file logging"`
	Level  string `yaml:"level" mapstructure:"level"       flag:"log level (debug, info, warn, error)"`
	Path   string `yaml:"path" mapstructure:"path"         flag:"log file path"`
}

func defaultLogPath() string {
	if isWritableDir("/var/log") {
		return "/var/log/easydocker.log"
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "easydocker", "easydocker.log")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "easydocker", "easydocker.log")
	}
	return "easydocker.log"
}

func isWritableDir(dir string) bool {
	f, err := os.CreateTemp(dir, ".easydocker-write-check-*")
	if err != nil {
		return false
	}

	_ = f.Close()
	_ = os.Remove(f.Name())
	return true
}
