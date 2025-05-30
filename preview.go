package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Pre-compiled regular expressions for performance
var (
	reHeaderHTML        = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	reUnorderedList     = regexp.MustCompile(`^(\s*)([-*+])\s+(.*)$`)
	reOrderedList       = regexp.MustCompile(`^(\s*)(\d+)\.\s+(.*)$`)
	reTaskList          = regexp.MustCompile(`^(\s*)[-*+]\s+\[([ x])\]\s+(.*)$`)
	reTableSeparator    = regexp.MustCompile(`^[\s|:\-]+$`)
	reLink              = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reImage             = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	reCodeSpanHTML      = regexp.MustCompile("`([^`]+)`")
	reBoldHTML1         = regexp.MustCompile(`\*\*([^\s*].*?[^\s*])\*\*`) // More specific bold
	reBoldHTML2         = regexp.MustCompile(`__([^\s_].*?[^\s_])__`)   // More specific bold
	reItalicHTML1       = regexp.MustCompile(`\*([^\s*].*?[^\s*])\*`)     // More specific italic
	reItalicHTML2       = regexp.MustCompile(`_([^\s_].*?[^\s_])_`)       // More specific italic
	reStrikethroughHTML = regexp.MustCompile(`~~([^~]+)~~`)
)

// Preview represents the markdown preview component
type Preview struct {
	content     *widget.RichText
	container   *fyne.Container
	visible     bool
	rawMarkdown string
}

// NewPreview creates a new preview instance
func NewPreview() *Preview {
	p := &Preview{
		content: widget.NewRichTextFromMarkdown(""), // Placeholder, updated by UpdateContent
		visible: true,
	}
	p.content.Wrapping = fyne.TextWrapWord
	return p
}

// Create creates the preview UI component
func (p *Preview) Create() fyne.CanvasObject {
	scrollContainer := container.NewScroll(
		container.NewPadded(p.content),
	)
	p.container = container.NewBorder(
		container.NewPadded(widget.NewCard("", "Preview", nil)),
		nil, nil, nil,
		scrollContainer,
	)
	return p.container
}

// UpdateContent updates the preview with new markdown content
func (p *Preview) UpdateContent(markdown string) {
	p.rawMarkdown = markdown
	segments := p.parseMarkdownToSegments(markdown)
	p.content.Segments = segments
	p.content.Refresh()
}

