package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderHeader() string {
	var left string
	if m.searchQuery != "" {
		left = lipgloss.NewStyle().Foreground(colorAccent).Render(" / ") +
			lipgloss.NewStyle().Foreground(colorText).Render(m.searchQuery)
	} else if m.state == stateSearch {
		left = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(" /... ")
	} else {
		left = lipgloss.NewStyle().Foreground(colorDim).Render(" press / to search ")
	}

	levelBadge := lipgloss.NewStyle().
		Background(colorSurface).
		Foreground(levelColor(m.levelFilter)).
		Padding(0, 1).
		Render(m.levelFilter)

	countStr := fmt.Sprintf(" %d entries ", len(m.filtered))
	if m.searchQuery != "" || m.levelFilter != "ALL" {
		countStr = fmt.Sprintf(" %d / %d ", len(m.filtered), len(m.entries))
	}

	right := levelBadge + " " + countStr

	availW := max(m.width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	spacer := lipgloss.NewStyle().Width(availW).Render("")
	line := lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
	return headerStyle.Width(m.width).Render(line)
}

func (m model) renderTable() string {
	if len(m.filtered) == 0 {
		if m.searchQuery != "" || m.levelFilter != "ALL" {
			return m.emptyView("No matching entries found. Press Esc to clear filters.")
		}
		return m.emptyView("No log entries available.")
	}

	availW := m.width
	bodyH := m.bodyHeight()
	rowsAvail := max(1, bodyH-1)

	const (
		colNumW = 5
		colTSW  = 20
		colLvlW = 6
		colSrcW = 16
	)

	sepChars := " \u2502 "
	sepW := 3
	totalSep := sepW * 4
	msgW := max(availW-colNumW-colTSW-colLvlW-colSrcW-totalSep-1, 10)

	hdrNum := lipgloss.NewStyle().Width(colNumW - 1).Render("#")
	hdrTS := lipgloss.NewStyle().Width(colTSW - 1).Render("TIMESTAMP")
	hdrLvl := lipgloss.NewStyle().Width(colLvlW - 1).Render("LEVEL")
	hdrMsg := lipgloss.NewStyle().Width(msgW).Render("MESSAGE")
	hdrSrc := lipgloss.NewStyle().Width(colSrcW - 1).Render("SOURCE")

	hdrParts := []string{
		" " + hdrNum, sepChars, hdrTS, sepChars, hdrLvl, sepChars, hdrMsg, sepChars, hdrSrc,
	}
	header := tableHeaderStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, hdrParts...),
	)

	if rowsAvail <= 0 {
		return header
	}

	m.ensureVisible(rowsAvail)

	var rows []string
	endIdx := min(len(m.filtered), m.topRow+rowsAvail)

	for i := m.topRow; i < endIdx; i++ {
		e := m.filtered[i]

		rowStyle := altRowStyle
		if i == m.selected {
			rowStyle = selectedRowStyle
		}

		rowNum := fmt.Sprintf("%d", i+1)
		ts := e.Timestamp.Format("01-02 15:04:05")
		lvl := e.Level
		msg := e.Message
		src := e.Source

		if len(msg) > msgW {
			msg = msg[:msgW-3] + "..."
		}
		if len(src) > colSrcW-3 {
			src = src[:colSrcW-5] + "..."
		}

		styleNum := lipgloss.NewStyle().Width(colNumW - 1)
		styleTS := lipgloss.NewStyle().Width(colTSW - 1).Foreground(colorDim)
		styleMsg := lipgloss.NewStyle().Width(msgW)
		styleSrc := lipgloss.NewStyle().Width(colSrcW - 1).Foreground(colorDim)

		numStr := styleNum.Render(rowNum)
		tsStr := styleTS.Render(ts)
		lvlStr := lipgloss.NewStyle().Width(colLvlW - 1).Render(levelStyles[lvl].Render(lvl))
		msgStr := styleMsg.Render(msg)
		srcStr := styleSrc.Render(src)

		parts := []string{
			" " + numStr, sepChars, tsStr, sepChars, lvlStr, sepChars, msgStr, sepChars, srcStr,
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
		rows = append(rows, rowStyle.Render(row))
	}

	return header + "\n" + strings.Join(rows, "\n")
}

