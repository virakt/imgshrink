package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/virakt/imgshrink/internal/api"
	"github.com/virakt/imgshrink/internal/compressor"
)

// ViewState represents the current view state
type ViewState int

const (
	ViewHome ViewState = iota
	ViewOptions
	ViewProgress
	ViewResult
)

// Model represents the main TUI application model
type Model struct {
	api    *api.ImageAPI
	styles *Styles
	state  ViewState
	width  int
	height int

	// Selected files and options
	files   []string
	options compressor.CompressionOptions
	results []*compressor.CompressionResult

	// Sub-models for each view
	homeModel     HomeModel
	optionsModel  OptionsModel
	progressModel ProgressModel
	resultModel   ResultModel

	// Error handling
	err error

	// Quit flag
	quitting bool
}

// HomeModel is a simplified home view model
type HomeModel struct {
	files     []string
	cursor    int
	inputPath string
	mode      string // "single" or "batch"
	err       error
}

// OptionsModel is a simplified options view model
type OptionsModel struct {
	options    compressor.CompressionOptions
	focusIndex int
	format     compressor.ImageFormat
}

// ProgressModel is a simplified progress view model
type ProgressModel struct {
	files        []string
	results      []*compressor.CompressionResult
	currentIndex int
	done         bool
}

// ResultModel is a simplified result view model
type ResultModel struct {
	results []*compressor.CompressionResult
	cursor  int
}

// NewModel creates a new TUI application model
func NewModel() Model {
	imageAPI := api.NewImageAPI()
	styles := DefaultStyles()

	return Model{
		api:     imageAPI,
		styles:  styles,
		state:   ViewHome,
		options: compressor.DefaultOptions(),
		homeModel: HomeModel{
			files: []string{},
			mode:  "single",
		},
		optionsModel: OptionsModel{
			options: compressor.DefaultOptions(),
		},
	}
}

// Init initializes the TUI application
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the TUI application
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == ViewResult || m.state == ViewHome {
				m.quitting = true
				return m, tea.Quit
			}

		case "esc":
			// Go back to previous view
			switch m.state {
			case ViewOptions:
				m.state = ViewHome
			case ViewProgress:
				// Can't go back during progress
			case ViewResult:
				m.state = ViewHome
				m.homeModel.files = []string{}
				m.results = nil
			}
			return m, nil

		case "left":
			// Navigate back
			switch m.state {
			case ViewOptions:
				m.state = ViewHome
				return m, nil
			case ViewResult:
				m.state = ViewHome
				m.homeModel.files = []string{}
				return m, nil
			}

		case "right", "enter":
			// Navigate forward
			switch m.state {
			case ViewHome:
				if len(m.homeModel.files) > 0 {
					// Detect format from first file
					if len(m.homeModel.files) > 0 {
						format, _ := compressor.GetImageFormat(m.homeModel.files[0])
						m.optionsModel.format = format
					}
					m.state = ViewOptions
					return m, nil
				}
			case ViewOptions:
				m.options = m.optionsModel.options
				m.files = m.homeModel.files
				m.state = ViewProgress
				m.progressModel = ProgressModel{
					files:   m.files,
					results: []*compressor.CompressionResult{},
				}
				return m, m.startCompression()
			case ViewProgress:
				if m.progressModel.done {
					m.results = m.progressModel.results
					m.resultModel = ResultModel{
						results: m.results,
					}
					m.state = ViewResult
					return m, nil
				}
			}

		case "r":
			// Restart from home
			if m.state == ViewResult {
				m.state = ViewHome
				m.homeModel.files = []string{}
				m.results = nil
				return m, nil
			}
		}

		// View-specific key handling
		switch m.state {
		case ViewHome:
			return m.updateHome(msg)
		case ViewOptions:
			return m.updateOptions(msg)
		case ViewProgress:
			return m.updateProgress(msg)
		case ViewResult:
			return m.updateResult(msg)
		}

	case compressionResultMsg:
		m.progressModel.results = append(m.progressModel.results, msg.result)
		m.progressModel.currentIndex++
		if m.progressModel.currentIndex >= len(m.progressModel.files) {
			m.progressModel.done = true
		}

		// Continue with next file
		if !m.progressModel.done {
			return m, m.compressNext()
		}
		return m, nil
	}

	return m, nil
}

// View renders the TUI application
func (m Model) View() string {
	if m.quitting {
		return m.styles.Text.Render("Goodbye! 👋\n")
	}

	var content string

	switch m.state {
	case ViewHome:
		content = m.viewHome()
	case ViewOptions:
		content = m.viewOptions()
	case ViewProgress:
		content = m.viewProgress()
	case ViewResult:
		content = m.viewResult()
	}

	// Build the full view
	var b strings.Builder

	// Header
	header := m.styles.Header.Render(" 🖼  ImgShrink - Image Compression Tool ")
	b.WriteString(header)
	b.WriteString("\n")

	// Content
	b.WriteString(content)

	// Footer with navigation hints
	footer := m.renderFooter()
	b.WriteString("\n")
	b.WriteString(footer)

	return m.styles.App.Render(b.String())
}