// parseMarkdownToSegments converts markdown to RichText segments
func (p *Preview) parseMarkdownToSegments(markdown string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	lines := strings.Split(markdown, "\n")

	inCodeBlock := false
	codeBlockContent := ""
	inList := false
	// listLevel := 0 // Not strictly needed for Fyne RichText simple list model
	inBlockquote := false
	blockquoteLines := []string{}
	paragraphLines := []string{}

	flushParagraph := func() {
		if len(paragraphLines) > 0 {
			text := strings.Join(paragraphLines, "\n") // Join with newline, then parse
			if strings.TrimSpace(text) != "" {
				segments = append(segments, p.parseInlineMarkdown(text)...)
				segments = append(segments, &widget.TextSegment{Text: "\n\n"}) // Paragraph separator
			}
			paragraphLines = []string{}
		}
	}

	flushBlockquote := func() {
		if len(blockquoteLines) > 0 {
			content := strings.Join(blockquoteLines, "\n")
			segments = append(segments, &widget.TextSegment{
				Text: "│ ", // Blockquote indicator
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNameDisabled,
				},
			})
			// Parse inline markdown within blockquote
			quotedSegments := p.parseInlineMarkdown(content)
			for _, seg := range quotedSegments {
				// Apply disabled style to text segments within the blockquote
				if textSeg, ok := seg.(*widget.TextSegment); ok {
					// If style is default, set disabled color. If already styled (e.g. bold), keep that but make it disabled.
					// This simple version just sets ColorName. A more complex one might merge.
					textSeg.Style.ColorName = theme.ColorNameDisabled
				}
				// Hyperlinks within blockquotes will retain their default link styling.
			}
			segments = append(segments, quotedSegments...)
			segments = append(segments, &widget.TextSegment{Text: "\n\n"}) // Separator after blockquote
			blockquoteLines = []string{}
			inBlockquote = false
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			flushParagraph()
			flushBlockquote()
			if inCodeBlock {
				if codeBlockContent != "" { // Ensure content is not just empty string
					segments = append(segments, &widget.TextSegment{
						Text: codeBlockContent,
						Style: widget.RichTextStyle{
							TextStyle: fyne.TextStyle{Monospace: true},
							ColorName: theme.ColorNameForeground, // Or a specific code color
						},
					})
				}
				segments = append(segments, &widget.TextSegment{Text: "\n\n"})
				inCodeBlock = false
				codeBlockContent = ""
			} else {
				inCodeBlock = true
				// language := strings.TrimPrefix(line, "```") // Language hint could be used later
			}
			continue
		}

		if inCodeBlock {
			if codeBlockContent != "" {
				codeBlockContent += "\n"
			}
			codeBlockContent += line
			continue
		}

		// Handle blockquotes
		if strings.HasPrefix(trimmed, ">") { // Allow ">text" and "> text"
			flushParagraph()
			if !inBlockquote {
				inBlockquote = true
			}
			blockquoteLines = append(blockquoteLines, strings.TrimPrefix(strings.TrimPrefix(trimmed, ">"), " "))
			continue
		} else if inBlockquote && trimmed != "" && !isList(trimmed) && !isHeader(trimmed) && !isRule(trimmed) { // Continue blockquote if line is not empty and not another block type
			blockquoteLines = append(blockquoteLines, trimmed)
			continue
		} else if inBlockquote {
			flushBlockquote() // End of blockquote (empty line or different block element)
		}

		// Handle headers
		if strings.HasPrefix(trimmed, "#") {
			flushParagraph(); flushBlockquote()
			level := 0
			textStart := 0
			for pos, r := range trimmed {
				if r == '#' {
					level++
				} else if r == ' ' {
					textStart = pos + 1
					break
				} else { // Not a valid header (e.g. #no_space)
					level = 0
					break
				}
			}

			if level > 0 && level <= 6 && textStart > 0 && textStart < len(trimmed) {
				headerText := strings.TrimSpace(trimmed[textStart:])
				segments = append(segments, &widget.TextSegment{
					Text: headerText,
					Style: widget.RichTextStyle{
						SizeName:  p.getHeaderSize(level),
						TextStyle: fyne.TextStyle{Bold: true},
					},
				})
				segments = append(segments, &widget.TextSegment{Text: "\n"}) // Single newline after header text
				if level <= 2 { // Separator for H1/H2
					segments = append(segments, &widget.SeparatorSegment{})
					segments = append(segments, &widget.TextSegment{Text: "\n"})
				}
				segments = append(segments, &widget.TextSegment{Text: "\n"}) // Extra newline for spacing after header block
				continue
			}
		}

		// Handle horizontal rules
		if isRule(trimmed) {
			flushParagraph(); flushBlockquote()
			segments = append(segments, &widget.SeparatorSegment{})
			segments = append(segments, &widget.TextSegment{Text: "\n\n"}) // Space after HR
			continue
		}

		// Handle tables (check after other block elements like headers, HRs)
		if strings.Contains(trimmed, "|") && i+1 < len(lines) && reTableSeparator.MatchString(strings.TrimSpace(lines[i+1])) {
			if tableSegments, consumed := p.parseTable(lines, i); tableSegments != nil {
				flushParagraph(); flushBlockquote()
				if inList { // End previous list if any
					segments = append(segments, &widget.TextSegment{Text: "\n"})
					inList = false
				}
				segments = append(segments, tableSegments...)
				i += consumed - 1 // outer loop will increment i
				continue
			}
		}
		
		// Handle lists
		isListItem := false
		// listIndent := 0 // For RichText, visual indent is primary

		if match := reUnorderedList.FindStringSubmatch(line); match != nil {
			flushParagraph(); flushBlockquote()
			isListItem = true
			// listIndent = len(match[1])
			if !inList {
				inList = true; /* listLevel = listIndent */
			}
			indentStr := strings.Repeat("  ", len(match[1])/2)
			segments = append(segments, &widget.TextSegment{Text: indentStr + "• "})
			segments = append(segments, p.parseInlineMarkdown(match[3])...)
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}
		if match := reOrderedList.FindStringSubmatch(line); match != nil {
			flushParagraph(); flushBlockquote()
			isListItem = true
			// listIndent = len(match[1])
			if !inList {
				inList = true; /* listLevel = listIndent */
			}
			indentStr := strings.Repeat("  ", len(match[1])/2)
			segments = append(segments, &widget.TextSegment{Text: indentStr + match[2] + ". "})
			segments = append(segments, p.parseInlineMarkdown(match[3])...)
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}
		if match := reTaskList.FindStringSubmatch(line); match != nil {
			flushParagraph(); flushBlockquote()
			isListItem = true
			// listIndent = len(match[1])
			if !inList {
				inList = true; /* listLevel = listIndent */
			}
			checkbox := "☐ "
			if match[2] == "x" { checkbox = "☑ " }
			indentStr := strings.Repeat("  ", len(match[1])/2)
			segments = append(segments, &widget.TextSegment{Text: indentStr + checkbox})
			segments = append(segments, p.parseInlineMarkdown(match[3])...)
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}

		if !isListItem && inList { // Current line is not a list item, but we were in a list
			inList = false
			// Add a bit more space after a list if followed by a paragraph
			if trimmed != "" { // Only if followed by non-empty line
				segments = append(segments, &widget.TextSegment{Text: "\n"})
			}
		}

		// Handle empty lines (paragraph breaks)
		if trimmed == "" {
			flushParagraph() // Flushes existing paragraph and adds \n\n
			// Multiple empty lines won't add more \n\n due to paragraphLines being empty
			continue
		}

		// Regular paragraph line
		paragraphLines = append(paragraphLines, line) // Keep original line for multi-line paragraphs
	}

	flushParagraph()
	flushBlockquote()
	// If ending with code block, it already added \n\n. If list, might need one.
	if inList { // Ensure space after a list if it's the last element
		segments = append(segments, &widget.TextSegment{Text: "\n"})
	}


	return segments
}

