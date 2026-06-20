package config

import (
	"os"
	"path/filepath"
)

type LoggingConfig struct {
	Enable bool   `yaml:"enable" mapstructure:"enable"`
	Level  string `yaml:"level" mapstructure:"level"`
	Path   string `yaml:"path" mapstructure:"path"`
}

type Config struct {
	Logging LoggingConfig `yaml:"log" mapstructure:"log"`
}

func Default() Config {
	return Config{
		Logging: LoggingConfig{
			Enable: false,
			Level:  "warn",
			Path:   defaultLogPath(),
		},
	}
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
	// ultimate fallback
	return "/var/log/easydocker.log"
}

func isWritableDir(dir string) bool {
	f, err := os.CreateTemp(dir, ".easydocker-write-check-*")
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(f.Name())
	return true
}
