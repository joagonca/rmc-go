// +build cairo

package export

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/joagonca/rmc-go/internal/parser"
	"github.com/ungerik/go-cairo"
)

// ExportToPDFCairo exports a scene tree directly to PDF using Cairo
func ExportToPDFCairo(tree *parser.SceneTree, w io.Writer) error {
	if tree == nil {
		return fmt.Errorf("scene tree cannot be nil")
	}
	if tree.Root == nil {
		return fmt.Errorf("scene tree root cannot be nil")
	}

	// Build anchor positions (including text-based anchors)
	anchorPos := buildAnchorPos(tree.RootText)

	// Calculate bounding box using the anchor positions
	xMin, xMax, yMin, yMax := getBoundingBox(tree.Root, anchorPos)

	// Include text area in bounding box calculation
	if tree.RootText != nil {
		textMinX := tree.RootText.PosX
		textMaxX := tree.RootText.PosX + float64(tree.RootText.Width)

		// Calculate text Y range by going through all paragraphs
		doc, err := parser.BuildTextDocument(tree.RootText)
		if err == nil && len(doc.Paragraphs) > 0 {
			yOffset := TextTopY
			textMinY := math.MaxFloat64
			textMaxY := -math.MaxFloat64

			for _, p := range doc.Paragraphs {
				lineHeight := lineHeights[p.Style]
				if lineHeight == 0 {
					lineHeight = 70
				}
				yOffset += lineHeight
				yPos := tree.RootText.PosY + yOffset

				textMinY = math.Min(textMinY, yPos)
				textMaxY = math.Max(textMaxY, yPos)
			}

			xMin = math.Min(xMin, textMinX)
			xMax = math.Max(xMax, textMaxX)
			yMin = math.Min(yMin, textMinY)
			yMax = math.Max(yMax, textMaxY)
		}
	}

	width := scale(xMax - xMin + 1)
	height := scale(yMax - yMin + 1)

	// Create a temporary file for PDF output
	// Cairo requires a file path, so we write to temp and then copy
	tmpFile, err := os.CreateTemp("", "rmc-cairo-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Create a Cairo PDF surface with the temp file
	pdfSurface := cairo.NewPDFSurface(tmpPath, width, height, cairo.PDF_VERSION_1_5)
	defer pdfSurface.Finish()

	// Set up coordinate system - translate to account for bounding box offset
	pdfSurface.Translate(-scale(xMin), -scale(yMin))

	// Render the content
	// Draw text first (if it exists)
	if tree.RootText != nil {
		if err := drawTextCairo(tree.RootText, pdfSurface); err != nil {
			return fmt.Errorf("failed to draw root text: %w", err)
		}
	}

	// Draw strokes/groups
	if err := drawGroupCairo(tree.Root, pdfSurface, anchorPos); err != nil {
		return fmt.Errorf("failed to draw group: %w", err)
	}

	// Finish the surface to flush all drawing operations
	pdfSurface.Finish()

	// Read the temporary PDF file and write to the output
	pdfData, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read generated PDF: %w", err)
	}

	if _, err := w.Write(pdfData); err != nil {
		return fmt.Errorf("failed to write PDF output: %w", err)
	}

	return nil
}

func drawGroupCairo(group *parser.Group, surface *cairo.Surface, anchorPos map[parser.CrdtID]float64) error {
	surface.Save()

	anchorX, anchorY := getAnchor(group, anchorPos)
	surface.Translate(scale(anchorX), scale(anchorY))

	if group.Children != nil {
		for _, item := range group.Children.Items {
			if item.Value == nil {
				continue
			}

			switch v := item.Value.(type) {
			case *parser.Group:
				if err := drawGroupCairo(v, surface, anchorPos); err != nil {
					return err
				}
			case *parser.Line:
				drawStrokeCairo(v, surface)
			case *parser.Text:
				if err := drawTextCairo(v, surface); err != nil {
					return err
				}
			}
		}
	}

	surface.Restore()
	return nil
}

