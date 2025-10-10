# rmc-go

A Go implementation for converting reMarkable tablet v6 format files (`.rm`) to PDF and SVG.

This is a port of the Python [rmc](https://github.com/ricklupton/rmc) tool, which uses [rmscene](https://github.com/ricklupton/rmscene) to read the reMarkable v6 file format.

## Features

- Read reMarkable v6 format files (software version 3+)
- Export to SVG format
- Export to PDF format (requires Inkscape)
- Handles strokes/drawings with different pen types and colors
- Command-line interface

## Installation

### Prerequisites

- Go 1.21 or later
- Inkscape (optional, required for PDF export)

### Build from source

```bash
git clone <repository-url>
cd rmc-go
go build -o rmc-go ./cmd/rmc-go
```

## Usage

### Export to PDF

```bash
./rmc-go file.rm -o output.pdf
```

### Export to SVG

```bash
./rmc-go file.rm -o output.svg
```

### Export to stdout

```bash
./rmc-go file.rm -t svg > output.svg
./rmc-go file.rm -t pdf > output.pdf
```

### Command-line options

```
Usage:
  rmc-go [input.rm] [flags]

Flags:
  -h, --help            help for rmc-go
  -o, --output string   Output file (default: stdout)
  -t, --type string     Output type: svg or pdf (default: guess from filename)
```

## Project Structure

```
rmc-go/
├── cmd/rmc-go/          # CLI application
│   └── main.go
├── internal/
│   ├── rmscene/         # v6 file format parser
│   │   ├── datastream.go      # Binary data stream reader
│   │   ├── block_reader.go    # Tagged block reader
│   │   ├── scene_stream.go    # Scene block parser
│   │   └── types.go           # Data structures
│   └── export/          # Export functionality
│       ├── svg.go             # SVG export
│       ├── pen.go             # Pen rendering
│       └── pdf.go             # PDF export (via SVG)
├── go.mod
└── README.md
```

## How it Works

The reMarkable v6 file format is a binary format that contains:

1. **Header**: File format identifier
2. **Blocks**: Tagged blocks containing different types of data:
   - Scene tree blocks (layers/groups)
   - Scene item blocks (lines, text, etc.)
   - Line/stroke data with points, pressure, speed
   - Text blocks with formatting

This implementation:

1. **Parses** the binary format using a DataStream and TaggedBlockReader
2. **Builds** a scene tree with groups (layers) and items (strokes/text)
3. **Exports** the scene tree to SVG, rendering strokes with appropriate pen styles
4. **Converts** SVG to PDF using Inkscape

## Supported Pen Types

- Ballpoint
- Fineliner
- Marker
- Pencil
- Mechanical Pencil
- Paintbrush/Brush
- Highlighter
- Eraser
- Calligraphy
- Shader

## Limitations

- Text rendering is not yet fully implemented
- Some newer block types may not be supported
- Parser is tolerant of errors and will skip unrecognized blocks

## Acknowledgements

This project is based on:

- [rmc](https://github.com/ricklupton/rmc) - Python converter for reMarkable files
- [rmscene](https://github.com/ricklupton/rmscene) - Python library for reading v6 format
- [ddvk's reader](https://github.com/ddvk/reader) - helped understand the file format

## License

MIT License (same as the original rmc project)

## Development Status

This is a work-in-progress implementation. The core functionality for reading stroke data and exporting to SVG/PDF is working, but some advanced features (text rendering, all block types) are not yet fully implemented.

Contributions and bug reports are welcome!
