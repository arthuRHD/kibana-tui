package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func exportToFile(entries []LogEntry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	levelCounts := map[string]int{}
	for _, e := range entries {
		levelCounts[e.Level]++
	}
	var lvlParts []string
	for lvl, count := range levelCounts {
		lvlParts = append(lvlParts, fmt.Sprintf("%s=%d", lvl, count))
	}

	header := fmt.Sprintf("# Kibana TUI Export — %s\n# %d entries — %s\n%s\n",
		time.Now().Format(time.RFC3339),
		len(entries),
		strings.Join(lvlParts, ", "),
		strings.Repeat("#", 80))

	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	for _, e := range entries {
		line := fmt.Sprintf("[%s] [%s] [%s] %s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Level,
			e.Source,
			e.Message,
		)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("writing entry: %w", err)
		}
	}

	trailer := fmt.Sprintf("\n# Export completed at %s — %d entries\n",
		time.Now().Format(time.RFC3339), len(entries))
	f.WriteString(trailer)

	return nil
}
