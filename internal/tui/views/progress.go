package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/virakt/imgshrink/internal/compressor"
	"github.com/virakt/imgshrink/internal/tui"
)

// ProgressView represents the compression progress view
type ProgressView struct {
	styles       *tui.Styles
	spinner      spinner.Model
	progress     progress.Model
	files        []string
	results      []*compressor.CompressionResult
	currentIndex int
	currentFile  string
	done         bool
	startTime    time.Time
	width        int
	height       int
}

// ProgressMsg is sent when a file compression completes
type ProgressMsg struct {
	Result *compressor.CompressionResult
}

// AllDoneMsg is sent when all files are compressed
type AllDoneMsg struct{}

// NewProgressView creates a new progress view
func NewProgressView(styles *tui.Styles, files []string) ProgressView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.TextSuccess

	p := progress.New(progress.WithDefaultGradient())

	return ProgressView{
		styles:       styles,
		spinner:      s,
		progress:     p,
		files:        files,
		results:      make([]*compressor.CompressionResult, 0, len(files)),
		currentIndex: 0,
		startTime:    time.Now(),
	}
}

// Init initializes the progress view
func (v ProgressView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.progress.Init())
}

// Update handles messages for the progress view
func (v ProgressView) Update(msg tea.Msg) (ProgressView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.progress.Width = msg.Width - 20

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd

	case progress.FrameMsg:
		progressModel, cmd := v.progress.Update(msg)
		v.progress = progressModel.(progress.Model)
		return v, cmd

	case ProgressMsg:
		v.results = append(v.results, msg.Result)
		v.currentIndex++

		if v.currentIndex >= len(v.files) {
			v.done = true
			return v, func() tea.Msg { return AllDoneMsg{} }
		}

		v.currentFile = v.files[v.currentIndex]
		return v, nil

	case AllDoneMsg:
		v.done = true
		return v, nil
	}

	return v, nil
}

// View renders the progress view
func (v ProgressView) View() string {
	var b strings.Builder

	b.WriteString(v.styles.Title.Render("🔄 Compressing Images"))
	b.WriteString("\n\n")

	if v.done {
		b.WriteString(v.styles.TextSuccess.Render("✓ Compression Complete!"))
		b.WriteString("\n\n")
	} else {
		// Current file
		b.WriteString(v.spinner.View())
		b.WriteString(" ")
		if v.currentIndex < len(v.files) {
			b.WriteString(v.styles.Text.Render(fmt.Sprintf("Processing: %s", v.files[v.currentIndex])))
		}
		b.WriteString("\n\n")
	}

	// Progress bar
	percent := float64(v.currentIndex) / float64(len(v.files))
	b.WriteString(v.progress.ViewAs(percent))
	b.WriteString("\n")
	b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf("  %d / %d files", v.currentIndex, len(v.files))))
	b.WriteString("\n\n")

	// Results so far
	if len(v.results) > 0 {
		b.WriteString(v.styles.TextBold.Render("Results:"))
		b.WriteString("\n")

		// Show last 5 results
		start := 0
		if len(v.results) > 5 {
			start = len(v.results) - 5
		}

		for i := start; i < len(v.results); i++ {
			result := v.results[i]
			if result.Success {
				reduction := tui.RenderReduction(result.Reduction, v.styles)
				b.WriteString(v.styles.TextSuccess.Render("  ✓ "))
				b.WriteString(v.styles.Text.Render(fmt.Sprintf("%s ", result.InputPath)))
				b.WriteString(reduction)
			} else {
				b.WriteString(v.styles.TextError.Render("  ✗ "))
				b.WriteString(v.styles.Text.Render(result.InputPath))
				if result.Error != nil {
					b.WriteString(v.styles.TextError.Render(fmt.Sprintf(" (%s)", result.Error.Error())))
				}
			}
			b.WriteString("\n")
		}

		if start > 0 {
			b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf("  ... and %d more\n", start)))
		}
	}

	// Elapsed time
	elapsed := time.Since(v.startTime)
	b.WriteString("\n")
	b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf("Elapsed: %s", elapsed.Round(time.Second))))

	// Help
	if v.done {
		b.WriteString("\n\n")
		b.WriteString(v.styles.HelpKey.Render("→") + v.styles.HelpDesc.Render(" view results"))
	}

	return v.styles.Content.Render(b.String())
}

// GetResults returns all compression results
func (v ProgressView) GetResults() []*compressor.CompressionResult {
	return v.results
}

// IsDone returns true if all files have been processed
func (v ProgressView) IsDone() bool {
	return v.done
}

// SetCurrentFile sets the current file being processed
func (v *ProgressView) SetCurrentFile(file string) {
	v.currentFile = file
}

// AddResult adds a compression result
func (v *ProgressView) AddResult(result *compressor.CompressionResult) {
	v.results = append(v.results, result)
	v.currentIndex++
	if v.currentIndex >= len(v.files) {
		v.done = true
	}
}
