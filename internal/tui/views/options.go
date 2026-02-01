package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/virakt/imgshrink/internal/compressor"
	"github.com/virakt/imgshrink/internal/tui"
)

// OptionsView represents the compression options view
type OptionsView struct {
	options    compressor.CompressionOptions
	styles     *tui.Styles
	focusIndex int
	inputs     []textinput.Model
	format     compressor.ImageFormat
	width      int
	height     int
}

// Option field indices
const (
	optQuality = iota
	optResizePercent
	optResizeWidth
	optResizeHeight
	optOutputDir
	optOutputSuffix
	optCompressionLevel
	optFieldCount
)

// NewOptionsView creates a new options view
func NewOptionsView(styles *tui.Styles, format compressor.ImageFormat) OptionsView {
	options := compressor.DefaultOptions()

	inputs := make([]textinput.Model, optFieldCount)

	// Quality input (JPEG)
	inputs[optQuality] = textinput.New()
	inputs[optQuality].Placeholder = "85"
	inputs[optQuality].SetValue(strconv.Itoa(options.Quality))
	inputs[optQuality].CharLimit = 3
	inputs[optQuality].Width = 10

	// Resize percent
	inputs[optResizePercent] = textinput.New()
	inputs[optResizePercent].Placeholder = "0 (no resize)"
	inputs[optResizePercent].CharLimit = 3
	inputs[optResizePercent].Width = 10

	// Resize width
	inputs[optResizeWidth] = textinput.New()
	inputs[optResizeWidth].Placeholder = "0 (auto)"
	inputs[optResizeWidth].CharLimit = 5
	inputs[optResizeWidth].Width = 10

	// Resize height
	inputs[optResizeHeight] = textinput.New()
	inputs[optResizeHeight].Placeholder = "0 (auto)"
	inputs[optResizeHeight].CharLimit = 5
	inputs[optResizeHeight].Width = 10

	// Output directory
	inputs[optOutputDir] = textinput.New()
	inputs[optOutputDir].Placeholder = "Same as input"
	inputs[optOutputDir].CharLimit = 256
	inputs[optOutputDir].Width = 40

	// Output suffix
	inputs[optOutputSuffix] = textinput.New()
	inputs[optOutputSuffix].Placeholder = "_compressed"
	inputs[optOutputSuffix].SetValue(options.OutputSuffix)
	inputs[optOutputSuffix].CharLimit = 32
	inputs[optOutputSuffix].Width = 20

	// Compression level (PNG)
	inputs[optCompressionLevel] = textinput.New()
	inputs[optCompressionLevel].Placeholder = "6"
	inputs[optCompressionLevel].SetValue(strconv.Itoa(options.CompressionLevel))
	inputs[optCompressionLevel].CharLimit = 1
	inputs[optCompressionLevel].Width = 10

	// Focus first input
	inputs[0].Focus()

	return OptionsView{
		options:    options,
		styles:     styles,
		inputs:     inputs,
		format:     format,
		focusIndex: 0,
	}
}

// Init initializes the options view
func (v OptionsView) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the options view
func (v OptionsView) Update(msg tea.Msg) (OptionsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down", "j":
			v.focusIndex++
			if v.focusIndex >= len(v.inputs) {
				v.focusIndex = 0
			}
			return v, v.updateFocus()

		case "shift+tab", "up", "k":
			v.focusIndex--
			if v.focusIndex < 0 {
				v.focusIndex = len(v.inputs) - 1
			}
			return v, v.updateFocus()

		case "p":
			// Toggle progressive (JPEG)
			v.options.Progressive = !v.options.Progressive
			return v, nil

		case "i":
			// Toggle interlaced (PNG)
			v.options.Interlaced = !v.options.Interlaced
			return v, nil

		case "m":
			// Toggle strip metadata
			v.options.StripMetadata = !v.options.StripMetadata
			return v, nil

		case "1":
			v.options.ChromaSubsample = "4:4:4"
			return v, nil
		case "2":
			v.options.ChromaSubsample = "4:2:2"
			return v, nil
		case "3":
			v.options.ChromaSubsample = "4:2:0"
			return v, nil
		}
	}

	// Update the focused input
	cmd := v.updateInputs(msg)
	v.parseInputs()

	return v, cmd
}