// Helper functions for parseMarkdownToSegments
func isList(trimmedLine string) bool {
	return reUnorderedList.MatchString(trimmedLine) ||
		reOrderedList.MatchString(trimmedLine) ||
		reTaskList.MatchString(trimmedLine)
}

func isHeader(trimmedLine string) bool {
	return strings.HasPrefix(trimmedLine, "#")
}

func isRule(trimmedLine string) bool {
	return trimmedLine == "---" || trimmedLine == "***" || trimmedLine == "___"
}


// parseInlineMarkdown handles inline markdown elements (code, bold, italic, links, etc.)
func (p *Preview) parseInlineMarkdown(text string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	parts := p.splitByCode(text) // Handle code spans first

	for i, part := range parts {
		if i%2 == 1 { // This part is code
			segments = append(segments, &widget.TextSegment{
				Text: part,
				Style: widget.RichTextStyle{
					TextStyle: fyne.TextStyle{Monospace: true},
					ColorName: theme.ColorNameForeground, // Or a specific code color
				},
			})
		} else { // This part is regular text, parse for other inline elements
			segments = append(segments, p.parseFormattedText(part)...)
		}
	}
	return segments
}

// splitByCode splits text by backticks for code spans, preserving empty segments.
// Output: [non-code, code, non-code, code, ..., non-code (possibly empty)]
func (p *Preview) splitByCode(text string) []string {
	var parts []string
	var current strings.Builder
	
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '`' {
			// Add segment before the backtick
			current.WriteString(text[start:i])
			parts = append(parts, current.String())
			current.Reset()
			
			// Find closing backtick
			endCode := strings.IndexByte(text[i+1:], '`')
			if endCode == -1 { // Unclosed backtick
				// Treat rest of string as literal
				current.WriteByte(text[i]) // Add the backtick itself
				start = i + 1
				// continue to append rest of string later
				break 
			}
			// Add code content
			parts = append(parts, text[i+1:i+1+endCode])
			start = i + 1 + endCode + 1
			i = start -1 // loop will increment
		}
	}
	// Add any remaining part of the string
	current.WriteString(text[start:])
	parts = append(parts, current.String())
	
	// Ensure structure is [non-code (part 0)], [code (part 1)], [non-code (part 2)] ...
    // If text starts with code, first part is empty. If ends with code, last non-code part can be empty.
	// The loop `if i%2 == 1` handles code parts correctly.

	return parts
}


