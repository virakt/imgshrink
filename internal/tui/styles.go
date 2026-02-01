package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	bgColor        = lipgloss.Color("#1F2937") // Dark gray
	fgColor        = lipgloss.Color("#F9FAFB") // Light gray
)

// Styles contains all the styles used in the TUI
type Styles struct {
	// App styles
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	// Title styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style

	// Content styles
	Content lipgloss.Style
	Box     lipgloss.Style

	// Text styles
	Text        lipgloss.Style
	TextMuted   lipgloss.Style
	TextBold    lipgloss.Style
	TextSuccess lipgloss.Style
	TextError   lipgloss.Style
	TextWarning lipgloss.Style

	// Input styles
	Input        lipgloss.Style
	InputFocused lipgloss.Style

	// Button styles
	Button       lipgloss.Style
	ButtonActive lipgloss.Style

	// List styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style

	// Progress styles
	ProgressBar  lipgloss.Style
	ProgressFill lipgloss.Style

	// Status styles
	StatusBar lipgloss.Style

	// Help styles
	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Tab styles
	Tab       lipgloss.Style
	TabActive lipgloss.Style

	// File info styles
	FileInfo lipgloss.Style
	FilePath lipgloss.Style
	FileSize lipgloss.Style

	// Result styles
	ResultSuccess lipgloss.Style
	ResultError   lipgloss.Style
	ResultStats   lipgloss.Style
}

// DefaultStyles returns the default styles for the TUI
func DefaultStyles() *Styles {
	s := &Styles{}

	// App styles
	s.App = lipgloss.NewStyle().
		Padding(1, 2)

	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(fgColor).
		Background(primaryColor).
		Padding(0, 2).
		MarginBottom(1)

	s.Footer = lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1)

	// Title styles
	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true)

	// Content styles
	s.Content = lipgloss.NewStyle().
		Padding(1, 0)

	s.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2)

	// Text styles
	s.Text = lipgloss.NewStyle().
		Foreground(fgColor)

	s.TextMuted = lipgloss.NewStyle().
		Foreground(mutedColor)

	s.TextBold = lipgloss.NewStyle().
		Bold(true).
		Foreground(fgColor)

	s.TextSuccess = lipgloss.NewStyle().
		Foreground(secondaryColor)

	s.TextError = lipgloss.NewStyle().
		Foreground(errorColor)

	s.TextWarning = lipgloss.NewStyle().
		Foreground(accentColor)

	// Input styles
	s.Input = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1)

	s.InputFocused = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1)

	// Button styles
	s.Button = lipgloss.NewStyle().
		Foreground(fgColor).
		Background(mutedColor).
		Padding(0, 2).
		MarginRight(1)

	s.ButtonActive = lipgloss.NewStyle().
		Foreground(fgColor).
		Background(primaryColor).
		Padding(0, 2).
		MarginRight(1)

	// List styles
	s.ListItem = lipgloss.NewStyle().
		PaddingLeft(2)

	s.ListItemSelected = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(primaryColor).
		Bold(true)

	// Progress styles
	s.ProgressBar = lipgloss.NewStyle().
		Foreground(mutedColor)

	s.ProgressFill = lipgloss.NewStyle().
		Foreground(secondaryColor)

	// Status bar
	s.StatusBar = lipgloss.NewStyle().
		Foreground(mutedColor).
		Background(bgColor).
		Padding(0, 1)

	// Help styles
	s.Help = lipgloss.NewStyle().
		Foreground(mutedColor)

	s.HelpKey = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	s.HelpDesc = lipgloss.NewStyle().
		Foreground(mutedColor)

	// Tab styles
	s.Tab = lipgloss.NewStyle().
		Foreground(mutedColor).
		Padding(0, 2)

	s.TabActive = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 2).
		Underline(true)

	// File info styles
	s.FileInfo = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(1, 2).
		MarginBottom(1)

	s.FilePath = lipgloss.NewStyle().
		Foreground(fgColor).
		Bold(true)

	s.FileSize = lipgloss.NewStyle().
		Foreground(accentColor)

	// Result styles
	s.ResultSuccess = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(secondaryColor).
		Padding(1, 2)

	s.ResultError = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Padding(1, 2)

	s.ResultStats = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true)

	return s
}

// Logo returns the ASCII art logo
func Logo() string {
	logo := `
 ___                 ____  _          _       _    
|_ _|_ __ ___   __ _/ ___|| |__  _ __(_)_ __ | | __
 | || '_ ` + "`" + ` _ \ / _` + "`" + ` \___ \| '_ \| '__| | '_ \| |/ /
 | || | | | | | (_| |___) | | | | |  | | | | |   < 
|___|_| |_| |_|\__, |____/|_| |_|_|  |_|_| |_|_|\_\
               |___/                               
`
	return lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render(logo)
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100)
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}

	return lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render(bar)
}

// RenderKeyValue renders a key-value pair
func RenderKeyValue(key, value string, styles *Styles) string {
	return styles.TextMuted.Render(key+": ") + styles.Text.Render(value)
}

// RenderReduction renders the reduction percentage with color
func RenderReduction(reduction float64, styles *Styles) string {
	var style lipgloss.Style
	switch {
	case reduction >= 50:
		style = styles.TextSuccess
	case reduction >= 20:
		style = styles.TextWarning
	default:
		style = styles.TextMuted
	}
	return style.Render(fmt.Sprintf("%.1f%%", reduction))
}
