// +build ignore

// This is an example of how to use rmc-go as a library in your own Go application.
// To run this example: go run example_library_usage.go

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/joagonca/rmc-go"
)

func main() {
	fmt.Println("rmc-go Library Usage Examples")
	fmt.Println("==============================\n")

	// Example 1: Simple file-to-file conversion
	example1_FileToFile()

	// Example 2: Convert with options (using legacy Inkscape renderer)
	example2_WithOptions()

	// Example 3: Convert file to bytes (in-memory)
	example3_FileToBytes()

	// Example 4: Convert from binary data to binary data
	example4_BytesToBytes()

	// Example 5: Using the lower-level API for more control
	example5_LowLevelAPI()
}

// Example 1: Simple file-to-file conversion
func example1_FileToFile() {
	fmt.Println("Example 1: File to File Conversion")
	fmt.Println("-----------------------------------")

	// Convert .rm file to PDF
	err := rmc.ConvertFile("input.rm", "output.pdf", nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Converted input.rm to output.pdf")
	}

	// Convert .rm file to SVG
	err = rmc.ConvertFile("input.rm", "output.svg", nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Converted input.rm to output.svg")
	}
	fmt.Println()
}

// Example 2: Convert with options (using legacy Inkscape renderer)
func example2_WithOptions() {
	fmt.Println("Example 2: With Custom Options")
	fmt.Println("-------------------------------")

	opts := &rmc.Options{
		UseLegacy: true, // Use Inkscape renderer instead of Cairo
	}

	err := rmc.ConvertFile("input.rm", "output_legacy.pdf", opts)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Converted using legacy Inkscape renderer")
	}
	fmt.Println()
}

// Example 3: Convert file to bytes (in-memory)
func example3_FileToBytes() {
	fmt.Println("Example 3: File to Bytes (In-Memory)")
	fmt.Println("-------------------------------------")

	// Read .rm file and convert to PDF bytes
	pdfData, err := rmc.ConvertFileToBytes("input.rm", rmc.FormatPDF, nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("✓ Converted to PDF in memory (%d bytes)\n", len(pdfData))

	// You can now do whatever you want with the PDF data:
	// - Send it over HTTP
	// - Store it in a database
	// - Write it to a file
	// - etc.

	// For example, write it to a file:
	err = os.WriteFile("output_from_bytes.pdf", pdfData, 0644)
	if err != nil {
		log.Printf("Error writing file: %v\n", err)
	} else {
		fmt.Println("✓ Wrote PDF data to output_from_bytes.pdf")
	}
	fmt.Println()
}

// Example 4: Convert from binary data to binary data
func example4_BytesToBytes() {
	fmt.Println("Example 4: Bytes to Bytes (Full In-Memory)")
	fmt.Println("-------------------------------------------")

	// Read .rm file into memory
	rmData, err := os.ReadFile("input.rm")
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("Read .rm file (%d bytes)\n", len(rmData))

	// Convert to PDF (all in memory)
	pdfData, err := rmc.ConvertFromBytes(rmData, rmc.FormatPDF, nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("✓ Converted to PDF (%d bytes)\n", len(pdfData))

	// Convert to SVG (all in memory)
	svgData, err := rmc.ConvertFromBytes(rmData, rmc.FormatSVG, nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("✓ Converted to SVG (%d bytes)\n", len(svgData))
	fmt.Println()
}

// Example 5: Using the lower-level API for more control
func example5_LowLevelAPI() {
	fmt.Println("Example 5: Low-Level API (Advanced)")
	fmt.Println("------------------------------------")

	// You can also use the parser and export packages directly
	// for more control over the conversion process

	// This example shows how to use readers and writers directly
	rmData := []byte("...") // Your .rm file data

	input := bytes.NewReader(rmData)
	output := &bytes.Buffer{}

	err := rmc.Convert(input, output, rmc.FormatPDF, rmc.DefaultOptions())
	if err != nil {
		log.Printf("Error: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("✓ Converted using low-level API (%d bytes)\n", output.Len())
	fmt.Println()
}
