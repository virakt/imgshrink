# ImgShrink 🖼️

A powerful Terminal User Interface (TUI) application for image compression written in Go. Reduce image file sizes without losing quality, with support for JPEG and PNG formats.

![ImgShrink Demo](https://via.placeholder.com/800x400?text=ImgShrink+TUI+Demo)

## Features

- 🎨 **Beautiful TUI** - Built with [Charm](https://charm.sh/) libraries (Bubble Tea, Lip Gloss)
- 📁 **Batch Processing** - Compress multiple images at once
- ⚙️ **Flexible Options** - Fine-tune compression settings
- 📊 **Detailed Results** - View compression statistics and savings
- 🔌 **API Layer** - Easy to integrate with GUI applications
- 💻 **CLI Mode** - Use without TUI for scripting

## Supported Formats

- **JPEG** (.jpg, .jpeg)
- **PNG** (.png)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/imgshrink.git
cd imgshrink

# Build
go build -o imgshrink ./cmd/imgshrink

# Install (optional)
go install ./cmd/imgshrink
```

### Using Go Install

```bash
go install github.com/imgshrink/cmd/imgshrink@latest
```

## Usage

### TUI Mode (Interactive)

```bash
# Start the TUI
imgshrink

# Start with files
imgshrink image.jpg photo.png

# Start with glob pattern
imgshrink *.jpg
```

### CLI Mode (Non-interactive)

```bash
# Compress a single file
imgshrink -c image.jpg

# Compress with custom quality (JPEG)
imgshrink -c -q 80 image.jpg

# Compress with custom compression level (PNG)
imgshrink -c -l 9 image.png

# Specify output directory
imgshrink -c -o ./compressed *.jpg

# Batch compress
imgshrink -c -q 75 -o ./output photos/*.jpg
```

### Command Line Options

| Option | Short | Description |
|--------|-------|-------------|
| `--help` | `-h` | Show help message |
| `--version` | `-v` | Show version |
| `--cli` | `-c` | Run in CLI mode (no TUI) |
| `--output` | `-o` | Output directory |
| `--quality` | `-q` | JPEG quality (1-100, default: 85) |
| `--level` | `-l` | PNG compression level (0-9, default: 6) |

## TUI Navigation

### Home View
- `Tab` - Toggle between single/batch mode
- `↑/↓` or `j/k` - Navigate file list
- `d` or `Backspace` - Remove selected file
- `→` or `Enter` - Continue to options
- `q` - Quit

### Options View
- `Tab` or `↑/↓` - Navigate options
- `+/-` - Adjust numeric values
- `p` - Toggle progressive (JPEG)
- `i` - Toggle interlaced (PNG)
- `m` - Toggle strip metadata
- `←` - Go back
- `→` or `Enter` - Start compression

### Progress View
- `→` - View results (when complete)

### Results View
- `↑/↓` or `j/k` - Navigate results
- `r` - Restart (new compression)
- `q` - Quit

## Compression Options

### JPEG Options
- **Quality** (1-100): Higher values = better quality, larger files
- **Progressive**: Enable progressive JPEG encoding
- **Chroma Subsampling**: 4:4:4, 4:2:2, or 4:2:0

### PNG Options
- **Compression Level** (0-9): Higher values = more compression, slower
- **Interlaced**: Enable Adam7 interlacing

### Common Options
- **Resize Percent**: Scale image by percentage
- **Resize Width/Height**: Scale to specific dimensions
- **Strip Metadata**: Remove EXIF and other metadata
- **Output Directory**: Where to save compressed files
- **Output Suffix**: Suffix for output filenames (default: `_compressed`)

## API Usage

ImgShrink provides an API layer for integration with other applications:


## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [imaging](https://github.com/disintegration/imaging) - Image processing

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Charm](https://charm.sh/) for the amazing TUI libraries
- The Go community for excellent image processing packages
