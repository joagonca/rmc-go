package export

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ctw00272/rmc-go/internal/parser"
)

// ExportToPDF exports a scene tree to PDF format via SVG conversion
func ExportToPDF(tree *parser.SceneTree, w io.Writer) error {
	// Create temporary SVG
	svgBuf := &bytes.Buffer{}
	if err := ExportToSVG(tree, svgBuf); err != nil {
		return fmt.Errorf("failed to generate SVG: %w", err)
	}

	// Create temp files
	svgFile, err := os.CreateTemp("", "rmc-*.svg")
	if err != nil {
		return fmt.Errorf("failed to create temp SVG file: %w", err)
	}
	// Ensure cleanup happens in correct order: close before remove
	defer func() {
		svgFile.Close()
		os.Remove(svgFile.Name())
	}()

	if _, err := svgFile.Write(svgBuf.Bytes()); err != nil {
		return fmt.Errorf("failed to write SVG: %w", err)
	}
	svgFile.Close()

	pdfFile, err := os.CreateTemp("", "rmc-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp PDF file: %w", err)
	}
	pdfName := pdfFile.Name()
	pdfFile.Close()
	// Remove PDF temp file after we're done reading it
	defer os.Remove(pdfName)

	// Try to convert with inkscape
	cmd := exec.Command("inkscape", svgFile.Name(), "--export-filename", pdfName)
	var cmdErr error
	if cmdErr = cmd.Run(); cmdErr != nil {
		// Try macOS path
		cmd = exec.Command("/Applications/Inkscape.app/Contents/MacOS/inkscape",
			svgFile.Name(), "--export-filename", pdfName)
		if cmdErr = cmd.Run(); cmdErr != nil {
			return fmt.Errorf("inkscape conversion failed (is Inkscape installed?): %w\n"+
				"  Install Inkscape from: https://inkscape.org/release/\n"+
				"  Or use SVG output with: -t svg", cmdErr)
		}
	}

	// Read and write PDF
	pdfData, err := os.ReadFile(pdfName)
	if err != nil {
		return fmt.Errorf("failed to read PDF: %w", err)
	}

	if _, err := w.Write(pdfData); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}
