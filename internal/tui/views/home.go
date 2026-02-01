package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/virakt/imgshrink/internal/api"
	"github.com/virakt/imgshrink/internal/compressor"
	"github.com/virakt/imgshrink/internal/tui"
)

// HomeView represents the home/file selection view
type HomeView struct {
	filepicker   filepicker.Model
	selectedFile string
	files        []string
	cursor       int
	api          *api.ImageAPI
	styles       *tui.Styles
	width        int
	height       int
	err          error
	mode         string // "single" or "batch"
	directory    string
}

// NewHomeView creates a new home view
func NewHomeView(imageAPI *api.ImageAPI, styles *tui.Styles) HomeView {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".jpg", ".jpeg", ".png"}
	fp.CurrentDirectory, _ = os.Getwd()

	return HomeView{
		filepicker: fp,
		api:        imageAPI,
		styles:     styles,
		mode:       "single",
		files:      []string{},
	}
}

// Init initializes the home view
func (v HomeView) Init() tea.Cmd {
	return v.filepicker.Init()
}

// Update handles messages for the home view
func (v HomeView) Update(msg tea.Msg) (HomeView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Toggle between single and batch mode
			if v.mode == "single" {
				v.mode = "batch"
			} else {
				v.mode = "single"
			}
			return v, nil

		case "enter":
			if v.mode == "batch" && v.directory != "" {
				// Scan directory for images
				files, err := v.api.ScanDirectory(v.directory, true)
				if err != nil {
					v.err = err
				} else {
					v.files = files
				}
			}

		case "backspace":
			if len(v.files) > 0 && v.cursor < len(v.files) {
				// Remove selected file
				v.files = append(v.files[:v.cursor], v.files[v.cursor+1:]...)
				if v.cursor >= len(v.files) && v.cursor > 0 {
					v.cursor--
				}
			}

		case "up", "k":
			if v.cursor > 0 {
				v.cursor--
			}

		case "down", "j":
			if v.cursor < len(v.files)-1 {
				v.cursor++
			}
		}
	}

	// Update file picker
	var cmd tea.Cmd
	v.filepicker, cmd = v.filepicker.Update(msg)

	// Check if a file was selected
	if didSelect, path := v.filepicker.DidSelectFile(msg); didSelect {
		// Validate the file
		if err := v.api.ValidateImage(path); err != nil {
			v.err = err
		} else {
			v.selectedFile = path
			// Add to files list if not already present
			found := false
			for _, f := range v.files {
				if f == path {
					found = true
					break
				}
			}
			if !found {
				v.files = append(v.files, path)
			}
			v.err = nil
		}
	}

	// Check if a directory was selected
	if didSelect, path := v.filepicker.DidSelectDisabledFile(msg); didSelect {
		// It might be a directory
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			v.directory = path
		}
	}

	return v, cmd
}

// View renders the home view
func (v HomeView) View() string {
	var b strings.Builder

	// Mode tabs
	singleTab := v.styles.Tab.Render("Single File")
	batchTab := v.styles.Tab.Render("Batch Mode")
	if v.mode == "single" {
		singleTab = v.styles.TabActive.Render("● Single File")
	} else {
		batchTab = v.styles.TabActive.Render("● Batch Mode")
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, singleTab, "  ", batchTab))
	b.WriteString("\n\n")

	// File picker
	b.WriteString(v.styles.Subtitle.Render("Select image(s) to compress:"))
	b.WriteString("\n\n")
	b.WriteString(v.filepicker.View())
	b.WriteString("\n")

	// Error message
	if v.err != nil {
		b.WriteString("\n")
		b.WriteString(v.styles.TextError.Render("Error: " + v.err.Error()))
		b.WriteString("\n")
	}

	// Selected files list
	if len(v.files) > 0 {
		b.WriteString("\n")
		b.WriteString(v.styles.TextBold.Render(fmt.Sprintf("Selected Files (%d):", len(v.files))))
		b.WriteString("\n")

		for i, file := range v.files {
			info, _ := v.api.GetImageInfo(file)

			prefix := "  "
			style := v.styles.ListItem
			if i == v.cursor {
				prefix = "▸ "
				style = v.styles.ListItemSelected
			}

			fileName := filepath.Base(file)
			sizeStr := ""
			if info != nil {
				sizeStr = fmt.Sprintf(" (%s)", compressor.FormatBytes(info.Size))
			}

			b.WriteString(style.Render(prefix + fileName + v.styles.TextMuted.Render(sizeStr)))
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString("\n")
	help := []string{
		v.styles.HelpKey.Render("tab") + v.styles.HelpDesc.Render(" toggle mode"),
		v.styles.HelpKey.Render("enter") + v.styles.HelpDesc.Render(" select"),
		v.styles.HelpKey.Render("backspace") + v.styles.HelpDesc.Render(" remove"),
		v.styles.HelpKey.Render("→") + v.styles.HelpDesc.Render(" continue"),
	}
	b.WriteString(strings.Join(help, "  "))

	return v.styles.Content.Render(b.String())
}

// GetSelectedFiles returns the list of selected files
func (v HomeView) GetSelectedFiles() []string {
	return v.files
}

// HasSelection returns true if at least one file is selected
func (v HomeView) HasSelection() bool {
	return len(v.files) > 0
}