// parseFormattedText handles links, images, and calls parseNestedStyles for bold/italic/strikethrough.
func (p *Preview) parseFormattedText(text string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	var currentText strings.Builder
	i := 0

	// First, replace images with a placeholder text representation
	// This is simpler than trying to interleave image parsing with other formatting.
	text = reImage.ReplaceAllStringFunc(text, func(match string) string {
		submatches := reImage.FindStringSubmatch(match)
		if len(submatches) == 3 {
			altText := submatches[1]
			// src := submatches[2] // URL not directly used in RichText for images here
			if altText == "" {
				altText = "image"
			}
			return fmt.Sprintf("[🖼️ %s]", altText) // Placeholder for image
		}
		return match
	})

	for i < len(text) {
		// Check for Links: [text](url)
		if text[i] == '[' {
			linkMatchIndices := reLink.FindStringSubmatchIndex(text[i:])
			if linkMatchIndices != nil && linkMatchIndices[0] == 0 { // Match starts at current position
				if currentText.Len() > 0 {
					segments = append(segments, &widget.TextSegment{Text: currentText.String()})
					currentText.Reset()
				}

				linkDisplayText := text[i+linkMatchIndices[2] : i+linkMatchIndices[3]]
				linkURL := text[i+linkMatchIndices[4] : i+linkMatchIndices[5]]

				if parsedURL, err := url.Parse(linkURL); err == nil {
					// Parse link display text for further formatting (bold, italic)
					// Starting with a fresh style for link content.
					linkContentSegments := p.parseNestedStyles(linkDisplayText, widget.RichTextStyle{})
					for _, seg := range linkContentSegments {
						if ts, ok := seg.(*widget.TextSegment); ok {
							segments = append(segments, &widget.HyperlinkSegment{
								Text:      ts.Text,
								URL:       parsedURL,
								TextStyle: ts.Style.TextStyle, // Apply text styles (bold, italic)
							})
						} else {
                             // Should not happen if parseNestedStyles only produces TextSegments
                            segments = append(segments, seg) 
                        }
					}
				} else {
					// Invalid URL, treat as plain text
					currentText.WriteString(text[i : i+linkMatchIndices[1]])
				}
				i += linkMatchIndices[1]
				continue
			}
		}

		// If no link, it might be bold, italic, etc.
		// This part is now handled by parseNestedStyles if we call it directly.
		// However, parseFormattedText is the top-level for non-code, non-image text.
		// It should break text into "link" or "other stuff".
		// "other stuff" is then passed to parseNestedStyles.

		// The current structure means parseFormattedText handles links,
		// and for any text *not* part of a link, it should eventually be processed by parseNestedStyles.
		// This implies that if no link is found, the character is added to currentText,
		// and *after* the loop, currentText is processed by parseNestedStyles.

		currentText.WriteByte(text[i])
		i++
	}

	// Process any remaining text accumulated in currentText for bold/italic/strikethrough
	if currentText.Len() > 0 {
		segments = append(segments, p.parseNestedStyles(currentText.String(), widget.RichTextStyle{})...)
	}

	return segments
}

// parseNestedStyles recursively parses text for bold, italic, strikethrough,
// applying them on top of a given currentStyle.
func (p *Preview) parseNestedStyles(text string, currentStyle widget.RichTextStyle) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	var accumulatedText strings.Builder
	i := 0

	for i < len(text) {
		// Precedence: Bold > Strikethrough > Italic (arbitrary but consistent)

		// Bold: **text** or __text__
		appliedStyle := false
		if i+1 < len(text) {
			marker := ""
			if text[i:i+2] == "**" { marker = "**" } else if text[i:i+2] == "__" { marker = "__" }

			if marker != "" {
				endPos := strings.Index(text[i+len(marker):], marker)
				if endPos != -1 {
					if accumulatedText.Len() > 0 {
						segments = append(segments, &widget.TextSegment{Text: accumulatedText.String(), Style: currentStyle})
						accumulatedText.Reset()
					}
					content := text[i+len(marker) : i+len(marker)+endPos]
					styleWithBold := currentStyle; styleWithBold.TextStyle.Bold = true
					segments = append(segments, p.parseNestedStyles(content, styleWithBold)...)
					i += len(marker) + endPos + len(marker)
					appliedStyle = true
				}
			}
		}
		if appliedStyle { continue }


		// Strikethrough: ~~text~~
		if i+1 < len(text) && text[i:i+2] == "~~" {
			endPos := strings.Index(text[i+2:], "~~")
			if endPos != -1 {
				if accumulatedText.Len() > 0 {
					segments = append(segments, &widget.TextSegment{Text: accumulatedText.String(), Style: currentStyle})
					accumulatedText.Reset()
				}
				content := text[i+2 : i+2+endPos]
				// Strikethrough: for Fyne, often done with color. It doesn't nest other styles well if it sets color.
				// Here, we make it a terminal style for its content for simplicity.
				styleWithStrike := currentStyle; styleWithStrike.ColorName = theme.ColorNameDisabled
				segments = append(segments, &widget.TextSegment{Text: content, Style: styleWithStrike})
				i += 2 + endPos + 2
				appliedStyle = true
			}
		}
		if appliedStyle { continue }

		// Italic: *text* or _text_
		// Ensure it's not part of a bold marker already handled or a different construct.
		if (text[i] == '*' || text[i] == '_') {
			markerChar := text[i]
			// Avoid consuming `*` from `**` if bold is handled by single char marker logic elsewhere.
			// Here, bold `**` `__` is handled distinctly, so single `*` `_` are for italic.
            // A more robust parser would check context (e.g. CommonMark rules for intra-word emphasis)
			
            // Basic check: not part of a longer marker like ** or __
            if !((markerChar == '*' && i+1 < len(text) && text[i+1] == '*') ||
                 (markerChar == '_' && i+1 < len(text) && text[i+1] == '_')) {

                endPos := strings.IndexByte(text[i+1:], markerChar)
                if endPos != -1 {
                    // Add more sophisticated checks here for valid italic pairs if needed.
                    if accumulatedText.Len() > 0 {
                        segments = append(segments, &widget.TextSegment{Text: accumulatedText.String(), Style: currentStyle})
                        accumulatedText.Reset()
                    }
                    content := text[i+1 : i+1+endPos]
                    styleWithItalic := currentStyle; styleWithItalic.TextStyle.Italic = true
                    segments = append(segments, p.parseNestedStyles(content, styleWithItalic)...)
                    i += 1 + endPos + 1
                    appliedStyle = true
                }
            }
		}
		if appliedStyle { continue }

		// If no style marker found at i, accumulate character
		accumulatedText.WriteByte(text[i])
		i++
	}

	// Add any remaining accumulated plain text
	if accumulatedText.Len() > 0 {
		segments = append(segments, &widget.TextSegment{Text: accumulatedText.String(), Style: currentStyle})
	}
	return segments
}


