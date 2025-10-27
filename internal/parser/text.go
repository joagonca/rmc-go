package parser

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

	// Build ordered list of text items and track their start positions
	var allText strings.Builder
	var itemStarts []struct {
		itemID CrdtID
		start  int
	}

	for _, item := range text.Items.Items {
		// Skip deleted items
		if item.DeletedLength > 0 {
			continue
		}

		// Extract text value
		if item.Value != nil {
			if str, ok := item.Value.(string); ok {
				startPos := allText.Len()
				allText.WriteString(str)
				itemStarts = append(itemStarts, struct {
					itemID CrdtID
					start  int
				}{itemID: item.ItemID, start: startPos})
			}
		}
	}

	fullText := allText.String()
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

	charPos := 0
	for i, line := range lines {
		// Determine style for this paragraph
		// The style at CrdtID(0, 0) applies to the first paragraph only
		// Rest default to plain style
		paraStyle := StylePlain
		if i == 0 {
			paraStyle = defaultStyle
		}

		// Find the item ID that corresponds to the start of this line
		startID := CrdtID{Part1: 0, Part2: uint64(i)} // Default fallback
		for _, is := range itemStarts {
			if is.start <= charPos && charPos < is.start+len(line)+1 {
				startID = is.itemID
				break
			}
		}

		// Even if the line is empty, we want to preserve it for layout
		// (empty lines still take up vertical space)
		para := Paragraph{
			Text:    line,
			Style:   paraStyle,
			StartID: startID,
		}

		doc.Paragraphs = append(doc.Paragraphs, para)

		// Move to next line (line length + newline)
		charPos += len(line) + 1
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