func (m Model) renderFooter() string {
	var hints []string

	switch m.state {
	case ViewHome:
		hints = []string{"a: add file", "d: remove", "→: continue", "q: quit"}
	case ViewOptions:
		hints = []string{"←: back", "→: compress", "tab: next field"}
	case ViewProgress:
		if m.progressModel.done {
			hints = []string{"→: view results"}
		} else {
			hints = []string{"compressing..."}
		}
	case ViewResult:
		hints = []string{"r: restart", "q: quit", "↑↓: navigate"}
	}

	var rendered []string
	for _, hint := range hints {
		parts := strings.SplitN(hint, ":", 2)
		if len(parts) == 2 {
			rendered = append(rendered,
				m.styles.HelpKey.Render(parts[0])+
					m.styles.HelpDesc.Render(":"+parts[1]))
		} else {
			rendered = append(rendered, m.styles.HelpDesc.Render(hint))
		}
	}

	return m.styles.Footer.Render(strings.Join(rendered, "  "))
}

// Home view methods
func (m Model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "a":
		// Add a file (simplified - in real app would use file picker)
		// For now, we'll use a text input approach
		return m, nil

	case "d", "backspace":
		if len(m.homeModel.files) > 0 && m.homeModel.cursor < len(m.homeModel.files) {
			m.homeModel.files = append(
				m.homeModel.files[:m.homeModel.cursor],
				m.homeModel.files[m.homeModel.cursor+1:]...,
			)
			if m.homeModel.cursor >= len(m.homeModel.files) && m.homeModel.cursor > 0 {
				m.homeModel.cursor--
			}
		}

	case "up", "k":
		if m.homeModel.cursor > 0 {
			m.homeModel.cursor--
		}

	case "down", "j":
		if m.homeModel.cursor < len(m.homeModel.files)-1 {
			m.homeModel.cursor++
		}

	case "tab":
		if m.homeModel.mode == "single" {
			m.homeModel.mode = "batch"
		} else {
			m.homeModel.mode = "single"
		}
	}

	return m, nil
}

func (m Model) viewHome() string {
	var b strings.Builder

	// Mode tabs
	singleTab := m.styles.Tab.Render("Single File")
	batchTab := m.styles.Tab.Render("Batch Mode")
	if m.homeModel.mode == "single" {
		singleTab = m.styles.TabActive.Render("● Single File")
	} else {
		batchTab = m.styles.TabActive.Render("● Batch Mode")
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, singleTab, "  ", batchTab))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Subtitle.Render("Add images to compress (JPEG/PNG only):"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString(m.styles.TextMuted.Render("  Drag and drop files or use 'a' to add files"))
	b.WriteString("\n")
	b.WriteString(m.styles.TextMuted.Render("  Pass file paths as command line arguments"))
	b.WriteString("\n\n")

	// File list
	if len(m.homeModel.files) == 0 {
		b.WriteString(m.styles.Box.Render(
			m.styles.TextMuted.Render("No files selected\n\n") +
				m.styles.Text.Render("Usage: imgshrink <file1> [file2] ..."),
		))
	} else {
		b.WriteString(m.styles.TextBold.Render(fmt.Sprintf("Selected Files (%d):", len(m.homeModel.files))))
		b.WriteString("\n")

		for i, file := range m.homeModel.files {
			prefix := "  "
			style := m.styles.ListItem
			if i == m.homeModel.cursor {
				prefix = "▸ "
				style = m.styles.ListItemSelected
			}

			info, _ := m.api.GetImageInfo(file)
			sizeStr := ""
			if info != nil {
				sizeStr = fmt.Sprintf(" (%s, %dx%d)",
					compressor.FormatBytes(info.Size),
					info.Width, info.Height)
			}

			b.WriteString(style.Render(prefix + file + m.styles.TextMuted.Render(sizeStr)))
			b.WriteString("\n")
		}
	}

	if m.homeModel.err != nil {
		b.WriteString("\n")
		b.WriteString(m.styles.TextError.Render("Error: " + m.homeModel.err.Error()))
	}

	return m.styles.Content.Render(b.String())
}

