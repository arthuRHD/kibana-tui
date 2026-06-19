package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateBrowse state = iota
	stateSearch
	stateDetail
	stateExport
	stateFilter
)

type model struct {
	width  int
	height int
	state  state

	entries  []LogEntry
	filtered []LogEntry
	selected int
	topRow   int

	searchInput textinput.Model
	searchQuery string

	levelFilter string
	filterIdx   int

	detailIdx int

	exportInput textinput.Model
	exportMsg   string

	sourceName string
	loading    bool
	loadErr    error
	loadStart  time.Time

	filePath   string
	esAddr     string
	esIndex    string
	esUser     string
	esPass     string
	sampleSize int
}

type loadCompleteMsg struct {
	entries []LogEntry
	source  string
	err     error
}

type tickMsg struct {
	time time.Time
}

func initialModel() model {
	si := textinput.New()
	si.Placeholder = "type query and press Enter..."
	si.Prompt = "> "
	si.CharLimit = 200
	si.Width = 60

	ei := textinput.New()
	ei.Placeholder = "logs_export.log"
	ei.Prompt = ""
	ei.CharLimit = 200
	ei.Width = 40

	return model{
		state:       stateBrowse,
		searchInput: si,
		exportInput: ei,
		levelFilter: "ALL",
		filterIdx:   0,
		selected:    0,
		topRow:      0,
	}
}

func (m *model) updateFiltered() {
	m.filtered = filterEntries(m.entries, m.searchQuery, m.levelFilter)
	if len(m.filtered) == 0 {
		m.selected = 0
		m.topRow = 0
		return
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m *model) ensureVisible(rowsPerPage int) {
	if m.selected < m.topRow {
		m.topRow = m.selected
	}
	if m.selected >= m.topRow+rowsPerPage {
		m.topRow = m.selected - rowsPerPage + 1
	}
}

func (m model) headerHeight() int {
	return 1
}

func (m model) footerHeight() int {
	if m.state == stateSearch {
		return 2
	}
	return 1
}

func (m model) bodyHeight() int {
	return max(1, m.height-m.headerHeight()-m.footerHeight())
}

func (m model) Init() tea.Cmd {
	if !m.loading {
		return nil
	}

	return tea.Batch(
		func() tea.Msg {
			var entries []LogEntry
			var err error

			switch {
			case m.esAddr != "":
				cfg := ESConfig{
					Address:  m.esAddr,
					Index:    m.esIndex,
					Username: m.esUser,
					Password: m.esPass,
					Size:     m.sampleSize,
				}
				entries, err = queryES(cfg)
				if err == nil && len(entries) == 0 {
					err = fmt.Errorf("no entries found in index %s", m.esIndex)
				}

			case m.filePath != "":
				entries, err = loadFromFile(m.filePath)

			default:
				entries = generateSampleData(m.sampleSize)
			}

			return loadCompleteMsg{entries: entries, source: m.sourceName, err: err}
		},
		tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg{t}
		}),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = max(20, msg.Width-50)
		m.exportInput.Width = max(20, msg.Width-30)
		return m, nil

	case loadCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err
		} else {
			m.entries = msg.entries
			m.sourceName = msg.source
		}
		m.updateFiltered()
		return m, nil

	case tickMsg:
		if m.loading {
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg{t}
			})
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		if m.state != stateBrowse && msg.String() == "q" {
			return m, nil
		}

		switch m.state {
		case stateBrowse:
			return m.updateBrowse(msg)
		case stateSearch:
			return m.updateSearch(msg)
		case stateDetail:
			return m.updateDetail(msg)
		case stateExport:
			return m.updateExport(msg)
		case stateFilter:
			return m.updateFilter(msg)
		}

	default:
		switch m.state {
		case stateSearch:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.searchQuery = m.searchInput.Value()
			m.updateFiltered()
			return m, cmd
		case stateExport:
			var cmd tea.Cmd
			m.exportInput, cmd = m.exportInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.filtered)-1 {
			m.selected++
		}
	case "left", "h":
		m.selected = 0
	case "right", "l":
		m.selected = max(0, len(m.filtered)-1)
	case "g":
		m.selected = 0
		m.topRow = 0
	case "G":
		m.selected = max(0, len(m.filtered)-1)
	case "pgup":
		pageSize := max(1, m.bodyHeight()-2)
		m.selected = max(0, m.selected-pageSize)
	case "pgdown":
		pageSize := max(1, m.bodyHeight()-2)
		m.selected = min(len(m.filtered)-1, m.selected+pageSize)
	case "/":
		m.state = stateSearch
		m.searchInput.Focus()
		m.searchInput.SetValue("")
		return m, textinput.Blink
	case "f":
		m.state = stateFilter
		m.filterIdx = 0
		levels := []string{"ALL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}
		for i, l := range levels {
			if l == m.levelFilter {
				m.filterIdx = i
				break
			}
		}
	case "e":
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.state = stateExport
		m.exportInput.SetValue(fmt.Sprintf("kibana-tui-export-%d.log", time.Now().Unix()))
		m.exportInput.Focus()
		m.exportMsg = ""
		return m, textinput.Blink
	case "enter":
		if len(m.filtered) > 0 {
			m.detailIdx = m.selected
			m.state = stateDetail
		}
	case "q":
		return m, tea.Quit
	}

	m.ensureVisible(max(1, m.bodyHeight()-2))
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchQuery = m.searchInput.Value()
		m.searchInput.Blur()
		m.state = stateBrowse
		m.selected = 0
		m.topRow = 0
		m.updateFiltered()
		return m, nil
	case "esc":
		m.searchInput.SetValue("")
		m.searchQuery = ""
		m.searchInput.Blur()
		m.state = stateBrowse
		m.updateFiltered()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchQuery = m.searchInput.Value()
		m.selected = 0
		m.topRow = 0
		m.updateFiltered()
		return m, cmd
	}
}

