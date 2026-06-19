package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var (
		filePath   string
		esAddr     string
		esIndex    string
		esUser     string
		esPass     string
		sampleSize int
	)

	flag.StringVar(&filePath, "file", "", "Path to JSON log file")
	flag.StringVar(&esAddr, "es", "", "Elasticsearch address (e.g., http://localhost:9200)")
	flag.StringVar(&esIndex, "es-index", "logs-*", "Elasticsearch index pattern")
	flag.StringVar(&esUser, "es-user", "", "Elasticsearch username")
	flag.StringVar(&esPass, "es-pass", "", "Elasticsearch password")
	flag.IntVar(&sampleSize, "n", 500, "Number of entries to load (sample data or ES max size)")
	flag.Parse()

	m := initialModel()
	m.loading = true
	m.loadStart = time.Now()

	switch {
	case esAddr != "":
		m.sourceName = fmt.Sprintf("ES:%s/%s", esAddr, esIndex)
		m.esAddr = esAddr
		m.esIndex = esIndex
		m.esUser = esUser
		m.esPass = esPass
		m.sampleSize = sampleSize
	case filePath != "":
		m.sourceName = filePath
		m.filePath = filePath
	default:
		m.sourceName = fmt.Sprintf("sample (%d)", sampleSize)
		m.sampleSize = sampleSize
	}

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