// getHeaderSize returns the appropriate size name for header level
func (p *Preview) getHeaderSize(level int) fyne.ThemeSizeName {
	switch level {
	case 1:
		return theme.SizeNameHeadingText
	case 2:
		return theme.SizeNameSubHeadingText
	default: // H3, H4, H5, H6 get regular text size but bold (handled by caller)
		return theme.SizeNameText
	}
}

// ToggleVisibility toggles the preview pane visibility
func (p *Preview) ToggleVisibility() {
	if p.visible {
		p.container.Hide()
	} else {
		p.container.Show()
	}
	p.visible = !p.visible
}


// GetHTML returns the markdown converted to HTML
func (p *Preview) GetHTML() string {
	// Basic HTML template with styling (GitHub-like)
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Markdown Export</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji"; font-size: 16px; line-height: 1.5; word-wrap: break-word; color: #24292f; background-color: #ffffff; margin: 0; padding: 0; }
        .markdown-body { box-sizing: border-box; min-width: 200px; max-width: 980px; margin: 0 auto; padding: 45px; }
        @media (max-width: 767px) { .markdown-body { padding: 15px; } }
        h1, h2, h3, h4, h5, h6 { margin-top: 24px; margin-bottom: 16px; font-weight: 600; line-height: 1.25; }
        h1 { font-size: 2em; } h2 { font-size: 1.5em; } h3 { font-size: 1.25em; } h4 { font-size: 1em; } h5 { font-size: .875em; } h6 { font-size: .85em; color: #57606a; }
        h1, h2 { padding-bottom: .3em; border-bottom: 1px solid #d0d7de; }
        p { margin-top: 0; margin-bottom: 16px; }
        a { color: #0969da; text-decoration: none; } a:hover { text-decoration: underline; }
        code { padding: .2em .4em; margin: 0; font-size: 85%; background-color: rgba(175,184,193,0.2); border-radius: 6px; font-family: ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace; }
        pre { margin-top: 0; margin-bottom: 16px; padding: 16px; overflow: auto; font-size: 85%; line-height: 1.45; background-color: #f6f8fa; border-radius: 6px; word-wrap: normal;}
        pre code { display: inline; max-width: auto; padding: 0; margin: 0; overflow: visible; line-height: inherit; background-color: transparent; border: 0; font-size: 100%; }
        blockquote { margin: 0 0 16px; padding: 0 1em; color: #57606a; border-left: .25em solid #d0d7de; }
        ul, ol { margin-top: 0; margin-bottom: 16px; padding-left: 2em; } ul ul, ul ol, ol ol, ol ul { margin-top:0; margin-bottom: 0; }
        li { word-wrap: break-all; } li > p { margin-top: 16px; } li + li { margin-top: .25em; }
        hr { height: .25em; padding: 0; margin: 24px 0; background-color: #d0d7de; border: 0; }
        table { display: block; width: max-content; max-width:100%; overflow: auto; border-spacing: 0; border-collapse: collapse; margin-top: 0; margin-bottom: 16px; }
        table th, table td { padding: 6px 13px; border: 1px solid #d0d7de; }
        table th { font-weight: 600; } table tr { background-color: #ffffff; } table tr:nth-child(2n) { background-color: #f6f8fa; }
        img { max-width: 100%; box-sizing: content-box; background-color: #ffffff; }
        strong { font-weight: 600; } em { font-style: italic; } del { text-decoration: line-through; }
		input[type="checkbox"] { margin-right: 0.5em; }
    </style>
</head>
<body> <article class="markdown-body"> %s </article> </body>
</html>`

	htmlContent := p.convertMarkdownToHTML(p.rawMarkdown)
	return fmt.Sprintf(htmlTemplate, htmlContent)
}

// convertMarkdownToHTML performs markdown to HTML conversion
func (p *Preview) convertMarkdownToHTML(markdown string) string {
	var html strings.Builder
	lines := strings.Split(markdown, "\n")

	inCodeBlock := false
	var codeBlockLang string
	inList := false
	listType := ""    // "ul" or "ol"
	// listIndentStack := []int{} // For proper nested list handling (more complex)
	inBlockquote := false
	blockquoteBuffer := []string{} // Buffer lines for a blockquote block

	flushParagraph := func(paragraphLines []string) {
		if len(paragraphLines) > 0 {
			content := strings.TrimSpace(strings.Join(paragraphLines, "\n"))
			if content != "" {
				html.WriteString("<p>")
				html.WriteString(p.processInlineHTML(content))
				html.WriteString("</p>\n")
			}
		}
	}
	
	var currentParagraphLines []string

	closeList := func() {
		if inList {
			// For complex nesting, would pop from listIndentStack and close multiple lists
			html.WriteString(fmt.Sprintf("</%s>\n", listType))
			inList = false
			listType = ""
		}
	}
	
	flushBlockquote := func() {
		if inBlockquote && len(blockquoteBuffer) > 0 {
			content := strings.Join(blockquoteBuffer, "\n")
			html.WriteString("<blockquote>\n")
			// Recursively convert the content of the blockquote
			html.WriteString(p.convertMarkdownToHTML(content))
			html.WriteString("</blockquote>\n")
			blockquoteBuffer = []string{}
			inBlockquote = false
		}
	}


	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code blocks
		if strings.HasPrefix(line, "```") {
			flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
			flushBlockquote(); closeList()
			if inCodeBlock {
				html.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				inCodeBlock = true
				codeBlockLang = escapeHTMLAttribute(strings.TrimSpace(strings.TrimPrefix(line, "```")))
				langClass := ""
				if codeBlockLang != "" {
					langClass = fmt.Sprintf(` class="language-%s"`, codeBlockLang)
				}
				html.WriteString(fmt.Sprintf("<pre><code%s>", langClass)) // No newline after <code>
			}
			continue
		}
		if inCodeBlock {
			html.WriteString(escapeHTMLContent(line) + "\n") // Add newline for each line in code block
			continue
		}
		
		// Check for end of blockquote if not a continued blockquote line
		if inBlockquote && !strings.HasPrefix(trimmed, ">") {
            // If line is empty, or starts a new element type, flush blockquote
			if trimmed == "" || isHeader(trimmed) || isList(trimmed) || isRule(trimmed) || (strings.Contains(trimmed, "|") && i+1 < len(lines) && reTableSeparator.MatchString(strings.TrimSpace(lines[i+1]))) {
                 flushBlockquote()
            }
        }

		// Blockquotes
		if strings.HasPrefix(trimmed, ">") {
			flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
			closeList()
			if !inBlockquote {
				inBlockquote = true
			}
			blockquoteBuffer = append(blockquoteBuffer, strings.TrimPrefix(strings.TrimPrefix(trimmed, ">"), " "))
			continue
		}
        // If in blockquote and line isn't empty, it might be a continuation if it's not another block type
        if inBlockquote && trimmed != "" {
             blockquoteBuffer = append(blockquoteBuffer, trimmed) // Add non-prefixed line as part of quote
             continue
        }


		// Headers
		if match := reHeaderHTML.FindStringSubmatch(trimmed); match != nil {
			flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
			flushBlockquote(); closeList()
			level := len(match[1])
			headerText := p.processInlineHTML(match[2])
			html.WriteString(fmt.Sprintf("<h%d>%s</h%d>\n", level, headerText, level))
			continue
		}

		// Horizontal rules
		if isRule(trimmed) {
			flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
			flushBlockquote(); closeList()
			html.WriteString("<hr>\n")
			continue
		}
		
		// Tables
		if strings.Contains(trimmed, "|") && i+1 < len(lines) && reTableSeparator.MatchString(strings.TrimSpace(lines[i+1])) {
			if tableHTML, consumed := p.convertTableToHTML(lines, i); tableHTML != "" {
				flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
				flushBlockquote(); closeList()
				html.WriteString(tableHTML)
				i += consumed -1 // outer loop will increment
				continue
			}
		}


		// Lists
		// Basic list handling, does not support complex nesting well.
		listItem := false
		currentListType := ""
		listItemContent := ""

		if match := reUnorderedList.FindStringSubmatch(line); match != nil { // keep original line for indent
			listItem = true; currentListType = "ul"; listItemContent = match[3]
		} else if match := reOrderedList.FindStringSubmatch(line); match != nil {
			listItem = true; currentListType = "ol"; listItemContent = match[3]
		} else if match := reTaskList.FindStringSubmatch(line); match != nil {
			listItem = true; currentListType = "ul" // Task lists are <ul>
			checked := ""; if match[2] == "x" { checked = " checked" }
			listItemContent = fmt.Sprintf(`<input type="checkbox" disabled%s> %s`, checked, match[3])
		}

		if listItem {
			flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
			flushBlockquote()
			// indent := len(reUnorderedList.FindStringSubmatch(line)[1]) // Example, needs proper regex matching
			if !inList || listType != currentListType /* || indent != currentIndent */ {
				closeList() // Close previous list if type changes or indent changes significantly
				html.WriteString(fmt.Sprintf("<%s>\n", currentListType))
				inList = true
				listType = currentListType
				// currentIndent = indent
			}
			html.WriteString(fmt.Sprintf("<li>%s</li>\n", p.processInlineHTML(listItemContent)))
			continue
		}
		
		// If we were in a list, but current line is not a list item
		if inList && trimmed != "" { // An indented non-list item line might be part of previous li (complex)
			closeList()
		} else if inList && trimmed == "" {
            // Empty line in a list might continue the list or break it.
            // For simplicity, let's say an empty line *after* list items might mean end of list.
            // This is tricky. If next line is indented list item, list continues.
            // The logic here assumes non-list item means list ends.
		}


		// Paragraphs and empty lines
		if trimmed == "" {
			flushParagraph(currentParagraphLines); currentParagraphLines = []string{}
			// An empty line could also terminate a list if not handled above
			if inList { closeList() } 
			if inBlockquote { flushBlockquote() } // Empty line also ends blockquote
		} else {
			currentParagraphLines = append(currentParagraphLines, line)
		}
	}

	flushParagraph(currentParagraphLines)
	flushBlockquote()
	closeList()

	return html.String()
}


// processInlineHTML handles inline markdown elements for HTML conversion
func (p *Preview) processInlineHTML(text string) string {
	// Order of replacement is important.
	// 1. Escape HTML special characters to prevent XSS or misinterpretation.
	processedText := escapeHTMLContent(text)

	// 2. Code spans `code` (should be processed before other formatting like * or _)
	processedText = reCodeSpanHTML.ReplaceAllString(processedText, "<code>$1</code>")

	// 3. Images ![alt](src)
	processedText = reImage.ReplaceAllStringFunc(processedText, func(match string) string {
		submatches := reImage.FindStringSubmatch(match) // Use the precompiled reImage
		if len(submatches) == 3 {
			altText := submatches[1]
			src := submatches[2]
			return fmt.Sprintf(`<img src="%s" alt="%s">`, escapeHTMLAttribute(src), escapeHTMLAttribute(altText))
		}
		return match
	})

	// 4. Links [text](url)
	processedText = reLink.ReplaceAllStringFunc(processedText, func(match string) string {
		submatches := reLink.FindStringSubmatch(match) // Use the precompiled reLink
		if len(submatches) == 3 {
			linkText := submatches[1] // Link text itself might contain other inline markdown (e.g. bold)
			href := submatches[2]
			// Recursively process link text for other inline elements (bold, italic, etc.)
			// but not for nested links or images.
			// A simplified approach for HTML: process bold/italic inside linkText *after* link tag is formed on raw linkText.
			// Or, process linkText for bold/italic, then use that as the display.
			// For now, keep it simple: linkText is used as is, then outer formatting applies.
			// The regexes below for bold/italic will catch them if they are outside links.
			// To process *inside* links, process linkText:
			// processedLinkText := p.processInlineHTML(linkText) // Careful: recursion and context
			return fmt.Sprintf(`<a href="%s">%s</a>`, escapeHTMLAttribute(href), linkText) // Simpler: linkText as is
		}
		return match
	})

	// 5. Bold **text** or __text__
	processedText = reBoldHTML1.ReplaceAllString(processedText, "<strong>$1</strong>")
	processedText = reBoldHTML2.ReplaceAllString(processedText, "<strong>$1</strong>")

	// 6. Italic *text* or _text_ (must come after bold to handle * vs **)
	processedText = reItalicHTML1.ReplaceAllString(processedText, "<em>$1</em>")
	processedText = reItalicHTML2.ReplaceAllString(processedText, "<em>$1</em>")

	// 7. Strikethrough ~~text~~
	processedText = reStrikethroughHTML.ReplaceAllString(processedText, "<del>$1</del>")
	
	return processedText
}


// escapeHTMLContent escapes HTML special characters
func escapeHTMLContent(text string) string {
	text = strings.ReplaceAll(text, "&", "&")
	text = strings.ReplaceAll(text, "<", "<")
	text = strings.ReplaceAll(text, ">", ">")
	return text
}

// escapeHTMLAttribute escapes characters for HTML attributes
func escapeHTMLAttribute(text string) string {
	text = escapeHTMLContent(text) // Basic escaping first
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "'") // ' is not universally supported in HTML4
	return text
}

// parseTable attempts to parse lines as a markdown table for RichText
func (p *Preview) parseTable(lines []string, startIndex int) ([]widget.RichTextSegment, int) {
	// Assumes caller has already checked for valid table start (line with | and separator line)
	var segments []widget.RichTextSegment
	// segments = append(segments, &widget.TextSegment{Text: "\n"}) // Space before table

	headerLine := strings.TrimSpace(lines[startIndex])
	// Separator line is lines[startIndex+1]

	numConsumed := 0
	// Process header
	headerCells := strings.Split(strings.Trim(headerLine, "|"), "|")
	for j, cell := range headerCells {
		trimmedCell := strings.TrimSpace(cell)
		if trimmedCell != "" {
			segments = append(segments, &widget.TextSegment{
				Text:  trimmedCell,
				Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
			})
		}
		if j < len(headerCells)-1 {
			segments = append(segments, &widget.TextSegment{Text: " | "}) // Visual separator
		}
	}
	segments = append(segments, &widget.TextSegment{Text: "\n"})
	numConsumed++

	// Add visual separator for RichText (replaces the ---|--- line)
	segments = append(segments, &widget.SeparatorSegment{})
	segments = append(segments, &widget.TextSegment{Text: "\n"})
	numConsumed++


	// Process body rows
	for i := startIndex + 2; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)
		if !strings.Contains(trimmedLine, "|") || !strings.HasPrefix(trimmedLine, "|") && !strings.HasSuffix(trimmedLine, "|") && len(strings.Split(trimmedLine, "|")) <=1  {
			break // End of table
		}
		
		rowCells := strings.Split(strings.Trim(trimmedLine, "|"), "|")
		for j, cell := range rowCells {
			trimmedCell := strings.TrimSpace(cell)
			// For RichText, inline markdown in table cells is complex, so render as plain text.
			if trimmedCell != "" {
				segments = append(segments, &widget.TextSegment{Text: trimmedCell})
			}
			if j < len(rowCells)-1 {
				segments = append(segments, &widget.TextSegment{Text: " | "})
			}
		}
		segments = append(segments, &widget.TextSegment{Text: "\n"})
		numConsumed++
	}
	
	segments = append(segments, &widget.TextSegment{Text: "\n"}) // Space after table
	return segments, numConsumed
}

// convertTableToHTML converts markdown table lines to HTML
func (p *Preview) convertTableToHTML(lines []string, startIndex int) (string, int) {
	// Assumes caller has already checked for valid table start
	var html strings.Builder
	html.WriteString("<table>\n<thead>\n<tr>\n")

	headerLine := strings.TrimSpace(lines[startIndex])
	numConsumed := 0

	// Parse header
	headers := strings.Split(strings.Trim(headerLine, "|"), "|")
	for _, header := range headers {
		trimmedHeader := strings.TrimSpace(header)
		// processInlineHTML for content within th/td
		html.WriteString(fmt.Sprintf("<th>%s</th>\n", p.processInlineHTML(trimmedHeader)))
	}
	html.WriteString("</tr>\n</thead>\n<tbody>\n")
	numConsumed += 2 // Header line + separator line

	// Parse body rows
	for i := startIndex + 2; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)
        // More robust check for end of table: if line doesn't look like a table row.
		if !strings.Contains(trimmedLine, "|") || (!strings.HasPrefix(trimmedLine, "|") && !strings.HasSuffix(trimmedLine, "|") && len(strings.Split(trimmedLine, "|")) <= 1) {
			break 
		}

		html.WriteString("<tr>\n")
		cells := strings.Split(strings.Trim(trimmedLine, "|"), "|")
		for _, cell := range cells {
			trimmedCell := strings.TrimSpace(cell)
			html.WriteString(fmt.Sprintf("<td>%s</td>\n", p.processInlineHTML(trimmedCell)))
		}
		html.WriteString("</tr>\n")
		numConsumed++
	}

	html.WriteString("</tbody>\n</table>\n")
	return html.String(), numConsumed
}

