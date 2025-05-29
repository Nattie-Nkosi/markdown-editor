package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Preview represents the markdown preview component
type Preview struct {
	content   *widget.RichText
	container *fyne.Container
	visible   bool
}

// NewPreview creates a new preview instance
func NewPreview() *Preview {
	p := &Preview{
		content: widget.NewRichTextFromMarkdown(""),
		visible: true,
	}
	
	return p
}

// Create creates the preview UI component
func (p *Preview) Create() fyne.CanvasObject {
	scrollContainer := container.NewScroll(p.content)
	p.container = container.NewBorder(
		widget.NewCard("Preview", "", nil),
		nil, nil, nil,
		scrollContainer,
	)
	return p.container
}

// UpdateContent updates the preview with new markdown content
func (p *Preview) UpdateContent(markdown string) {
	p.content.ParseMarkdown(markdown)
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
        h2 { font-size: 1.5em; }
        h3 { font-size: 1.25em; }
        code {
            background-color: #f6f8fa;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
        }
        pre {
            background-color: #f6f8fa;
            padding: 16px;
            border-radius: 6px;
            overflow-x: auto;
        }
        pre code {
            background-color: transparent;
            padding: 0;
        }
        blockquote {
            margin: 0;
            padding: 0 1em;
            color: #666;
            border-left: 4px solid #ddd;
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
        }
        hr {
            border: 0;
            height: 1px;
            background: #e1e4e8;
            margin: 24px 0;
        }
        img {
            max-width: 100%;
            height: auto;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 16px 0;
        }
        table th, table td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        table th {
            background-color: #f6f8fa;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <article>
        %s
    </article>
</body>
</html>`
	
	// Convert markdown to HTML (basic conversion)
	// In a real application, you'd use a proper markdown-to-HTML library
	html := p.convertMarkdownToHTML(p.content.String())
	
	return fmt.Sprintf(htmlTemplate, html)
}

// convertMarkdownToHTML performs basic markdown to HTML conversion
// Note: This is a simplified version. In production, use a proper markdown parser
func (p *Preview) convertMarkdownToHTML(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var html strings.Builder
	inCodeBlock := false
	inList := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Code blocks
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				html.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				html.WriteString("<pre><code>")
				inCodeBlock = true
			}
			continue
		}
		
		if inCodeBlock {
			html.WriteString(escapeHTML(line) + "\n")
			continue
		}
		
		// Headers
		if strings.HasPrefix(trimmed, "# ") {
			html.WriteString(fmt.Sprintf("<h1>%s</h1>\n", escapeHTML(trimmed[2:])))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			html.WriteString(fmt.Sprintf("<h2>%s</h2>\n", escapeHTML(trimmed[3:])))
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			html.WriteString(fmt.Sprintf("<h3>%s</h3>\n", escapeHTML(trimmed[4:])))
			continue
		}
		
		// Lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				html.WriteString("<ul>\n")
				inList = true
			}
			html.WriteString(fmt.Sprintf("<li>%s</li>\n", escapeHTML(trimmed[2:])))
			continue
		} else if inList && trimmed == "" {
			html.WriteString("</ul>\n")
			inList = false
		}
		
		// Blockquotes
		if strings.HasPrefix(trimmed, "> ") {
			html.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", escapeHTML(trimmed[2:])))
			continue
		}
		
		// Horizontal rule
		if trimmed == "---" || trimmed == "***" {
			html.WriteString("<hr>\n")
			continue
		}
		
		// Paragraphs
		if trimmed != "" {
			html.WriteString(fmt.Sprintf("<p>%s</p>\n", p.processInlineMarkdown(trimmed)))
		}
	}
	
	// Close any open lists
	if inList {
		html.WriteString("</ul>\n")
	}
	
	return html.String()
}

// processInlineMarkdown handles inline markdown elements
func (p *Preview) processInlineMarkdown(text string) string {
	// Bold
	text = replacePattern(text, "**", "**", "<strong>", "</strong>")
	text = replacePattern(text, "__", "__", "<strong>", "</strong>")
	
	// Italic
	text = replacePattern(text, "*", "*", "<em>", "</em>")
	text = replacePattern(text, "_", "_", "<em>", "</em>")
	
	// Code
	text = replacePattern(text, "`", "`", "<code>", "</code>")
	
	// Links - basic pattern
	// This is simplified - a full implementation would use regex
	if strings.Contains(text, "](") {
		// Basic link replacement
		text = strings.ReplaceAll(text, "[", "<a href='#'>")
		text = strings.ReplaceAll(text, "]", "</a>")
	}
	
	return escapeHTML(text)
}

// replacePattern replaces markdown patterns with HTML
func replacePattern(text, startDelim, endDelim, startTag, endTag string) string {
	result := text
	for {
		start := strings.Index(result, startDelim)
		if start == -1 {
			break
		}
		
		afterStart := start + len(startDelim)
		end := strings.Index(result[afterStart:], endDelim)
		if end == -1 {
			break
		}
		
		end += afterStart
		content := result[afterStart:end]
		replacement := startTag + content + endTag
		
		result = result[:start] + replacement + result[end+len(endDelim):]
	}
	return result
}

// escapeHTML escapes HTML special characters
func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}