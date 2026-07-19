package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := loadConfig()

	m := initialModel()
	m.config = cfg
	m.sourceName = cfg.SourceName()
	if cfg.ESDefaultESQLQuery != "" {
		m.esqlInput.SetValue(cfg.ESDefaultESQLQuery)
	}
	m.loading = true
	m.loadStart = time.Now()

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if m, ok := finalModel.(model); ok && m.exportMsg != "" {
		fmt.Println()
		fmt.Println(m.exportMsg)
	}
}
