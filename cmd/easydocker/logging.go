package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"easydocker/internal/config"
)

func InitLogging(cfg config.Config) error {
	if !cfg.Logging.Enable {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return nil
	}

	lvl := parseLogLevel(cfg.Logging.Level)

	dir := filepath.Dir(cfg.Logging.Path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create log directory %q: %w", dir, err)
	}

	f, err := os.OpenFile(cfg.Logging.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", cfg.Logging.Path, err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: lvl})))
	return nil
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