// Options view methods
func (m Model) updateOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.optionsModel.focusIndex++
		if m.optionsModel.focusIndex > 6 {
			m.optionsModel.focusIndex = 0
		}

	case "shift+tab", "up":
		m.optionsModel.focusIndex--
		if m.optionsModel.focusIndex < 0 {
			m.optionsModel.focusIndex = 6
		}

	case "+", "=":
		switch m.optionsModel.focusIndex {
		case 0: // Quality
			if m.optionsModel.options.Quality < 100 {
				m.optionsModel.options.Quality += 5
			}
		case 1: // Compression level
			if m.optionsModel.options.CompressionLevel < 9 {
				m.optionsModel.options.CompressionLevel++
			}
		}

	case "-", "_":
		switch m.optionsModel.focusIndex {
		case 0: // Quality
			if m.optionsModel.options.Quality > 1 {
				m.optionsModel.options.Quality -= 5
			}
		case 1: // Compression level
			if m.optionsModel.options.CompressionLevel > 0 {
				m.optionsModel.options.CompressionLevel--
			}
		}

	case "p":
		m.optionsModel.options.Progressive = !m.optionsModel.options.Progressive

	case "i":
		m.optionsModel.options.Interlaced = !m.optionsModel.options.Interlaced

	case "m":
		m.optionsModel.options.StripMetadata = !m.optionsModel.options.StripMetadata
	}

	return m, nil
}

func (m Model) viewOptions() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("⚙ Compression Options"))
	b.WriteString("\n\n")

	opts := m.optionsModel.options
	format := m.optionsModel.format

	// Quality slider (JPEG)
	if format == compressor.FormatJPEG {
		b.WriteString(m.renderOption(0, "Quality",
			fmt.Sprintf("%d%%", opts.Quality),
			RenderProgressBar(float64(opts.Quality), 20),
			"+/- to adjust"))
	}

	// Compression level (PNG)
	if format == compressor.FormatPNG {
		b.WriteString(m.renderOption(1, "Compression",
			fmt.Sprintf("%d/9", opts.CompressionLevel),
			RenderProgressBar(float64(opts.CompressionLevel)*100/9, 20),
			"+/- to adjust"))
	}

	b.WriteString("\n")
	b.WriteString(m.styles.TextBold.Render("Toggle Options:"))
	b.WriteString("\n")

	// Toggle options
	b.WriteString(m.renderToggle("Progressive", opts.Progressive, "p"))
	b.WriteString(m.renderToggle("Strip Metadata", opts.StripMetadata, "m"))
	if format == compressor.FormatPNG {
		b.WriteString(m.renderToggle("Interlaced", opts.Interlaced, "i"))
	}

	b.WriteString("\n")
	b.WriteString(m.styles.TextBold.Render("Output Settings:"))
	b.WriteString("\n")
	b.WriteString(m.styles.TextMuted.Render(fmt.Sprintf("  Suffix: %s", opts.OutputSuffix)))
	b.WriteString("\n")

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "(same as input)"
	}
	b.WriteString(m.styles.TextMuted.Render(fmt.Sprintf("  Output Dir: %s", outputDir)))

	return m.styles.Content.Render(b.String())
}

func (m Model) renderOption(index int, label, value, bar, hint string) string {
	var b strings.Builder

	focused := m.optionsModel.focusIndex == index
	labelStyle := m.styles.TextMuted
	if focused {
		labelStyle = m.styles.TextBold
	}

	prefix := "  "
	if focused {
		prefix = "▸ "
	}

	b.WriteString(labelStyle.Render(fmt.Sprintf("%s%-15s", prefix, label)))
	b.WriteString(m.styles.Text.Render(fmt.Sprintf("%-6s ", value)))
	b.WriteString(bar)
	b.WriteString("  ")
	b.WriteString(m.styles.TextMuted.Render(hint))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderToggle(label string, value bool, key string) string {
	var b strings.Builder

	b.WriteString(m.styles.TextMuted.Render(fmt.Sprintf("  %-15s ", label)))

	if value {
		b.WriteString(m.styles.TextSuccess.Render("● ON "))
	} else {
		b.WriteString(m.styles.TextMuted.Render("○ OFF"))
	}

	b.WriteString(m.styles.TextMuted.Render(fmt.Sprintf("  [%s]", key)))
	b.WriteString("\n")

	return b.String()
}

// Progress view methods
func (m Model) updateProgress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Progress view doesn't handle many keys
	return m, nil
}

