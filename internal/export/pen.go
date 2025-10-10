package export

import (
	"fmt"
	"math"

	"github.com/ctw00272/rmc-go/internal/rmscene"
)

type RGB struct {
	R, G, B int
}

var rmPalette = map[rmscene.PenColor]RGB{
	rmscene.ColorBlack:       {0, 0, 0},
	rmscene.ColorGray:        {144, 144, 144},
	rmscene.ColorWhite:       {255, 255, 255},
	rmscene.ColorYellow:      {251, 247, 25},
	rmscene.ColorGreen:       {0, 255, 0},
	rmscene.ColorPink:        {255, 192, 203},
	rmscene.ColorBlue:        {78, 105, 201},
	rmscene.ColorRed:         {179, 62, 57},
	rmscene.ColorGrayOverlap: {125, 125, 125},
	rmscene.ColorGreen2:      {161, 216, 125},
	rmscene.ColorCyan:        {139, 208, 229},
	rmscene.ColorMagenta:     {183, 130, 205},
	rmscene.ColorYellow2:     {247, 232, 81},
}

type pen struct {
	name           string
	baseWidth      float64
	baseColor      RGB
	segmentLength  int
	baseOpacity    float64
	strokeLinecap  string
	strokeOpacity  float64
	thicknessScale float64
}

func createPen(penType rmscene.Pen, color rmscene.PenColor, thicknessScale float64) *pen {
	baseColor, ok := rmPalette[color]
	if !ok {
		baseColor = RGB{0, 0, 0}
	}

	p := &pen{
		baseColor:      baseColor,
		segmentLength:  1000,
		baseOpacity:    1.0,
		strokeLinecap:  "round",
		strokeOpacity:  1.0,
		thicknessScale: thicknessScale,
	}

	switch penType {
	case rmscene.PenBallpoint1, rmscene.PenBallpoint2:
		p.name = "Ballpoint"
		p.baseWidth = thicknessScale
		p.segmentLength = 5
	case rmscene.PenFineliner1, rmscene.PenFineliner2:
		p.name = "Fineliner"
		p.baseWidth = thicknessScale * 1.8
	case rmscene.PenMarker1, rmscene.PenMarker2:
		p.name = "Marker"
		p.baseWidth = thicknessScale
		p.segmentLength = 3
	case rmscene.PenPencil1, rmscene.PenPencil2:
		p.name = "Pencil"
		p.baseWidth = thicknessScale
		p.segmentLength = 2
	case rmscene.PenMechanicalPencil1, rmscene.PenMechanicalPencil2:
		p.name = "MechanicalPencil"
		p.baseWidth = thicknessScale * thicknessScale
		p.baseOpacity = 0.7
	case rmscene.PenPaintbrush1, rmscene.PenPaintbrush2:
		p.name = "Brush"
		p.baseWidth = thicknessScale
		p.segmentLength = 2
		p.strokeLinecap = "round"
	case rmscene.PenHighlighter1, rmscene.PenHighlighter2:
		p.name = "Highlighter"
		p.baseWidth = 15
		p.strokeLinecap = "square"
		p.baseOpacity = 0.3
		p.strokeOpacity = 0.2
	case rmscene.PenEraser:
		p.name = "Eraser"
		p.baseWidth = thicknessScale * 2
		p.strokeLinecap = "square"
		p.baseColor = rmPalette[rmscene.ColorWhite]
	case rmscene.PenEraserArea:
		p.name = "EraseArea"
		p.baseWidth = thicknessScale
		p.strokeLinecap = "square"
		p.baseOpacity = 0
	case rmscene.PenCalligraphy:
		p.name = "Calligraphy"
		p.baseWidth = thicknessScale
		p.segmentLength = 2
	case rmscene.PenShader:
		p.name = "Shader"
		p.baseWidth = 12
		p.strokeLinecap = "round"
		p.baseOpacity = 0.1
	default:
		p.name = "Unknown"
		p.baseWidth = thicknessScale
	}

	return p
}

func (p *pen) getSegmentColor(point rmscene.Point, lastWidth float64) string {
	switch p.name {
	case "Ballpoint":
		speed := float64(point.Speed) / 4.0
		pressure := float64(point.Pressure) / 255.0
		intensity := (0.1 * -(speed / 35.0)) + (1.2 * pressure) + 0.5
		intensity = clamp(intensity)
		gray := int(math.Min(math.Abs(intensity-1)*255, 60))
		return fmt.Sprintf("rgb(%d,%d,%d)", gray, gray, gray)

	case "Brush":
		speed := float64(point.Speed) / 4.0
		pressure := float64(point.Pressure) / 255.0
		intensity := math.Pow(pressure, 1.5) - 0.2*(speed/50.0)
		intensity = 1.5 * clamp(intensity)
		revIntensity := math.Abs(intensity - 1)
		r := int(revIntensity * float64(255-p.baseColor.R))
		g := int(revIntensity * float64(255-p.baseColor.G))
		b := int(revIntensity * float64(255-p.baseColor.B))
		return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)

	default:
		return fmt.Sprintf("rgb(%d,%d,%d)", p.baseColor.R, p.baseColor.G, p.baseColor.B)
	}
}

func (p *pen) getSegmentWidth(point rmscene.Point, lastWidth float64) float64 {
	speed := float64(point.Speed) / 4.0
	pressure := float64(point.Pressure) / 255.0
	width := float64(point.Width) / 4.0
	tilt := directionToTilt(point.Direction)

	switch p.name {
	case "Ballpoint":
		return (0.5 + pressure) + width - 0.5*(speed/50.0)

	case "Marker":
		return 0.9*(width-0.4*tilt) + (0.1 * lastWidth)

	case "Pencil":
		segWidth := 0.7 * ((((0.8 * p.baseWidth) + (0.5 * pressure)) * width) -
			(0.25 * math.Pow(tilt, 1.8)) - (0.6 * (speed / 50.0)))
		maxWidth := p.baseWidth * 10
		if segWidth > maxWidth {
			return maxWidth
		}
		return segWidth

	case "Brush":
		return 0.7 * (((1 + (1.4 * pressure)) * width) - (0.5 * tilt) - (speed / 50.0))

	case "Calligraphy":
		return 0.9*(((1+pressure)*width)-(0.3*tilt)) + (0.1 * lastWidth)

	default:
		return p.baseWidth
	}
}

func (p *pen) getSegmentOpacity(point rmscene.Point, lastWidth float64) float64 {
	speed := float64(point.Speed) / 4.0
	pressure := float64(point.Pressure) / 255.0

	switch p.name {
	case "Pencil":
		opacity := (0.1 * -(speed / 35.0)) + pressure
		return clamp(opacity) - 0.1

	default:
		return p.baseOpacity
	}
}

func directionToTilt(direction uint8) float64 {
	return float64(direction) * (math.Pi * 2) / 255.0
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
