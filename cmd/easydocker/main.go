package main

import (
	"log"

	"easydocker/internal/core"
	"easydocker/internal/docker"
	"easydocker/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	repo := docker.NewRepository()
	svc := core.NewService(repo)
	p := tea.NewProgram(tui.New(svc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("run easydocker: %v", err)
	}
}
