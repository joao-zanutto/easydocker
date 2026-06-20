package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"easydocker/internal/config"
	"easydocker/internal/core"
	"easydocker/internal/docker"
	"easydocker/internal/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "easydocker",
	Short: "TUI for Docker container management and troubleshooting",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cfgPath, err := loadConfig(cmd)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if err := initLogging(cfg); err != nil {
			return fmt.Errorf("init logging: %w", err)
		}

		var appliedConfig []string
		appliedConfig = append(appliedConfig, fmt.Sprintf("log.enable:  %t", cfg.Logging.Enable))
		appliedConfig = append(appliedConfig, fmt.Sprintf("log.level:   %s", cfg.Logging.Level))
		appliedConfig = append(appliedConfig, fmt.Sprintf("log.path:    %s", cfg.Logging.Path))
		cfgFileLine := "config:      (none)"
		if cfgPath != "" {
			absPath, err := filepath.Abs(cfgPath)
			if err == nil {
				cfgPath = absPath
			}
			cfgFileLine = "config:      " + cfgPath
		}
		appliedConfig = append(appliedConfig, "", cfgFileLine)

		repo := docker.NewRepository()
		svc := core.NewService(repo)
		p := tea.NewProgram(tui.New(svc, appliedConfig, cfgPath))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("run easydocker: %v", err)
		}
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("easydocker %s (%s, %s)\n", version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().String("config", "", "path to config file")
	rootCmd.Flags().Bool("log-enable", config.Default().Logging.Enable, "enable file logging")
	rootCmd.Flags().String("log-level", config.Default().Logging.Level, "log level (debug, info, warn, error)")
	rootCmd.Flags().String("log-path", config.Default().Logging.Path, "log file path")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("easydocker: %v", err)
	}
}

func loadConfig(cmd *cobra.Command) (config.Config, string, error) {
	v := viper.New()

	explicitPath, _ := cmd.Flags().GetString("config")
	if explicitPath == "" {
		explicitPath = os.Getenv("EASYDOCKER_CONFIG")
	}

	hasExplicitPath := explicitPath != ""
	if hasExplicitPath {
		v.SetConfigFile(explicitPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			v.AddConfigPath(filepath.Join(xdg, "easydocker"))
		}
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "easydocker"))
		}
	}

	cfgPath := ""
	if err := v.ReadInConfig(); err != nil {
		var cfgErr viper.ConfigFileNotFoundError
		if hasExplicitPath || !errors.As(err, &cfgErr) {
			return config.Config{}, "", fmt.Errorf("read config: %w", err)
		}
	} else {
		cfgPath = v.ConfigFileUsed()
	}

	v.SetEnvPrefix("EASYDOCKER")
	v.SetEnvKeyReplacer(strings.NewReplacer("log.", "LOG_"))
	v.AutomaticEnv()

	_ = v.BindPFlag("log.enable", cmd.Flags().Lookup("log-enable"))
	_ = v.BindPFlag("log.level", cmd.Flags().Lookup("log-level"))
	_ = v.BindPFlag("log.path", cmd.Flags().Lookup("log-path"))

	return config.Config{
		Logging: config.LoggingConfig{
			Enable: v.GetBool("log.enable"),
			Level:  v.GetString("log.level"),
			Path:   v.GetString("log.path"),
		},
	}, cfgPath, nil
}

func initLogging(cfg config.Config) error {
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