func (v *OptionsView) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(v.inputs))
	for i := range v.inputs {
		if i == v.focusIndex {
			cmds[i] = v.inputs[i].Focus()
		} else {
			v.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (v *OptionsView) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(v.inputs))
	for i := range v.inputs {
		v.inputs[i], cmds[i] = v.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (v *OptionsView) parseInputs() {
	// Parse quality
	if val, err := strconv.Atoi(v.inputs[optQuality].Value()); err == nil {
		if val >= 1 && val <= 100 {
			v.options.Quality = val
		}
	}

	// Parse resize percent
	if val, err := strconv.ParseFloat(v.inputs[optResizePercent].Value(), 64); err == nil {
		if val >= 0 && val <= 100 {
			v.options.ResizePercent = val
		}
	}

	// Parse resize width
	if val, err := strconv.Atoi(v.inputs[optResizeWidth].Value()); err == nil {
		if val >= 0 {
			v.options.ResizeWidth = val
		}
	}

	// Parse resize height
	if val, err := strconv.Atoi(v.inputs[optResizeHeight].Value()); err == nil {
		if val >= 0 {
			v.options.ResizeHeight = val
		}
	}

	// Parse output directory
	v.options.OutputDir = v.inputs[optOutputDir].Value()

	// Parse output suffix
	if v.inputs[optOutputSuffix].Value() != "" {
		v.options.OutputSuffix = v.inputs[optOutputSuffix].Value()
	}

	// Parse compression level
	if val, err := strconv.Atoi(v.inputs[optCompressionLevel].Value()); err == nil {
		if val >= 0 && val <= 9 {
			v.options.CompressionLevel = val
		}
	}
}

// View renders the options view
func (v OptionsView) View() string {
	var b strings.Builder

	b.WriteString(v.styles.Title.Render("⚙ Compression Options"))
	b.WriteString("\n\n")

	// Common options section
	b.WriteString(v.styles.TextBold.Render("Common Options"))
	b.WriteString("\n")
	b.WriteString(v.renderInput("Resize %", optResizePercent, "Percentage to resize (0 = no resize)"))
	b.WriteString(v.renderInput("Width", optResizeWidth, "Target width in pixels (0 = auto)"))
	b.WriteString(v.renderInput("Height", optResizeHeight, "Target height in pixels (0 = auto)"))
	b.WriteString(v.renderInput("Output Dir", optOutputDir, "Output directory (empty = same as input)"))
	b.WriteString(v.renderInput("Suffix", optOutputSuffix, "Suffix for output filename"))
	b.WriteString("\n")

	// Toggle options
	b.WriteString(v.renderToggle("Strip Metadata", v.options.StripMetadata, "m"))
	b.WriteString("\n")

	// Format-specific options
	if v.format == compressor.FormatJPEG {
		b.WriteString(v.styles.TextBold.Render("JPEG Options"))
		b.WriteString("\n")
		b.WriteString(v.renderInput("Quality", optQuality, "1-100 (higher = better quality, larger file)"))
		b.WriteString(v.renderToggle("Progressive", v.options.Progressive, "p"))
		b.WriteString("\n")

		// Chroma subsampling
		b.WriteString(v.styles.TextMuted.Render("  Chroma Subsampling: "))
		subsampleOptions := []string{"4:4:4", "4:2:2", "4:2:0"}
		for i, opt := range subsampleOptions {
			if v.options.ChromaSubsample == opt {
				b.WriteString(v.styles.ButtonActive.Render(opt))
			} else {
				b.WriteString(v.styles.Button.Render(opt))
			}
			b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf(" [%d] ", i+1)))
		}
		b.WriteString("\n")
	} else if v.format == compressor.FormatPNG {
		b.WriteString(v.styles.TextBold.Render("PNG Options"))
		b.WriteString("\n")
		b.WriteString(v.renderInput("Compression", optCompressionLevel, "0-9 (higher = more compression)"))
		b.WriteString(v.renderToggle("Interlaced", v.options.Interlaced, "i"))
	}

	// Help
	b.WriteString("\n\n")
	help := []string{
		v.styles.HelpKey.Render("tab/↑↓") + v.styles.HelpDesc.Render(" navigate"),
		v.styles.HelpKey.Render("←") + v.styles.HelpDesc.Render(" back"),
		v.styles.HelpKey.Render("→") + v.styles.HelpDesc.Render(" compress"),
	}
	b.WriteString(strings.Join(help, "  "))

	return v.styles.Content.Render(b.String())
}

func (v OptionsView) renderInput(label string, index int, hint string) string {
	var b strings.Builder

	focused := v.focusIndex == index
	labelStyle := v.styles.TextMuted
	if focused {
		labelStyle = v.styles.TextBold
	}

	b.WriteString(labelStyle.Render(fmt.Sprintf("  %-12s ", label)))

	if focused {
		b.WriteString(v.styles.InputFocused.Render(v.inputs[index].View()))
	} else {
		b.WriteString(v.styles.Input.Render(v.inputs[index].View()))
	}

	b.WriteString("  ")
	b.WriteString(v.styles.TextMuted.Render(hint))
	b.WriteString("\n")

	return b.String()
}

func (v OptionsView) renderToggle(label string, value bool, key string) string {
	var b strings.Builder

	b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf("  %-12s ", label)))

	if value {
		b.WriteString(v.styles.TextSuccess.Render("● ON "))
	} else {
		b.WriteString(v.styles.TextMuted.Render("○ OFF"))
	}

	b.WriteString(v.styles.TextMuted.Render(fmt.Sprintf("  [%s]", key)))
	b.WriteString("\n")

	return b.String()
}

// GetOptions returns the current compression options
func (v OptionsView) GetOptions() compressor.CompressionOptions {
	return v.options
}

// SetFormat sets the image format for format-specific options
func (v *OptionsView) SetFormat(format compressor.ImageFormat) {
	v.format = format
}
