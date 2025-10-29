// Package rmc provides high-level convenience functions for converting reMarkable v6 files
// to PDF and SVG formats. This package is designed to be used both as a library and as a CLI tool.
//
// For more control over the conversion process, use the parser and export packages directly.
package rmc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/joagonca/rmc-go/export"
	"github.com/joagonca/rmc-go/parser"
)

// Format represents the output format type
type Format string

const (
	// FormatPDF represents PDF output format
	FormatPDF Format = "pdf"
	// FormatSVG represents SVG output format
	FormatSVG Format = "svg"
)

// Options contains configuration options for conversion
type Options struct {
	// UseLegacy uses the Inkscape-based PDF renderer instead of Cairo (default: false)
	UseLegacy bool
}

// DefaultOptions returns the default conversion options
func DefaultOptions() *Options {
	return &Options{
		UseLegacy: false,
	}
}

// ConvertFile converts a reMarkable .rm file to the specified output format.
// The output format is inferred from the output file extension if not explicitly specified.
//
// Example:
//
//	err := rmc.ConvertFile("input.rm", "output.pdf", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ConvertFile(inputPath, outputPath string, opts *Options) error {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Infer format from output path
	format := inferFormat(outputPath)

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Convert
	return Convert(inputFile, outputFile, format, opts)
}

// Convert converts a reMarkable .rm file from a reader to the specified output format.
// This function is useful when working with in-memory data or custom readers/writers.
//
// Example:
//
//	input := bytes.NewReader(rmData)
//	var output bytes.Buffer
//	err := rmc.Convert(input, &output, rmc.FormatPDF, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	pdfData := output.Bytes()
func Convert(input io.Reader, output io.Writer, format Format, opts *Options) error {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Parse the .rm file
	tree, err := parser.ReadSceneTree(input)
	if err != nil {
		return fmt.Errorf("failed to parse .rm file: %w", err)
	}

	// Export based on format
	switch format {
	case FormatSVG:
		if err := export.ExportToSVG(tree, output); err != nil {
			return fmt.Errorf("failed to export to SVG: %w", err)
		}
	case FormatPDF:
		if err := export.ExportToPDF(tree, output, opts.UseLegacy); err != nil {
			return fmt.Errorf("failed to export to PDF: %w", err)
		}
	default:
		return fmt.Errorf("unknown format: %s (supported: pdf, svg)", format)
	}

	return nil
}

// ConvertToBytes converts a reMarkable .rm file from binary data to the specified output format,
// returning the result as a byte slice.
//
// Example:
//
//	rmData, _ := os.ReadFile("input.rm")
//	pdfData, err := rmc.ConvertToBytes(rmData, rmc.FormatPDF, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.pdf", pdfData, 0644)
func ConvertToBytes(data []byte, format Format, opts *Options) ([]byte, error) {
	input := bytes.NewReader(data)
	output := &bytes.Buffer{}

	if err := Convert(input, output, format, opts); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

// ConvertFromBytes is an alias for ConvertToBytes for clarity
func ConvertFromBytes(data []byte, format Format, opts *Options) ([]byte, error) {
	return ConvertToBytes(data, format, opts)
}

// ConvertFileToBytes reads a reMarkable .rm file and converts it to the specified format,
// returning the result as a byte slice.
//
// Example:
//
//	pdfData, err := rmc.ConvertFileToBytes("input.rm", rmc.FormatPDF, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ConvertFileToBytes(inputPath string, format Format, opts *Options) ([]byte, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	return ConvertToBytes(data, format, opts)
}

// ConvertBytesToFile converts a reMarkable .rm file from binary data to the specified format
// and writes it to a file.
//
// Example:
//
//	rmData, _ := os.ReadFile("input.rm")
//	err := rmc.ConvertBytesToFile(rmData, "output.pdf", rmc.FormatPDF, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ConvertBytesToFile(data []byte, outputPath string, format Format, opts *Options) error {
	outputData, err := ConvertToBytes(data, format, opts)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// inferFormat infers the output format from a file path based on extension
func inferFormat(path string) Format {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".svg":
		return FormatSVG
	case ".pdf":
		return FormatPDF
	default:
		return FormatPDF // default to PDF
	}
}
