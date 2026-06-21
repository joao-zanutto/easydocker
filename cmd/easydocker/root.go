package main

import (
	"errors"
	"fmt"
	"log"
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

		if err := InitLogging(cfg); err != nil {
			return fmt.Errorf("init logging: %w", err)
		}

		appliedConfig := cfg.DisplayLines(cfgPath)

		repo := docker.NewRepository()
		svc := core.NewService(repo)
		p := tea.NewProgram(tui.New(svc, appliedConfig, cfgPath, cfg.Viewer.Log.Lines))
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
	config.RegisterFlags(rootCmd)
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
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetViperDefaults(v)
	v.AutomaticEnv()
	config.BindFlags(v, cmd)

	return config.ReadConfig(v), cfgPath, nil
}


