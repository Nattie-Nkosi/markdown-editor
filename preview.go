package main

import (
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
		content: widget.NewRichTextFromMarkdown(""),
		visible: true,
	}

	// Configure the RichText widget
	p.content.Wrapping = fyne.TextWrapWord

	return p
}

// Create creates the preview UI component
func (p *Preview) Create() fyne.CanvasObject {
	// Create a custom scroll container with padding
	scrollContainer := container.NewScroll(
		container.NewPadded(p.content),
	)

	// Create the main container with a card header
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

	// Create custom rich text segments for better rendering
	segments := p.parseMarkdownToSegments(markdown)
	p.content.Segments = segments
	p.content.Refresh()
}

// parseMarkdownToSegments converts markdown to RichText segments with better formatting
func (p *Preview) parseMarkdownToSegments(markdown string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	lines := strings.Split(markdown, "\n")

	inCodeBlock := false
	codeBlockContent := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// End of code block
				segments = append(segments, &widget.TextSegment{
					Text: codeBlockContent,
					Style: widget.RichTextStyle{
						TextStyle: fyne.TextStyle{Monospace: true},
					},
				})
				segments = append(segments, &widget.TextSegment{
					Text: "\n",
				})
				inCodeBlock = false
				codeBlockContent = ""
				} else {
				// Start of code block
				inCodeBlock = true
				strings.TrimPrefix(trimmed, "```")
				if i > 0 && len(segments) > 0 { // ensure segments is not empty before potentially adding a newline
					// Add a newline before the code block if it's not the first element
					// and the previous segment wasn't already ending with a newline or was a separator.
					lastSeg := segments[len(segments)-1]
					if ts, ok := lastSeg.(*widget.TextSegment); !(ok && strings.HasSuffix(ts.Text, "\n")) {
						if _, okSep := lastSeg.(*widget.SeparatorSegment); !okSep {
							// segments = append(segments, &widget.TextSegment{Text: "\n"})
                            // It seems the original code tried to add a newline if i > 0, 
                            // but code blocks often don't need an extra preceding newline if they follow other block elements directly.
                            // Let's rely on the natural newlines or the newlines after headers/paragraphs.
                            // The crucial part is handling the newline *after* the code block.
						}
					}
				}
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

		// Handle headers
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, r := range trimmed {
				if r == '#' {
					level++
				} else {
					break
				}
			}

			headerText := strings.TrimSpace(trimmed[level:])
			if i > 0 && len(segments) > 0 { // Ensure not the first line and segments exist
				// Add a newline before header if not first element, similar to code block logic
                lastSeg := segments[len(segments)-1]
                if ts, ok := lastSeg.(*widget.TextSegment); !(ok && strings.HasSuffix(ts.Text, "\n")) {
                     if _, okSep := lastSeg.(*widget.SeparatorSegment); !okSep {
                        // segments = append(segments, &widget.TextSegment{Text: "\n"})
                        // Again, let natural flow handle this or specific newlines from paragraph ends.
                     }
                }
			}

			segments = append(segments, &widget.TextSegment{
				Text: headerText,
				Style: widget.RichTextStyle{
					SizeName:  p.getHeaderSize(level),
					TextStyle: fyne.TextStyle{Bold: true},
				},
			})
			segments = append(segments, &widget.TextSegment{Text: "\n"})

			// Add separator for H1 and H2
			if level <= 2 {
				segments = append(segments, &widget.SeparatorSegment{})
				segments = append(segments, &widget.TextSegment{Text: "\n"})
			}
			continue
		}

		// Handle lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			segments = append(segments, &widget.TextSegment{
				Text: "• " + trimmed[2:], // Consider using widget.ListSegment for richer list support if available/desired
			})
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}

		// Handle numbered lists
		if len(trimmed) > 2 && trimmed[1] == '.' && trimmed[0] >= '0' && trimmed[0] <= '9' {
			// This basic handling can be improved for multi-digit numbers
			// For now, just treat as text.
			segments = append(segments, &widget.TextSegment{
				Text: trimmed,
			})
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}

		// Handle blockquotes
		if strings.HasPrefix(trimmed, "> ") {
			segments = append(segments, &widget.TextSegment{
				Text: "│ " + trimmed[2:],
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNameDisabled, // Or a custom style
				},
			})
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}

		// Handle horizontal rules
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			segments = append(segments, &widget.SeparatorSegment{})
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}

		// Handle images - convert to a safe placeholder
		if strings.Contains(line, "![") && strings.Contains(line, "](") {
			processedLine := p.processImageSyntax(line)
			segments = append(segments, p.parseInlineMarkdown(processedLine)...)
			segments = append(segments, &widget.TextSegment{Text: "\n"})
			continue
		}

		// Regular paragraph
		if trimmed != "" {
			segments = append(segments, p.parseInlineMarkdown(line)...)
			segments = append(segments, &widget.TextSegment{Text: "\n"})
		} else if i < len(lines)-1 && len(segments) > 0 {
			// Empty line, add a visual break if it's not the start and previous wasn't already a full newline.
			// This helps separate paragraphs.
            lastSeg := segments[len(segments)-1]
            if ts, ok := lastSeg.(*widget.TextSegment); ok && ts.Text == "\n" {
                // Already have a single newline segment, maybe make it a bigger break or paragraph style
            } else {
			    segments = append(segments, &widget.TextSegment{Text: "\n"})
            }
		}
	}
    // Remove trailing newlines if they are redundant or ensure a single one for proper spacing
    if len(segments) > 0 {
        lastSeg := segments[len(segments)-1]
        if ts, ok := lastSeg.(*widget.TextSegment); ok && ts.Text == "\n" && len(segments) > 1 {
             prevSeg := segments[len(segments)-2]
             if pts, ok2 := prevSeg.(*widget.TextSegment); ok2 && strings.HasSuffix(pts.Text, "\n\n") {
                 // If previous text segment already ends with double newline, this extra \n segment might be too much
                 // segments = segments[:len(segments)-1] // Or adjust logic for paragraph spacing
             }
        }
    }


	return segments
}

