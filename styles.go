package main

import "github.com/charmbracelet/lipgloss"

var (
	colorError  = lipgloss.Color("#F38BA8")
	colorWarn   = lipgloss.Color("#F9E2AF")
	colorInfo   = lipgloss.Color("#89B4FA")
	colorDebug  = lipgloss.Color("#CBA6F7")
	colorTrace  = lipgloss.Color("#6C7086")
	colorBorder = lipgloss.Color("#45475A")
	colorText   = lipgloss.Color("#CDD6F4")
	colorDim    = lipgloss.Color("#585B70")
	colorAccent = lipgloss.Color("#89B4FA")
	colorGreen  = lipgloss.Color("#A6E3A1")
	colorRed    = lipgloss.Color("#F38BA8")
	colorSurface = lipgloss.Color("#313244")
	colorBase    = lipgloss.Color("#1E1E2E")
	colorMantle  = lipgloss.Color("#181825")

	headerStyle = lipgloss.NewStyle().
			Background(colorMantle).
			Foreground(colorText).
			PaddingLeft(1)

	tableHeaderStyle = lipgloss.NewStyle().
				Background(colorSurface).
				Foreground(lipgloss.Color("#BAC2DE")).
				Bold(true)

	selectedRowStyle = lipgloss.NewStyle().
				Background(colorSurface)

	altRowStyle = lipgloss.NewStyle().
			Background(colorBase)

	footerStyle = lipgloss.NewStyle().
			Background(colorMantle).
			Foreground(colorDim).
			PaddingLeft(1)

	overlayBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Background(colorBase).
				Padding(1, 2)

	overlayTitleStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	levelStyles = map[string]lipgloss.Style{
		"ERROR": lipgloss.NewStyle().Foreground(colorError).Bold(true),
		"WARN":  lipgloss.NewStyle().Foreground(colorWarn).Bold(true),
		"INFO":  lipgloss.NewStyle().Foreground(colorInfo),
		"DEBUG": lipgloss.NewStyle().Foreground(colorDebug),
		"TRACE": lipgloss.NewStyle().Foreground(colorTrace),
	}

	searchBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorBorder).
			PaddingLeft(1)

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(colorWarn).
				Bold(true)

	helpStyle = lipgloss.NewStyle().Foreground(colorDim)
	keyStyle  = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
)

func levelColor(level string) lipgloss.Color {
	switch level {
	case "ERROR":
		return colorError
	case "WARN":
		return colorWarn
	case "INFO":
		return colorInfo
	case "DEBUG":
		return colorDebug
	case "TRACE":
		return colorTrace
	default:
		return colorText
	}
}