func (m model) renderDetailView() string {
	if m.detailIdx < 0 || m.detailIdx >= len(m.filtered) {
		return ""
	}

	e := m.filtered[m.detailIdx]

	title := overlayTitleStyle.Render(" Log Detail ")
	navInfo := lipgloss.NewStyle().Foreground(colorDim).Render(
		fmt.Sprintf(" %d / %d ", m.detailIdx+1, len(m.filtered)))

	ts := lipgloss.NewStyle().Bold(true).Render("Timestamp: ") +
		lipgloss.NewStyle().Render(e.Timestamp.Format("2006-01-02 15:04:05.000 MST"))

	lvl := lipgloss.NewStyle().Bold(true).Render("Level:     ") +
		levelStyles[e.Level].Render(e.Level)

	src := lipgloss.NewStyle().Bold(true).Render("Source:    ") +
		lipgloss.NewStyle().Render(e.Source)

	msg := lipgloss.NewStyle().Bold(true).Render("Message:   ") +
		lipgloss.NewStyle().Render(e.Message)

	var fieldLines []string
	if len(e.Fields) > 0 {
		fieldLines = append(fieldLines, lipgloss.NewStyle().Bold(true).Render("Fields:"))
		for k, v := range e.Fields {
			fieldLines = append(fieldLines, fmt.Sprintf("  %s: %v",
				lipgloss.NewStyle().Foreground(colorAccent).Render(k), v))
		}
	}

	var contentParts []string
	contentParts = append(contentParts, title, navInfo, "", ts, lvl, src, "",
		lipgloss.NewStyle().Width(50).Render(msg))
	if len(fieldLines) > 0 {
		contentParts = append(contentParts, "", strings.Join(fieldLines, "\n"))
	}
	contentParts = append(contentParts, "",
		lipgloss.NewStyle().Foreground(colorDim).Render(" \u2191\u2193 navigate  Esc close "))

	content := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

	maxW := max(min(m.width-8, 60), 40)

	box := overlayBoxStyle.Width(maxW).Render(content)
	return box
}

func (m model) renderExportView() string {
	availW := max(min(m.width-8, 50), 30)

	title := overlayTitleStyle.Render(" Export Logs ")
	countInfo := fmt.Sprintf(" %d entries will be exported", len(m.filtered))

	filenameInput := m.exportInput.View()
	if filenameInput == "" {
		filenameInput = " "
	}

	helpTxt := lipgloss.NewStyle().Foreground(colorDim).Render(" Enter to export  Esc to cancel ")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		countInfo, "",
		lipgloss.NewStyle().Foreground(colorAccent).Render(" Filename:")+" "+filenameInput,
		"",
		helpTxt,
	)

	box := overlayBoxStyle.Width(availW).Render(content)
	return box
}

func (m model) renderFilterView() string {
	levels := []string{"ALL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

	title := overlayTitleStyle.Render(" Filter by Level ")

	var items []string
	for i, lvl := range levels {
		cursor := " "
		if i == m.filterIdx {
			cursor = "\u25b6"
		}
		checked := " "
		if lvl == m.levelFilter {
			checked = "\u2713"
		}
		lvlDisplay := lvl
		if s, ok := levelStyles[lvl]; ok {
			lvlDisplay = s.Render(lvl)
		}
		itemStyle := lipgloss.NewStyle()
		if i == m.filterIdx {
			itemStyle = itemStyle.Foreground(colorAccent).Bold(true)
		}
		items = append(items, itemStyle.Render(fmt.Sprintf(" %s [%s] %s", cursor, checked, lvlDisplay)))
	}

	helpTxt := lipgloss.NewStyle().Foreground(colorDim).Render(" \u2191\u2193 select  Enter apply  Esc cancel ")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		strings.Join(items, "\n"),
		"",
		helpTxt,
	)

	availW := max(min(m.width-8, 40), 30)

	box := overlayBoxStyle.Width(availW).Render(content)
	return box
}