// processImageSyntax converts image markdown to a safe placeholder
func (p *Preview) processImageSyntax(line string) string {
	result := line

	for strings.Contains(result, "![") && strings.Contains(result, "](") {
		start := strings.Index(result, "![")
		altEnd := strings.Index(result[start:], "]")
		if altEnd == -1 {
			break
		}

		urlStart := start + altEnd
		if urlStart >= len(result) || result[urlStart:urlStart+2] != "](" {
			break
		}

		urlEnd := strings.Index(result[urlStart+2:], ")")
		if urlEnd == -1 {
			break
		}

		altText := result[start+2 : start+altEnd]
		imageURL := result[urlStart+2 : urlStart+2+urlEnd]

		placeholder := fmt.Sprintf("[🖼️ Image: %s]", altText)
		if imageURL != "" {
			placeholder = fmt.Sprintf("[🖼️ %s: %s]", altText, imageURL)
		}

		result = result[:start] + placeholder + result[urlStart+2+urlEnd+1:]
	}

	return result
}

// parseInlineMarkdown handles inline markdown elements
func (p *Preview) parseInlineMarkdown(text string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment

	parts := strings.Split(text, "`")
	for i, part := range parts {
		if i%2 == 0 {
			// Regular text - parse for other markdown
			segments = append(segments, p.parseTextFormatting(part)...)
		} else {
			// Inline Code
			if part != "" { // Avoid creating empty code segments
				segments = append(segments, &widget.TextSegment{ // Changed to TextSegment with Monospace style
						Text: part,
						Style: widget.RichTextStyle{
							TextStyle: fyne.TextStyle{Monospace: true},
						},
					})
			}
		}
	}

	return segments
}

