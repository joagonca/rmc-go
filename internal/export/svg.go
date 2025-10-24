package export

import (
	"fmt"
	"io"
	"math"

	"github.com/ctw00272/rmc-go/internal/rmscene"
)

const (
	ScreenWidth  = 1404
	ScreenHeight = 1872
	ScreenDPI    = 226
	Scale        = 72.0 / ScreenDPI
	TextTopY     = -88.0 // Base Y offset for text
)

var lineHeights = map[rmscene.ParagraphStyle]float64{
	rmscene.StylePlain:           70,
	rmscene.StyleBullet:          35,
	rmscene.StyleBullet2:         35,
	rmscene.StyleBold:            70,
	rmscene.StyleHeading:         150,
	rmscene.StyleCheckbox:        35,
	rmscene.StyleCheckboxChecked: 35,
}

// ExportToSVG exports a scene tree to SVG format
func ExportToSVG(tree *rmscene.SceneTree, w io.Writer) error {
	// Calculate bounding box
	xMin, xMax, yMin, yMax := getBoundingBox(tree.Root, make(map[rmscene.CrdtID]float64))

	width := scale(xMax - xMin + 1)
	height := scale(yMax - yMin + 1)

	// Write SVG header
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" height="%.1f" width="%.1f" viewBox="%.1f %.1f %.1f %.1f">
`, height, width, scale(xMin), scale(yMin), width, height)

	fmt.Fprintf(w, "\t<g id=\"p1\" style=\"display:inline\">\n")

	// Render RootText if it exists
	if tree.RootText != nil {
		drawText(tree.RootText, w, "\t\t")
	}

	// Draw content
	anchorPos := buildAnchorPos(tree.RootText)
	drawGroup(tree.Root, w, anchorPos, "\t\t")

	// Close
	fmt.Fprintf(w, "\t</g>\n")
	fmt.Fprintf(w, "</svg>\n")

	return nil
}

func scale(v float64) float64 {
	return v * Scale
}

func buildAnchorPos(text *rmscene.Text) map[rmscene.CrdtID]float64 {
	anchorPos := make(map[rmscene.CrdtID]float64)

	// Special anchors
	anchorPos[rmscene.CrdtID{Part1: 0, Part2: 281474976710654}] = 100
	anchorPos[rmscene.CrdtID{Part1: 0, Part2: 281474976710655}] = 100

	// TODO: Add text-based anchors when text is parsed

	return anchorPos
}

func getBoundingBox(group *rmscene.Group, anchorPos map[rmscene.CrdtID]float64) (float64, float64, float64, float64) {
	xMin := -float64(ScreenWidth) / 2
	xMax := float64(ScreenWidth) / 2
	yMin := 0.0
	yMax := float64(ScreenHeight)

	if group.Children == nil {
		return xMin, xMax, yMin, yMax
	}

	for _, item := range group.Children.Items {
		if item.Value == nil {
			continue
		}

		switch v := item.Value.(type) {
		case *rmscene.Group:
			anchorX, anchorY := getAnchor(v, anchorPos)
			xMinT, xMaxT, yMinT, yMaxT := getBoundingBox(v, anchorPos)
			xMin = math.Min(xMin, xMinT+anchorX)
			xMax = math.Max(xMax, xMaxT+anchorX)
			yMin = math.Min(yMin, yMinT+anchorY)
			yMax = math.Max(yMax, yMaxT+anchorY)

		case *rmscene.Line:
			for _, p := range v.Points {
				xMin = math.Min(xMin, float64(p.X))
				xMax = math.Max(xMax, float64(p.X))
				yMin = math.Min(yMin, float64(p.Y))
				yMax = math.Max(yMax, float64(p.Y))
			}
		}
	}

	return xMin, xMax, yMin, yMax
}

func getAnchor(group *rmscene.Group, anchorPos map[rmscene.CrdtID]float64) (float64, float64) {
	anchorX := 0.0
	anchorY := 0.0

	if group.AnchorID != nil && group.AnchorOriginX != nil {
		anchorX = float64(group.AnchorOriginX.Value)
		if y, ok := anchorPos[group.AnchorID.Value]; ok {
			anchorY = y
		}
	}

	return anchorX, anchorY
}

func drawGroup(group *rmscene.Group, w io.Writer, anchorPos map[rmscene.CrdtID]float64, indent string) {
	anchorX, anchorY := getAnchor(group, anchorPos)
	fmt.Fprintf(w, "%s<g id=\"%s\" transform=\"translate(%.3f, %.3f)\">\n",
		indent, group.NodeID, scale(anchorX), scale(anchorY))

	if group.Children != nil {
		for _, item := range group.Children.Items {
			if item.Value == nil {
				continue
			}

			switch v := item.Value.(type) {
			case *rmscene.Group:
				drawGroup(v, w, anchorPos, indent+"\t")
			case *rmscene.Line:
				drawStroke(v, w, indent+"\t")
			case *rmscene.Text:
				drawText(v, w, indent+"\t")
			}
		}
	}

	fmt.Fprintf(w, "%s</g>\n", indent)
}

func drawStroke(line *rmscene.Line, w io.Writer, indent string) {
	pen := createPen(line.Tool, line.Color, line.ThicknessScale)

	lastXPos := -1.0
	lastYPos := -1.0
	lastSegmentWidth := 0.0

	for i, point := range line.Points {
		xPos := float64(point.X)
		yPos := float64(point.Y)

		if i%pen.segmentLength == 0 {
			// End previous segment
			if lastXPos != -1.0 {
				fmt.Fprintf(w, "\"/>\n")
			}

			segmentColor := pen.getSegmentColor(point, lastSegmentWidth)
			segmentWidth := pen.getSegmentWidth(point, lastSegmentWidth)
			segmentOpacity := pen.getSegmentOpacity(point, lastSegmentWidth)

			fmt.Fprintf(w, "%s<polyline ", indent)
			fmt.Fprintf(w, "style=\"fill:none; stroke:%s; stroke-width:%.3f; opacity:%.3f\" ",
				segmentColor, scale(segmentWidth), segmentOpacity)
			fmt.Fprintf(w, "stroke-linecap=\"%s\" ", pen.strokeLinecap)
			fmt.Fprintf(w, "points=\"")

			if lastXPos != -1.0 {
				fmt.Fprintf(w, "%.3f,%.3f ", scale(lastXPos), scale(lastYPos))
			}

			lastSegmentWidth = segmentWidth
		}

		lastXPos = xPos
		lastYPos = yPos

		fmt.Fprintf(w, "%.3f,%.3f ", scale(xPos), scale(yPos))
	}

	fmt.Fprintf(w, "\" />\n")
}

func drawText(text *rmscene.Text, w io.Writer, indent string) {
	// Convert text to TextDocument
	doc, err := rmscene.BuildTextDocument(text)
	if err != nil {
		return
	}

	// Write opening group tag
	fmt.Fprintf(w, "%s<g class=\"root-text\" style=\"display:inline\">\n", indent)

	// Write CSS style block
	writeTextStyles(w, indent+"\t")

	// Iterate through paragraphs
	yOffset := TextTopY
	for _, p := range doc.Paragraphs {
		// Get line height for this style
		lineHeight := lineHeights[p.Style]
		if lineHeight == 0 {
			lineHeight = 70 // default
		}
		yOffset += lineHeight

		// Calculate position
		xPos := text.PosX
		yPos := text.PosY + yOffset

		// Get CSS class name
		className := getStyleClassName(p.Style)

		// Write text element (skip empty lines as they just add spacing)
		trimmedText := p.Text // Don't trim - preserve spacing
		if trimmedText != "" {
			fmt.Fprintf(w, "%s<text x=\"%.3f\" y=\"%.3f\" class=\"%s\">%s</text>\n",
				indent+"\t", scale(xPos), scale(yPos), className, htmlEscape(trimmedText))
		}
	}

	// Close group
	fmt.Fprintf(w, "%s</g>\n", indent)
}

func writeTextStyles(w io.Writer, indent string) {
	fmt.Fprintf(w, "%s<style>\n", indent)
	fmt.Fprintf(w, "%s\ttext.heading { font: 14pt serif; }\n", indent)
	fmt.Fprintf(w, "%s\ttext.bold { font: 8pt sans-serif; font-weight: bold; }\n", indent)
	fmt.Fprintf(w, "%s\ttext, text.plain { font: 7pt sans-serif; }\n", indent)
	fmt.Fprintf(w, "%s\ttext.bullet { font: 7pt sans-serif; }\n", indent)
	fmt.Fprintf(w, "%s\ttext.bullet2 { font: 7pt sans-serif; }\n", indent)
	fmt.Fprintf(w, "%s\ttext.checkbox { font: 7pt sans-serif; }\n", indent)
	fmt.Fprintf(w, "%s\ttext.checkbox-checked { font: 7pt sans-serif; }\n", indent)
	fmt.Fprintf(w, "%s</style>\n", indent)
}

func getStyleClassName(style rmscene.ParagraphStyle) string {
	switch style {
	case rmscene.StyleHeading:
		return "heading"
	case rmscene.StyleBold:
		return "bold"
	case rmscene.StylePlain:
		return "plain"
	case rmscene.StyleBullet:
		return "bullet"
	case rmscene.StyleBullet2:
		return "bullet2"
	case rmscene.StyleCheckbox:
		return "checkbox"
	case rmscene.StyleCheckboxChecked:
		return "checkbox-checked"
	default:
		return "plain"
	}
}

func htmlEscape(s string) string {
	// Escape HTML special characters
	result := ""
	for _, ch := range s {
		switch ch {
		case '&':
			result += "&amp;"
		case '<':
			result += "&lt;"
		case '>':
			result += "&gt;"
		case '"':
			result += "&quot;"
		case '\'':
			result += "&#39;"
		default:
			result += string(ch)
		}
	}
	return result
}
