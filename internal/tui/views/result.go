package views

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/virakt/imgshrink/internal/compressor"
	"github.com/virakt/imgshrink/internal/tui"
)

// ResultView represents the compression results view
type ResultView struct {
	styles       *tui.Styles
	results      []*compressor.CompressionResult
	cursor       int
	scrollOffset int
	width        int
	height       int
}

// NewResultView creates a new result view
func NewResultView(styles *tui.Styles, results []*compressor.CompressionResult) ResultView {
	return ResultView{
		styles:  styles,
		results: results,
		cursor:  0,
	}
}

// Init initializes the result view
func (v ResultView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the result view
func (v ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.cursor > 0 {
				v.cursor--
				if v.cursor < v.scrollOffset {
					v.scrollOffset = v.cursor
				}
			}

		case "down", "j":
			if v.cursor < len(v.results)-1 {
				v.cursor++
				maxVisible := v.height - 15
				if maxVisible < 5 {
					maxVisible = 5
				}
				if v.cursor >= v.scrollOffset+maxVisible {
					v.scrollOffset = v.cursor - maxVisible + 1
				}
			}

		case "home", "g":
			v.cursor = 0
			v.scrollOffset = 0

		case "end", "G":
			v.cursor = len(v.results) - 1
			maxVisible := v.height - 15
			if maxVisible < 5 {
				maxVisible = 5
			}
			if v.cursor >= maxVisible {
				v.scrollOffset = v.cursor - maxVisible + 1
			}
		}
	}

	return v, nil
}

// View renders the result view
func (v ResultView) View() string {
	var b strings.Builder

	b.WriteString(v.styles.Title.Render("📊 Compression Results"))
	b.WriteString("\n\n")

	// Summary statistics
	stats := v.calculateStats()

	summaryBox := v.styles.Box.Render(
		v.styles.TextBold.Render("Summary\n\n") +
			tui.RenderKeyValue("Total Files", fmt.Sprintf("%d", stats.totalFiles), v.styles) + "\n" +
			tui.RenderKeyValue("Successful", v.styles.TextSuccess.Render(fmt.Sprintf("%d", stats.successCount)), v.styles) + "\n" +
			tui.RenderKeyValue("Failed", v.styles.TextError.Render(fmt.Sprintf("%d", stats.failCount)), v.styles) + "\n" +
			tui.RenderKeyValue("Original Size", compressor.FormatBytes(stats.totalInput), v.styles) + "\n" +
			tui.RenderKeyValue("Compressed Size", compressor.FormatBytes(stats.totalOutput), v.styles) + "\n" +
			tui.RenderKeyValue("Total Saved", v.styles.TextSuccess.Render(compressor.FormatBytes(stats.totalInput-stats.totalOutput)), v.styles) + "\n" +
			tui.RenderKeyValue("Reduction", tui.RenderReduction(stats.avgReduction, v.styles), v.styles),
	)
	b.WriteString(summaryBox)
	b.WriteString("\n\n")

	// Results list
	b.WriteString(v.styles.TextBold.Render("File Details:"))
	b.WriteString("\n")

	maxVisible := v.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	endIndex := v.scrollOffset + maxVisible
	if endIndex > len(v.results) {
		endIndex = len(v.results)
	}

	for i := v.scrollOffset; i < endIndex; i++ {
		result := v.results[i]
		isSelected := i == v.cursor

		prefix := "  "
		if isSelected {
			prefix = "▸ "
		}

		fileName := filepath.Base(result.InputPath)

		var line string
		if result.Success {
			reduction := fmt.Sprintf("%.1f%%", result.Reduction)
			sizeInfo := fmt.Sprintf("%s → %s",
				compressor.FormatBytes(result.InputSize),
				compressor.FormatBytes(result.OutputSize))

			if isSelected {
				line = v.styles.ListItemSelected.Render(prefix+"✓ "+fileName) +
					v.styles.TextMuted.Render(" ("+sizeInfo+") ") +
					tui.RenderReduction(result.Reduction, v.styles)
			} else {
				line = v.styles.TextSuccess.Render(prefix+"✓ ") +
					v.styles.Text.Render(fileName) +
					v.styles.TextMuted.Render(" ("+sizeInfo+") ") +
					v.styles.TextMuted.Render(reduction)
			}
		} else {
			errMsg := "unknown error"
			if result.Error != nil {
				errMsg = result.Error.Error()
			}

			if isSelected {
				line = v.styles.ListItemSelected.Render(prefix+"✗ "+fileName) +
					v.styles.TextError.Render(" - "+errMsg)
			} else {
				line = v.styles.TextError.Render(prefix+"✗ ") +
					v.styles.Text.Render(fileName) +
					v.styles.TextError.Render(" - "+errMsg)
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(v.results) > maxVisible {
		b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf("\n  Showing %d-%d of %d",
			v.scrollOffset+1, endIndex, len(v.results))))
	}

	// Selected file details
	if v.cursor < len(v.results) {
		result := v.results[v.cursor]
		b.WriteString("\n\n")
		b.WriteString(v.styles.TextBold.Render("Selected File Details:"))
		b.WriteString("\n")

		detailsBox := v.styles.FileInfo.Render(
			tui.RenderKeyValue("Input", result.InputPath, v.styles) + "\n" +
				tui.RenderKeyValue("Output", result.OutputPath, v.styles) + "\n" +
				tui.RenderKeyValue("Dimensions", fmt.Sprintf("%dx%d", result.Width, result.Height), v.styles),
		)
		b.WriteString(detailsBox)
	}

	// Help
	b.WriteString("\n\n")
	help := []string{
		v.styles.HelpKey.Render("↑↓") + v.styles.HelpDesc.Render(" navigate"),
		v.styles.HelpKey.Render("q") + v.styles.HelpDesc.Render(" quit"),
		v.styles.HelpKey.Render("r") + v.styles.HelpDesc.Render(" restart"),
	}
	b.WriteString(strings.Join(help, "  "))

	return v.styles.Content.Render(b.String())
}

type resultStats struct {
	totalFiles   int
	successCount int
	failCount    int
	totalInput   int64
	totalOutput  int64
	avgReduction float64
}

func (v ResultView) calculateStats() resultStats {
	stats := resultStats{
		totalFiles: len(v.results),
	}

	for _, result := range v.results {
		if result.Success {
			stats.successCount++
			stats.totalInput += result.InputSize
			stats.totalOutput += result.OutputSize
		} else {
			stats.failCount++
		}
	}

	if stats.totalInput > 0 {
		stats.avgReduction = compressor.CalculateReduction(stats.totalInput, stats.totalOutput)
	}

	return stats
}

// GetResults returns all results
func (v ResultView) GetResults() []*compressor.CompressionResult {
	return v.results
}
