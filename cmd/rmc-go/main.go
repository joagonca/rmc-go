package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joagonca/rmc-go/internal/export"
	"github.com/joagonca/rmc-go/internal/parser"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	outputType string
	useNative  bool
)

var rootCmd = &cobra.Command{
	Use:   "rmc-go [input.rm]",
	Short: "Convert reMarkable v6 files to PDF/SVG",
	Long: `rmc-go is a tool to convert reMarkable tablet v6 format files to PDF or SVG.

Example usage:
  rmc-go file.rm -o output.pdf
  rmc-go file.rm -o output.svg
  rmc-go file.rm -t pdf > output.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().StringVarP(&outputType, "type", "t", "", "Output type: svg or pdf (default: guess from filename)")
	rootCmd.Flags().BoolVar(&useNative, "native", false, "Use native Cairo renderer for PDF export (requires CGo)")
}

func run(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Determine output type
	format := outputType
	if format == "" {
		if outputFile != "" {
			format = guessFormat(outputFile)
		} else {
			format = "pdf" // default
		}
	}

	// Open input file
	f, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer f.Close()

	// Parse the .rm file
	tree, err := parser.ReadSceneTree(f)
	if err != nil {
		return fmt.Errorf("failed to parse .rm file: %w", err)
	}

	// Determine output writer
	var out *os.File
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	// Export
	switch strings.ToLower(format) {
	case "svg":
		if err := export.ExportToSVG(tree, out); err != nil {
			return fmt.Errorf("failed to export to SVG: %w", err)
		}
	case "pdf":
		if err := export.ExportToPDF(tree, out, useNative); err != nil {
			return fmt.Errorf("failed to export to PDF: %w", err)
		}
	default:
		return fmt.Errorf("unknown format: %s (supported: svg, pdf)", format)
	}

	return nil
}

func guessFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".svg":
		return "svg"
	case ".pdf":
		return "pdf"
	default:
		return "pdf"
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
