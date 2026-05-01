package main

import (
	"log"

	"easydocker/internal/core"
	"easydocker/internal/docker"
	"easydocker/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	repo := docker.NewRepository()
	svc := core.NewService(repo)
	p := tea.NewProgram(tui.New(svc))
	if _, err := p.Run(); err != nil {
		log.Fatalf("run easydocker: %v", err)
	}
}
