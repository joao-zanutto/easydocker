package main

import (
	"log"
	"log/slog"
	"os"

	"easydocker/internal/core"
	"easydocker/internal/docker"
	"easydocker/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	logFile, err := os.OpenFile("easydocker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("error closing log file: %v", err)
		}
	}()

	slog.SetDefault(slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelWarn})))

	repo := docker.NewRepository()
	svc := core.NewService(repo)
	p := tea.NewProgram(tui.New(svc))
	if _, err := p.Run(); err != nil {
		log.Fatalf("run easydocker: %v", err)
	}
}
