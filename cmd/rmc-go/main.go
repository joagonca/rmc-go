package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joagonca/rmc-go/internal/export"
	"github.com/joagonca/rmc-go/internal/parser"
	"github.com/spf13/cobra"
)

var (
	outputFile  string
	outputType  string
	useNative   bool
	contentFile string
)

var rootCmd = &cobra.Command{
	Use:   "rmc-go [input.rm|folder]",
	Short: "Convert reMarkable v6 files to PDF/SVG",
	Long: `rmc-go is a tool to convert reMarkable tablet v6 format files to PDF or SVG.

Example usage:
  rmc-go file.rm -o output.pdf
  rmc-go file.rm -o output.svg
  rmc-go file.rm -t pdf > output.pdf
  rmc-go folder/ -o output.pdf  # Multipage PDF from all .rm files in folder
  rmc-go folder/ -o output.pdf --content folder.content  # Use .content file for page ordering`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().StringVarP(&outputType, "type", "t", "", "Output type: svg or pdf (default: guess from filename)")
	rootCmd.Flags().BoolVar(&useNative, "native", false, "Use native Cairo renderer for PDF export (requires CGo)")
	rootCmd.Flags().StringVar(&contentFile, "content", "", "Path to .content file for page ordering (only used with folders)")
}

func run(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	// Check if input is a file or directory
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to access input path: %w", err)
	}

	// Determine output type
	format := outputType
	if format == "" {
		if outputFile != "" {
			format = guessFormat(outputFile)
		} else {
			format = "pdf" // default
		}
	}

	// Handle directory input
	if info.IsDir() {
		return handleDirectory(inputPath, format)
	}

	// Handle single file input
	return handleSingleFile(inputPath, format)
}

func handleSingleFile(inputFile string, format string) error {
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

func handleDirectory(inputDir string, format string) error {
	// Validate that SVG output is not requested for folders
	if strings.ToLower(format) == "svg" {
		return fmt.Errorf("multipage output is only supported for PDF format, not SVG")
	}

	// Collect all .rm files from the directory
	files, err := collectRmFiles(inputDir)
	if err != nil {
		return fmt.Errorf("failed to collect .rm files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .rm files found in directory: %s", inputDir)
	}

	// Try to order files using .content file if specified
	usedContentFile := false
	if contentFile != "" {
		var orderedFiles []string
		orderedFiles, usedContentFile = parser.OrderFilesByContent(files, contentFile)
		if usedContentFile {
			files = orderedFiles
			fmt.Fprintf(os.Stderr, "Using page ordering from content file: %s\n", contentFile)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Could not use content file %s, falling back to modification time ordering\n", contentFile)
		}
	}

	// If no content file was used, sort by modification time (oldest first)
	if !usedContentFile {
		sort.Slice(files, func(i, j int) bool {
			infoI, _ := os.Stat(files[i])
			infoJ, _ := os.Stat(files[j])
			return infoI.ModTime().Before(infoJ.ModTime())
		})
		if contentFile == "" {
			fmt.Fprintf(os.Stderr, "Warning: Using modification time for page ordering. For reliable ordering, use --content flag.\n")
		}
	}

	// Parse all .rm files into scene trees
	var trees []*parser.SceneTree
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", file, err)
		}
		tree, err := parser.ReadSceneTree(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", file, err)
		}
		trees = append(trees, tree)
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

	// Export multipage PDF
	if err := export.ExportToMultipagePDF(trees, out, useNative); err != nil {
		return fmt.Errorf("failed to export multipage PDF: %w", err)
	}

	return nil
}

func collectRmFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".rm") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
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
