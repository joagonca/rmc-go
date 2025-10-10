package export

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ctw00272/rmc-go/internal/rmscene"
)

// ExportToPDF exports a scene tree to PDF format via SVG conversion
func ExportToPDF(tree *rmscene.SceneTree, w io.Writer) error {
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
	defer os.Remove(svgFile.Name())
	defer svgFile.Close()

	if _, err := svgFile.Write(svgBuf.Bytes()); err != nil {
		return fmt.Errorf("failed to write SVG: %w", err)
	}
	svgFile.Close()

	pdfFile, err := os.CreateTemp("", "rmc-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp PDF file: %w", err)
	}
	defer os.Remove(pdfFile.Name())
	pdfName := pdfFile.Name()
	pdfFile.Close()

	// Try to convert with inkscape
	cmd := exec.Command("inkscape", svgFile.Name(), "--export-filename", pdfName)
	if err := cmd.Run(); err != nil {
		// Try macOS path
		cmd = exec.Command("/Applications/Inkscape.app/Contents/MacOS/inkscape",
			svgFile.Name(), "--export-filename", pdfName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("inkscape not found or conversion failed: %w", err)
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