func (m model) renderFooter() string {
	if m.state == stateSearch {
		return m.renderSearchBar()
	}
	if m.state == stateEsql {
		return m.renderEsqlBar()
	}

	keys := []struct {
		key  string
		desc string
	}{
		{"\u2191\u2193", "nav"},
		{"/", "search"},
		{"Enter", "detail"},
		{"e", "export"},
		{"f", "filter"},
		{"Q", "esql"},
		{"g/G", "top/bot"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		kStr := keyStyle.Render(k.key)
		dStr := helpStyle.Render(k.desc)
		parts = append(parts, kStr+" "+dStr)
	}
	shortcuts := strings.Join(parts, "  ")

	info := lipgloss.NewStyle().Foreground(colorDim).Render(m.sourceName)

	footerW := m.width
	gap := max(footerW-lipgloss.Width(shortcuts)-lipgloss.Width(info), 0)
	gapStr := strings.Repeat(" ", gap)

	line := shortcuts + gapStr + info
	return footerStyle.Width(m.width).Render(line)
}

func (m model) renderSearchBar() string {
	prompt := searchPromptStyle.Render(" /> ")
	input := m.searchInput.View()
	if input == "" {
		input = " "
	}

	countStr := lipgloss.NewStyle().Foreground(colorDim).Render(
		fmt.Sprintf(" %d / %d ", len(m.filtered), len(m.entries)))

	content := lipgloss.JoinHorizontal(lipgloss.Top, prompt, input, countStr)
	return searchBarStyle.Width(m.width).Render(content)
}

func (m model) emptyView(msg string) string {
	availW := m.width
	bodyH := m.bodyHeight()

	content := lipgloss.NewStyle().
		Foreground(colorDim).
		Italic(true).
		Render(msg)

	full := lipgloss.Place(availW, bodyH,
		lipgloss.Center, lipgloss.Center,
		content,
		lipgloss.WithWhitespaceBackground(colorBase),
	)

	return full
}

func (m model) renderEsqlBar() string {
	query := m.esqlQuery
	if !m.esqlRunning {
		query = m.esqlInput.Value()
	}
	prompt := searchPromptStyle.Render(" esql> ")
	input := ""
	if m.esqlRunning {
		input = lipgloss.NewStyle().Foreground(colorDim).Render(query + " ...")
	} else {
		input = m.esqlInput.View()
	}
	if input == "" {
		input = " "
	}
	content := lipgloss.JoinHorizontal(lipgloss.Top, prompt, input)
	return searchBarStyle.Width(m.width).Render(content)
}

func (m model) renderEsqlLoadingView() string {
	msg := fmt.Sprintf(" Running ES|QL query...\n\n  %s",
		lipgloss.NewStyle().Foreground(colorDim).Render(m.esqlQuery))

	box := overlayBoxStyle.
		Width(60).
		Render(lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(" ES|QL Query "),
			"",
			msg,
			"",
			lipgloss.NewStyle().Foreground(colorDim).Render(" Please wait "),
		))

	return box
}

func (m model) renderEsqlErrorView() string {
	title := lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render(" ES|QL Error ")

	errMsg := lipgloss.NewStyle().Foreground(colorText).Render(fmt.Sprintf("  %v", m.esqlError))
	queryInfo := lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf("  Query: %s", m.esqlQuery))

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		errMsg,
		"",
		queryInfo,
		"",
		lipgloss.NewStyle().Foreground(colorDim).Render(" Press Esc or Q to retry "),
	)

	availW := max(min(m.width-8, 60), 40)
	box := overlayBoxStyle.
		BorderForeground(colorRed).
		Width(availW).
		Render(content)

	return box
}