// parseTextFormatting handles bold, italic, and links
func (p *Preview) parseTextFormatting(text string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	current := text

	// This function needs a more robust parser for nested and complex cases.
	// For now, let's handle bold and italic separately and sequentially.
	// A better way would be to tokenize and then build segments.

	// Priority: Bold (**) then Italic (*)
	// This simplified parser doesn't handle nesting like **bold *italic* bold** well.

	var process func(input string, depth int) []widget.RichTextSegment
	process = func(input string, depth int) []widget.RichTextSegment {
		localSegments := []widget.RichTextSegment{}

		// Bold: **text** or __text__
		if strings.Contains(input, "**") {
			parts := strings.SplitN(input, "**", 3)
			if len(parts) == 3 { // Found a pair
				localSegments = append(localSegments, process(parts[0], depth+1)...)
				localSegments = append(localSegments, &widget.TextSegment{
					Text: parts[1],
					Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
				})
				localSegments = append(localSegments, process(parts[2], depth+1)...)
				return localSegments
			}
		}
		// Italic: *text* or _text_
		if strings.Contains(input, "*") {
			parts := strings.SplitN(input, "*", 3)
			if len(parts) == 3 { // Found a pair
				localSegments = append(localSegments, process(parts[0], depth+1)...)
				localSegments = append(localSegments, &widget.TextSegment{
					Text: parts[1],
					Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Italic: true}},
				})
				localSegments = append(localSegments, process(parts[2], depth+1)...)
				return localSegments
			}
		}
		
		// Links: [text](url)
		// Basic link detection, not exhaustive.
		// A full regex or more complex state machine is better.
		startLink := strings.Index(input, "[")
		if startLink != -1 {
			endLinkText := strings.Index(input[startLink:], "]")
			if endLinkText != -1 {
				endLinkText += startLink
				if len(input) > endLinkText+1 && input[endLinkText+1] == '(' {
					endUrl := strings.Index(input[endLinkText+2:], ")")
					if endUrl != -1 {
						endUrl += endLinkText + 2
						
						// Text before the link
						if startLink > 0 {
							localSegments = append(localSegments, &widget.TextSegment{Text: input[:startLink]})
						}
						
						linkText := input[startLink+1 : endLinkText]
						linkUrl := input[endLinkText+2 : endUrl]
						
						parsedURL, err := url.Parse(linkUrl)
						if err == nil {
							localSegments = append(localSegments, &widget.HyperlinkSegment{
								Text: linkText,
								URL:  parsedURL,
							})
						} else { // Fallback to text if URL is invalid
							localSegments = append(localSegments, &widget.TextSegment{Text: fmt.Sprintf("[%s](%s)", linkText, linkUrl)})
						}
						
						// Text after the link
						if endUrl < len(input)-1 {
							localSegments = append(localSegments, process(input[endUrl+1:], depth+1)...)
						}
						return localSegments
					}
				}
			}
		}


		// If no markdown found at this level, just add as plain text
		if input != "" {
			localSegments = append(localSegments, &widget.TextSegment{Text: input})
		}
		return localSegments
	}

	segments = process(current, 0)
	return segments
}