func drawStrokeCairo(line *parser.Line, surface *cairo.Surface) {
	pen := createPen(line.Tool, line.Color, line.ColorOverride, line.ThicknessScale)

	lastSegmentWidth := 0.0

	for i, point := range line.Points {
		xPos := float64(point.X)
		yPos := float64(point.Y)

		if i%pen.segmentLength == 0 {
			// Start new segment with updated properties
			segmentColor := pen.getSegmentColorRGB(point, lastSegmentWidth)
			segmentWidth := pen.getSegmentWidth(point, lastSegmentWidth)
			segmentOpacity := pen.getSegmentOpacity(point, lastSegmentWidth)

			// Set color with opacity
			surface.SetSourceRGBA(
				float64(segmentColor.R)/255.0,
				float64(segmentColor.G)/255.0,
				float64(segmentColor.B)/255.0,
				segmentOpacity,
			)

			// Set line width
			surface.SetLineWidth(scale(segmentWidth))

			// Set line cap
			if pen.strokeLinecap == "round" {
				surface.SetLineCap(cairo.LINE_CAP_ROUND)
			} else if pen.strokeLinecap == "square" {
				surface.SetLineCap(cairo.LINE_CAP_SQUARE)
			} else {
				surface.SetLineCap(cairo.LINE_CAP_BUTT)
			}

			surface.SetLineJoin(cairo.LINE_JOIN_ROUND)

			// Start new path
			if i == 0 {
				surface.MoveTo(scale(xPos), scale(yPos))
			}

			lastSegmentWidth = segmentWidth
		}

		if i > 0 {
			surface.LineTo(scale(xPos), scale(yPos))
		}

		// Stroke at segment boundaries
		if i > 0 && (i+1)%pen.segmentLength == 0 {
			surface.Stroke()
			// Move to current position to continue
			surface.MoveTo(scale(xPos), scale(yPos))
		}
	}

	// Stroke any remaining path
	surface.Stroke()
}

func drawTextCairo(text *parser.Text, surface *cairo.Surface) error {
	// Convert text to TextDocument
	doc, err := parser.BuildTextDocument(text)
	if err != nil {
		return fmt.Errorf("failed to build text document: %w", err)
	}

	// Iterate through paragraphs
	yOffset := TextTopY
	bulletNumber := 1
	for _, p := range doc.Paragraphs {
		// Get line height for this style
		lineHeight := lineHeights[p.Style]
		if lineHeight == 0 {
			lineHeight = 70
		}
		yOffset += lineHeight

		// Calculate position
		xPos := text.PosX
		yPos := text.PosY + yOffset

		// Skip empty lines
		trimmedText := p.Text
		if trimmedText == "" {
			continue
		}

		// Add appropriate prefix based on style
		prefix := getParagraphPrefix(p.Style, &bulletNumber)
		displayText := prefix + trimmedText

		// Set font based on style
		setTextFontCairo(surface, p.Style)

		// Set text color (black)
		surface.SetSourceRGB(0, 0, 0)

		// Draw text
		surface.MoveTo(scale(xPos), scale(yPos))
		surface.ShowText(displayText)
	}

	return nil
}

func setTextFontCairo(surface *cairo.Surface, style parser.ParagraphStyle) {
	switch style {
	case parser.StyleHeading:
		surface.SelectFontFace("serif", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
		surface.SetFontSize(14.0)
	case parser.StyleBold:
		surface.SelectFontFace("sans-serif", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_BOLD)
		surface.SetFontSize(8.0)
	default:
		surface.SelectFontFace("sans-serif", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
		surface.SetFontSize(7.0)
	}
}

// Helper method to get RGB color for Cairo (instead of CSS string)
func (p *pen) getSegmentColorRGB(point parser.Point, lastWidth float64) RGB {
	switch p.name {
	case "Ballpoint":
		speed := float64(point.Speed) / 4.0
		pressure := float64(point.Pressure) / 255.0
		intensity := (0.1 * -(speed / 35.0)) + (1.2 * pressure) + 0.5
		intensity = clamp(intensity)
		factor := math.Min(math.Abs(intensity-1), 0.235)
		r := int(float64(p.baseColor.R) * (1 - factor))
		g := int(float64(p.baseColor.G) * (1 - factor))
		b := int(float64(p.baseColor.B) * (1 - factor))
		return RGB{R: r, G: g, B: b}

	case "Brush":
		speed := float64(point.Speed) / 4.0
		pressure := float64(point.Pressure) / 255.0
		intensity := math.Pow(pressure, 1.5) - 0.2*(speed/50.0)
		intensity = clamp(intensity)
		r := int(float64(p.baseColor.R) * intensity)
		g := int(float64(p.baseColor.G) * intensity)
		b := int(float64(p.baseColor.B) * intensity)
		return RGB{R: r, G: g, B: b}

	default:
		return p.baseColor
	}
}
