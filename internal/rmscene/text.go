package rmscene

import (
	"fmt"
	"strings"
)

// Paragraph represents a text paragraph with style
type Paragraph struct {
	Text    string
	Style   ParagraphStyle
	StartID CrdtID
}

// TextDocument represents a structured text document
type TextDocument struct {
	Paragraphs []Paragraph
}

// BuildTextDocument converts a Text object into a structured document
// by reconstructing strings from CRDT sequences and grouping by style
func BuildTextDocument(text *Text) (*TextDocument, error) {
	if text == nil || text.Items == nil {
		return &TextDocument{Paragraphs: []Paragraph{}}, nil
	}

	// Build ordered list of text items
	var textContent strings.Builder
	var itemIDs []CrdtID // Track item IDs for style mapping

	for _, item := range text.Items.Items {
		// Skip deleted items
		if item.DeletedLength > 0 {
			continue
		}

		// Extract text value
		if item.Value != nil {
			if str, ok := item.Value.(string); ok {
				textContent.WriteString(str)
				itemIDs = append(itemIDs, item.ItemID)
			}
		}
	}

	fullText := textContent.String()
	if fullText == "" {
		return &TextDocument{Paragraphs: []Paragraph{}}, nil
	}

	// Split by newlines to get paragraphs
	lines := strings.Split(fullText, "\n")

	// Get default style (usually mapped to CrdtID(0, 0))
	defaultStyle := StylePlain
	if styleValue, exists := text.Styles[CrdtID{Part1: 0, Part2: 0}]; exists {
		defaultStyle = styleValue.Value
	}

	// Build paragraphs
	doc := &TextDocument{
		Paragraphs: make([]Paragraph, 0, len(lines)),
	}

	for i, line := range lines {
		// Determine style for this paragraph
		// The style at CrdtID(0, 0) applies to the first paragraph only
		// Rest default to plain style
		paraStyle := StylePlain
		if i == 0 {
			paraStyle = defaultStyle
		}

		// Even if the line is empty, we want to preserve it for layout
		// (empty lines still take up vertical space)
		para := Paragraph{
			Text:    line,
			Style:   paraStyle,
			StartID: CrdtID{Part1: 0, Part2: uint64(i)}, // Use line index as ID
		}

		doc.Paragraphs = append(doc.Paragraphs, para)
	}

	return doc, nil
}

// String returns a string representation of the text document
func (doc *TextDocument) String() string {
	var sb strings.Builder
	for i, para := range doc.Paragraphs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(para.Text)
	}
	return sb.String()
}

// GetStyleName returns a human-readable name for a paragraph style
func GetStyleName(style ParagraphStyle) string {
	switch style {
	case StyleBasic:
		return "basic"
	case StylePlain:
		return "plain"
	case StyleHeading:
		return "heading"
	case StyleBold:
		return "bold"
	case StyleBullet:
		return "bullet"
	case StyleBullet2:
		return "bullet2"
	case StyleCheckbox:
		return "checkbox"
	case StyleCheckboxChecked:
		return "checkbox-checked"
	default:
		return fmt.Sprintf("unknown-%d", style)
	}
}