// getHeaderSize returns the appropriate size name for header level
func (p *Preview) getHeaderSize(level int) fyne.ThemeSizeName {
	switch level {
	case 1:
		return theme.SizeNameHeadingText
	case 2:
		return theme.SizeNameSubHeadingText
	// Add more cases if H3, H4, etc., should have distinct sizes.
	// Fyne's default theme may not define more than Heading and SubHeading for RichText by default.
	default:
		// For H3 and below, we can use bold text, or make them slightly larger than normal text
		// if the theme supports it or by explicitly setting font size (more complex).
		// Returning SizeNameText makes them regular size but they'll still be bolded by the TextStyle.
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
	// Basic HTML template with styling
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Markdown Export</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        article {
            background-color: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1, h2, h3, h4, h5, h6 {
            margin-top: 24px;
            margin-bottom: 16px;
            font-weight: 600;
            line-height: 1.25;
        }
        h1 { font-size: 2em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
        h2 { font-size: 1.5em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
        h3 { font-size: 1.25em; }
        h4 { font-size: 1em; }
        h5 { font-size: 0.875em; }
        h6 { font-size: 0.85em; color: #666; }
        code {
            background-color: #f0f0f0; /* Slightly darker for better visibility */
            padding: 0.2em 0.4em;
            margin: 0;
            font-size: 85%;
            border-radius: 3px;
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
        }
        pre {
            background-color: #f6f8fa;
            padding: 16px;
            border-radius: 6px;
            overflow-x: auto;
            line-height: 1.45;
            font-size: 0.9em; /* Match GitHub's pre font-size relative to body */
        }
        pre code {
            background-color: transparent;
            padding: 0;
            margin: 0;
            font-size: 100%; /* Code inside pre should inherit pre's font size */
            border-radius: 0;
            border: 0;
        }
        blockquote {
            margin: 1em 0;
            padding: 0 1em;
            color: #57606a; /* GitHub's blockquote color */
            border-left: 0.25em solid #d0d7de; /* GitHub's blockquote border */
        }
        a {
            color: #0969da;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        ul, ol {
            padding-left: 2em;
            margin-top: 0;
            margin-bottom: 16px;
        }
        li {
            margin-bottom: 0.25em;
        }
        hr {
            border: 0;
            height: 0.25em; /* Thicker hr like GitHub */
            background: #d0d7de; /* GitHub's hr color */
            margin: 24px 0;
        }
        img {
            max-width: 100%;
            height: auto;
            margin-top: 8px;
            margin-bottom: 8px;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 16px 0;
            display: block; /* For overflow */
            overflow-x: auto; /* For wide tables */
        }
        table th, table td {
            border: 1px solid #d0d7de; /* GitHub's table border color */
            padding: 6px 13px; /* GitHub's table padding */
            text-align: left;
        }
        table th {
            background-color: #f6f8fa;
            font-weight: 600;
        }
        p {
             margin-top: 0;
             margin-bottom: 16px;
        }
    </style>
</head>
<body>
    <article>
        %s
    </article>
</body>
</html>`

	// Convert markdown to HTML
	html := p.convertMarkdownToHTML(p.rawMarkdown)

	return fmt.Sprintf(htmlTemplate, html)
}

// escapeHTML escapes HTML special characters minimally for content.
// Note: Attribute values need more careful escaping (e.g., quotes).
func escapeHTMLContent(text string) string {
	text = strings.ReplaceAll(text, "&", "&") // Must be first
	text = strings.ReplaceAll(text, "<", "<")
	text = strings.ReplaceAll(text, ">", ">")
	return text
}

// escapeHTMLAttribute escapes characters for safe use in HTML attribute values.
func escapeHTMLAttribute(text string) string {
	text = escapeHTMLContent(text) // Basic escaping
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "'") // or ' but ' is more widely supported
	return text
}


// convertMarkdownToHTML performs basic markdown to HTML conversion
func (p *Preview) convertMarkdownToHTML(markdown string) string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	var html strings.Builder
	inCodeBlock := false
	var codeBlockLang string
	inList := false // Generic list flag
	listType := ""  // "ul" or "ol"
	inParagraph := false

	closeParagraph := func() {
		if inParagraph {
			html.WriteString("</p>\n")
			inParagraph = false
		}
	}
	
	ensureParagraph := func() {
		if !inParagraph && !inCodeBlock && listType == "" { // Don't start <p> inside lists or code blocks
			html.WriteString("<p>")
			inParagraph = true
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code blocks (```)
		if strings.HasPrefix(line, "```") { // Use `line` not `trimmed` to preserve indentation for potential languages
			closeParagraph()
			if inList {
				html.WriteString(fmt.Sprintf("</%s>\n", listType))
				inList = false
				listType = ""
			}
			if inCodeBlock {
				html.WriteString("</code></pre>\n")
				inCodeBlock = false
				codeBlockLang = ""
			} else {
				codeBlockLang = strings.TrimPrefix(line, "```")
				langClass := ""
				if codeBlockLang != "" {
					langClass = fmt.Sprintf(` class="language-%s"`, escapeHTMLAttribute(codeBlockLang))
				}
				html.WriteString(fmt.Sprintf("<pre><code%s>", langClass))
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			html.WriteString(escapeHTMLContent(line) + "\n")
			continue
		}
		
		// Process empty lines
		if trimmed == "" {
			closeParagraph()
			if inList { // Empty line might terminate a list item or the list
				html.WriteString(fmt.Sprintf("</%s>\n", listType))
				inList = false
				listType = ""
			}
			// Add a blank line as a paragraph break if needed, or let CSS handle margin
			// html.WriteString("<br>\n") // Or handle paragraph spacing with CSS
			continue
		}


		// Headers (H1-H6)
		if strings.HasPrefix(trimmed, "#") {
			closeParagraph()
			if inList {
				html.WriteString(fmt.Sprintf("</%s>\n", listType)); inList = false; listType = ""
			}
			level := 0
			textStart := 0
			for j, r := range trimmed {
				if r == '#' {
					level++
				} else if r == ' ' {
					textStart = j + 1
					break
				} else { // Not a header if # is not followed by space (or EOL)
					level = 0 
					break
				}
			}
			if textStart == 0 && level > 0 { // Handles cases like "###Header" without space
				textStart = level
			}


			if level > 0 && level <= 6 {
				headerText := p.processInlineHTML(trimmed[textStart:])
				html.WriteString(fmt.Sprintf("<h%d>%s</h%d>\n", level, headerText, level))
				continue
			}
		}

		// Unordered Lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			closeParagraph()
			if !inList || listType != "ul" {
				if inList { // Close previous list if different type
					html.WriteString(fmt.Sprintf("</%s>\n", listType))
				}
				html.WriteString("<ul>\n")
				inList = true
				listType = "ul"
			}
			html.WriteString(fmt.Sprintf("<li>%s</li>\n", p.processInlineHTML(trimmed[2:])))
			continue
		}
		// Ordered Lists (simple check: 1. item)
		if len(trimmed) > 2 && trimmed[1] == '.' && trimmed[0] >= '0' && trimmed[0] <= '9' {
			isOrderedList := true
			numStr := ""
			for k, char := range trimmed {
				if char >= '0' && char <= '9' {
					numStr += string(char)
				} else if char == '.' && k < len(trimmed)-1 && trimmed[k+1] == ' ' {
					// Check for " ." to confirm it's part of list marker
					break 
				} else {
					isOrderedList = false
					break
				}
			}
			if isOrderedList {
				numLen := len(numStr)
				if len(trimmed) > numLen+1 && trimmed[numLen] == '.' && trimmed[numLen+1] == ' ' {
					closeParagraph()
					if !inList || listType != "ol" {
						if inList {
							html.WriteString(fmt.Sprintf("</%s>\n", listType))
						}
						html.WriteString("<ol>\n") // Could add 'start' attribute if numStr != "1"
						inList = true
						listType = "ol"
					}
					html.WriteString(fmt.Sprintf("<li>%s</li>\n", p.processInlineHTML(trimmed[numLen+2:])))
					continue
				}
			}
		}
		
		// If current line is not a list item but we were in a list, close it.
		if inList && !(strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || (len(trimmed) > 2 && trimmed[1] == '.' && trimmed[0] >= '0' && trimmed[0] <= '9')) {
			html.WriteString(fmt.Sprintf("</%s>\n", listType))
			inList = false
			listType = ""
		}


		// Blockquotes
		if strings.HasPrefix(trimmed, "> ") {
			closeParagraph()
			if inList {
				html.WriteString(fmt.Sprintf("</%s>\n", listType)); inList = false; listType = ""
			}
			// Recursively handle multiple > signs for nested blockquotes if desired.
			// For now, simple blockquote.
			html.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", p.processInlineHTML(trimmed[2:])))
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			closeParagraph()
			if inList {
				html.WriteString(fmt.Sprintf("</%s>\n", listType)); inList = false; listType = ""
			}
			html.WriteString("<hr>\n")
			continue
		}
		
		// Default to paragraph if none of the above
		// The ensureParagraph will create a <p> if not already in one.
		// Image handling and inline HTML processing is done by p.processInlineHTML now.
		ensureParagraph()
		html.WriteString(p.processInlineHTML(line))
		if i < len(lines)-1 { // Add a space to allow next line's content to flow, or <br> if strict line breaks are desired
		    html.WriteString(" ") // Or remove, and let paragraphs handle breaks.
		}


	}

	closeParagraph() // Close any dangling paragraph
	if inList { // Close any dangling list
		html.WriteString(fmt.Sprintf("</%s>\n", listType))
	}

	return html.String()
}


// convertImageToHTML converts markdown image syntax to HTML
func (p *Preview) convertImageToHTML(altText, imageURL string) string {
	return fmt.Sprintf(`<img src="%s" alt="%s">`, escapeHTMLAttribute(imageURL), escapeHTMLAttribute(altText))
}

// processInlineHTML handles inline markdown elements for HTML conversion
func (p *Preview) processInlineHTML(text string) string {
    // Order of replacement matters: Strong before Em, and escape HTML first.
    // This is a simplified parser. A regex-based or proper tokenizer/parser would be more robust.
    
    // Phase 1: Handle images ![alt](url)
    // This needs to run before other transformations that might break the image syntax.
    var processedText strings.Builder
    lastIndex := 0
    for {
        imgTagStart := strings.Index(text[lastIndex:], "![")
        if imgTagStart == -1 {
            processedText.WriteString(text[lastIndex:])
            break
        }
        imgTagStart += lastIndex // Adjust to absolute index in `text`

        altEnd := strings.Index(text[imgTagStart+2:], "]")
        if altEnd == -1 {
            processedText.WriteString(text[lastIndex:]) // Malformed, write rest and stop
            break
        }
        altEnd += imgTagStart + 2 // Absolute index of ']'

        if altEnd+1 >= len(text) || text[altEnd+1] != '(' {
            processedText.WriteString(text[lastIndex:altEnd+1]) // Not an image link, write up to ']' and continue
            lastIndex = altEnd + 1
            continue
        }
        
        urlEnd := strings.Index(text[altEnd+2:], ")")
        if urlEnd == -1 {
            processedText.WriteString(text[lastIndex:]) // Malformed, write rest and stop
            break
        }
        urlEnd += altEnd + 2 // Absolute index of ')'

        altText := text[imgTagStart+2 : altEnd]
        imageURL := text[altEnd+2 : urlEnd]

        processedText.WriteString(text[lastIndex:imgTagStart]) // Text before image
        processedText.WriteString(p.convertImageToHTML(altText, imageURL)) // Image HTML
        lastIndex = urlEnd + 1
    }
    text = processedText.String()


    // Phase 2: Handle links [text](url) - must be after images to avoid conflict
    processedText.Reset()
    lastIndex = 0
    for {
        linkStart := strings.Index(text[lastIndex:], "[")
        if linkStart == -1 {
            processedText.WriteString(text[lastIndex:])
            break
        }
        linkStart += lastIndex

        linkTextEnd := strings.Index(text[linkStart+1:], "]")
        if linkTextEnd == -1 {
            processedText.WriteString(text[lastIndex:])
            break
        }
        linkTextEnd += linkStart + 1

        if linkTextEnd+1 >= len(text) || text[linkTextEnd+1] != '(' {
            processedText.WriteString(text[lastIndex : linkTextEnd+1])
            lastIndex = linkTextEnd + 1
            continue
        }
        
        urlContentEnd := strings.Index(text[linkTextEnd+2:], ")")
        if urlContentEnd == -1 {
            processedText.WriteString(text[lastIndex:])
            break
        }
        urlContentEnd += linkTextEnd + 2

        linkText := text[linkStart+1 : linkTextEnd]
        url := text[linkTextEnd+2 : urlContentEnd]

        processedText.WriteString(escapeHTMLContent(text[lastIndex:linkStart]))
        // The linkText itself might contain inline markdown (e.g., bold text in link)
        // For simplicity, we are escaping it directly here. A recursive call would be needed for nested markdown.
        processedText.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, escapeHTMLAttribute(url), p.processInlineHTML(linkText))) // Recursive call for linkText
        lastIndex = urlContentEnd + 1
    }
    text = processedText.String()

    // Phase 3: Handle strong, em, code after links and images are converted to HTML tags
    // to prevent their content from being markdown-processed.
    // This simple replace approach is fragile for nested/overlapping cases.
    // A proper parser would tokenize based on delimiters.
    
    // Temporarily replace HTML tags to protect them from markdown processing
    // This is a common trick but has its limitations
    tempTags := make(map[string]string)
    tagCounter := 0
    protectTag := func(htmlTag string) string {
        placeholder := fmt.Sprintf("HTMLTAG%dPLACEHOLDER", tagCounter)
        tagCounter++
        tempTags[placeholder] = htmlTag
        return placeholder
    }
    
    // Protect existing a and img tags generated above
    text = strings.ReplaceAll(text, "</a>", protectTag("</a>"))
    // Regex would be better for <a href="..."> and <img src="...">
    // Simple string replace for now - this is not robust
    // This protection step becomes very complex with attributes.

    text = escapeHTMLContent(text) // Escape remaining text *before* applying markdown formatting tags
    
    // Bold (matches __text__ or **text**)
    text = replacePattern(text, "**", "**", "<strong>", "</strong>", false)
    text = replacePattern(text, "__", "__", "<strong>", "</strong>", false)
    // Italic (matches _text_ or *text*)
    text = replacePattern(text, "*", "*", "<em>", "</em>", false)
    text = replacePattern(text, "_", "_", "<em>", "</em>", false)
    // Inline Code (matches `text`)
    text = replacePattern(text, "`", "`", "<code>", "</code>", true) // `true` to not escape content inside code


    // Restore protected HTML tags
    for placeholder, originalTag := range tempTags {
        text = strings.ReplaceAll(text, placeholder, originalTag)
    }

    return text
}

// replacePattern replaces markdown patterns with HTML.
// `contentAlreadyEscaped` indicates if the content between delimiters should avoid double-escaping.
// `isCode` means the content should not be processed for further markdown.
func replacePattern(text, startDelim, endDelim, startTag, endTag string, isCode bool) string {
	var result strings.Builder
	lastIndex := 0

	for {
		start := strings.Index(text[lastIndex:], startDelim)
		if start == -1 {
			result.WriteString(text[lastIndex:])
			break
		}
		start += lastIndex // Adjust to absolute index

		// Write text before the match
		result.WriteString(text[lastIndex:start])

		// Find end delimiter
		afterStartDelim := start + len(startDelim)
		end := strings.Index(text[afterStartDelim:], endDelim)
		if end == -1 { // No closing delimiter, treat startDelim as literal
			result.WriteString(text[start:]) // Write from startDelim to end of string
			lastIndex = len(text)
			break
		}
		end += afterStartDelim // Adjust to absolute index

		content := text[afterStartDelim:end]
        // If it's not code, content might need further processing or is already (partially) HTML.
        // If it IS code, we want the raw content without further HTML escaping.
        // The `escapeHTMLContent` in `processInlineHTML` has already escaped the bulk.
        // Here, we are wrapping, so content should generally be passed through.

		result.WriteString(startTag)
        result.WriteString(content) // Content is assumed to be appropriately handled/escaped before this function for non-code
		result.WriteString(endTag)

		lastIndex = end + len(endDelim)
	}
	return result.String()
}