func (m Model) viewProgress() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("🔄 Compressing Images"))
	b.WriteString("\n\n")

	progress := m.progressModel
	total := len(progress.files)
	current := progress.currentIndex

	if progress.done {
		b.WriteString(m.styles.TextSuccess.Render("✓ Compression Complete!"))
		b.WriteString("\n\n")
	} else if current < total {
		b.WriteString(m.styles.Text.Render(fmt.Sprintf("Processing: %s", progress.files[current])))
		b.WriteString("\n\n")
	}

	// Progress bar
	percent := float64(current) / float64(total) * 100
	b.WriteString(RenderProgressBar(percent, 40))
	b.WriteString("\n")
	b.WriteString(m.styles.TextMuted.Render(fmt.Sprintf("  %d / %d files (%.0f%%)", current, total, percent)))
	b.WriteString("\n\n")

	// Recent results
	if len(progress.results) > 0 {
		b.WriteString(m.styles.TextBold.Render("Recent:"))
		b.WriteString("\n")

		start := 0
		if len(progress.results) > 5 {
			start = len(progress.results) - 5
		}

		for i := start; i < len(progress.results); i++ {
			result := progress.results[i]
			if result.Success {
				b.WriteString(m.styles.TextSuccess.Render("  ✓ "))
				b.WriteString(m.styles.Text.Render(result.InputPath))
				b.WriteString(" ")
				b.WriteString(RenderReduction(result.Reduction, m.styles))
			} else {
				b.WriteString(m.styles.TextError.Render("  ✗ "))
				b.WriteString(m.styles.Text.Render(result.InputPath))
			}
			b.WriteString("\n")
		}
	}

	return m.styles.Content.Render(b.String())
}

// Result view methods
func (m Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.resultModel.cursor > 0 {
			m.resultModel.cursor--
		}
	case "down", "j":
		if m.resultModel.cursor < len(m.resultModel.results)-1 {
			m.resultModel.cursor++
		}
	}
	return m, nil
}

func (m Model) viewResult() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("📊 Compression Results"))
	b.WriteString("\n\n")

	results := m.resultModel.results

	// Calculate stats
	var totalInput, totalOutput int64
	var successCount, failCount int
	for _, r := range results {
		if r.Success {
			successCount++
			totalInput += r.InputSize
			totalOutput += r.OutputSize
		} else {
			failCount++
		}
	}

	reduction := compressor.CalculateReduction(totalInput, totalOutput)

	// Summary box
	summary := fmt.Sprintf(
		"Files: %d successful, %d failed\n"+
			"Original: %s → Compressed: %s\n"+
			"Total Saved: %s (%.1f%% reduction)",
		successCount, failCount,
		compressor.FormatBytes(totalInput),
		compressor.FormatBytes(totalOutput),
		compressor.FormatBytes(totalInput-totalOutput),
		reduction,
	)
	b.WriteString(m.styles.Box.Render(summary))
	b.WriteString("\n\n")

	// Results list
	b.WriteString(m.styles.TextBold.Render("Details:"))
	b.WriteString("\n")

	for i, result := range results {
		prefix := "  "
		if i == m.resultModel.cursor {
			prefix = "▸ "
		}

		if result.Success {
			b.WriteString(m.styles.TextSuccess.Render(prefix + "✓ "))
			b.WriteString(m.styles.Text.Render(result.InputPath))
			b.WriteString(m.styles.TextMuted.Render(fmt.Sprintf(" (%s → %s) ",
				compressor.FormatBytes(result.InputSize),
				compressor.FormatBytes(result.OutputSize))))
			b.WriteString(RenderReduction(result.Reduction, m.styles))
		} else {
			b.WriteString(m.styles.TextError.Render(prefix + "✗ "))
			b.WriteString(m.styles.Text.Render(result.InputPath))
			if result.Error != nil {
				b.WriteString(m.styles.TextError.Render(" - " + result.Error.Error()))
			}
		}
		b.WriteString("\n")
	}

	return m.styles.Content.Render(b.String())
}

// Compression commands
type compressionResultMsg struct {
	result *compressor.CompressionResult
}

func (m Model) startCompression() tea.Cmd {
	if len(m.files) == 0 {
		return nil
	}

	return m.compressFile(0)
}

func (m Model) compressNext() tea.Cmd {
	nextIndex := m.progressModel.currentIndex
	if nextIndex >= len(m.files) {
		return nil
	}
	return m.compressFile(nextIndex)
}

func (m Model) compressFile(index int) tea.Cmd {
	return func() tea.Msg {
		if index >= len(m.files) {
			return nil
		}

		result, _ := m.api.CompressImage(m.files[index], m.options)
		return compressionResultMsg{result: result}
	}
}

// AddFiles adds files to the home model
func (m *Model) AddFiles(files []string) {
	for _, file := range files {
		if err := m.api.ValidateImage(file); err == nil {
			// Check if already added
			found := false
			for _, f := range m.homeModel.files {
				if f == file {
					found = true
					break
				}
			}
			if !found {
				m.homeModel.files = append(m.homeModel.files, file)
			}
		}
	}
}

// Run starts the TUI application
func Run(files []string) error {
	m := NewModel()
	m.AddFiles(files)

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