func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", " ", "q":
		m.state = stateBrowse
	case "up", "k":
		if m.detailIdx > 0 {
			m.detailIdx--
		}
	case "down", "j":
		if m.detailIdx < len(m.filtered)-1 {
			m.detailIdx++
		}
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateExport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		path := m.exportInput.Value()
		if path == "" {
			path = "logs_export.log"
		}
		m.exportInput.Blur()

		entries := make([]LogEntry, len(m.filtered))
		copy(entries, m.filtered)

		if err := exportToFile(entries, path); err != nil {
			m.exportMsg = fmt.Sprintf("Export failed: %v", err)
		} else {
			m.exportMsg = fmt.Sprintf("Exported %d entries to %s", len(entries), path)
		}
		m.state = stateBrowse
		return m, nil
	case "esc":
		m.exportInput.SetValue("")
		m.exportInput.Blur()
		m.state = stateBrowse
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.exportInput, cmd = m.exportInput.Update(msg)
		return m, cmd
	}
}

func (m model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	levels := []string{"ALL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

	switch msg.String() {
	case "enter":
		m.levelFilter = levels[m.filterIdx]
		m.state = stateBrowse
		m.selected = 0
		m.topRow = 0
		m.updateFiltered()
		return m, nil
	case "esc":
		m.state = stateBrowse
		return m, nil
	case "up", "k":
		m.filterIdx = (m.filterIdx - 1 + len(levels)) % len(levels)
	case "down", "j":
		m.filterIdx = (m.filterIdx + 1) % len(levels)
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	if m.loading {
		return m.loadingView()
	}

	if m.loadErr != nil {
		return m.errorView()
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	bodyH := m.bodyHeight()

	var body string
	switch m.state {
	case stateDetail:
		body = lipgloss.Place(m.width, bodyH,
			lipgloss.Center, lipgloss.Center,
			m.renderDetailView(),
			lipgloss.WithWhitespaceBackground(colorBase),
		)
	case stateExport:
		body = lipgloss.Place(m.width, bodyH,
			lipgloss.Center, lipgloss.Center,
			m.renderExportView(),
			lipgloss.WithWhitespaceBackground(colorBase),
		)
	case stateFilter:
		body = lipgloss.Place(m.width, bodyH,
			lipgloss.Center, lipgloss.Center,
			m.renderFilterView(),
			lipgloss.WithWhitespaceBackground(colorBase),
		)
	default:
		body = m.renderTable()
		if body == "" {
			body = strings.Repeat("\n", bodyH)
		}
	}

	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) < bodyH {
		padding := make([]string, bodyH-len(bodyLines))
		bodyLines = append(bodyLines, padding...)
	}
	body = strings.Join(bodyLines[:bodyH], "\n")

	content := lipgloss.JoinVertical(lipgloss.Top, header, body, footer)

	if m.exportMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(colorGreen).
			Background(colorMantle).
			Padding(0, 1).
			Width(m.width)
		msgLine := msgStyle.Render(m.exportMsg)
		lines := strings.Split(content, "\n")
		if len(lines) > 1 {
			lines[len(lines)-1] = msgLine
			content = strings.Join(lines, "\n")
		}
	}

	return content
}

func (m model) loadingView() string {
	elapsed := time.Since(m.loadStart)
	spinner := []string{"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f"}
	frame := spinner[int(elapsed.Milliseconds()/100)%len(spinner)]

	msg := fmt.Sprintf(" %s Loading %s... (%s) ", frame, m.sourceName, elapsed.Round(100*time.Millisecond))

	box := overlayBoxStyle.
		Width(50).
		Render(lipgloss.Place(46, 3, lipgloss.Center, lipgloss.Center, msg))

	return lipgloss.Place(80, 24,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceBackground(colorBase),
	)
}

func (m model) errorView() string {
	msg := fmt.Sprintf("Error: %v", m.loadErr)
	box := overlayBoxStyle.
		BorderForeground(colorRed).
		Width(60).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render(" Failed to load data "),
				"",
				msg,
				"",
				lipgloss.NewStyle().Foreground(colorDim).Render(" Press q to quit "),
			),
		)

	return lipgloss.Place(80, 24,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceBackground(colorBase),
	)
}
