package main

import (
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Preview represents the markdown preview component
type Preview struct {
    content     *widget.RichText
    container   *fyne.Container
    visible     bool
    rawMarkdown string
    md          goldmark.Markdown
}

// NewPreview creates a new preview instance
func NewPreview() *Preview {
	// Configure goldmark with common extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // GitHub Flavored Markdown (tables, strikethrough, etc.)
			extension.Typographer,   // Smart quotes and dashes
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML
		),
	)

	p := &Preview{
		content: widget.NewRichTextFromMarkdown(""),
		visible: true,
		md:      md,
	}
	p.content.Wrapping = fyne.TextWrapWord
	return p
}

// Create creates the preview UI component
func (p *Preview) Create() fyne.CanvasObject {
	// Simple container with the RichText widget
	scrollContainer := container.NewScroll(
		container.NewPadded(p.content),
	)
	
	// Card header for the preview
	header := widget.NewCard("", "Preview", nil)
	
	p.container = container.NewBorder(
		container.NewPadded(header),
		nil, nil, nil,
		scrollContainer,
	)
	return p.container
}

// UpdateContent updates the preview with new markdown content
func (p *Preview) UpdateContent(markdown string) {
	p.rawMarkdown = markdown
	
	var buf bytes.Buffer
	if err := p.md.Convert([]byte(markdown), &buf); err != nil {
		buf.WriteString(fmt.Sprintf("<p>Error converting markdown: %v</p>", err))
	}
	p.content.ParseMarkdown(buf.String())
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

// GetHTML returns the markdown converted to HTML using goldmark
func (p *Preview) GetHTML() string {
	var buf bytes.Buffer
	
	// Convert markdown to HTML using goldmark
	if err := p.md.Convert([]byte(p.rawMarkdown), &buf); err != nil {
		return fmt.Sprintf("<p>Error converting markdown: %v</p>", err)
	}
	
	// Wrap in a simple HTML template with basic styling
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Markdown Export</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            font-size: 16px;
            line-height: 1.6;
            color: #333;
            background-color: #fff;
            margin: 0;
            padding: 20px;
            max-width: 800px;
            margin: 0 auto;
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
        h6 { font-size: 0.85em; color: #777; }
        p { margin-top: 0; margin-bottom: 16px; }
        a { color: #0969da; text-decoration: none; }
        a:hover { text-decoration: underline; }
        code {
            padding: 0.2em 0.4em;
            margin: 0;
            font-size: 85%%;
            background-color: rgba(27,31,35,0.05);
            border-radius: 3px;
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }
        pre {
            padding: 16px;
            overflow: auto;
            font-size: 85%%;
            line-height: 1.45;
            background-color: #f6f8fa;
            border-radius: 3px;
        }
        pre code {
            display: inline;
            padding: 0;
            margin: 0;
            border: 0;
            background-color: transparent;
        }
        blockquote {
            padding: 0 1em;
            color: #6a737d;
            border-left: 0.25em solid #dfe2e5;
            margin: 0 0 16px 0;
        }
        ul, ol {
            padding-left: 2em;
            margin-top: 0;
            margin-bottom: 16px;
        }
        li { margin-bottom: 0.25em; }
        table {
            border-spacing: 0;
            border-collapse: collapse;
            margin-bottom: 16px;
        }
        table th, table td {
            padding: 6px 13px;
            border: 1px solid #dfe2e5;
        }
        table th {
            font-weight: 600;
            background-color: #f6f8fa;
        }
        table tr {
            background-color: #fff;
            border-top: 1px solid #c6cbd1;
        }
        table tr:nth-child(2n) {
            background-color: #f6f8fa;
        }
        hr {
            height: 0.25em;
            padding: 0;
            margin: 24px 0;
            background-color: #e1e4e8;
            border: 0;
        }
        img {
            max-width: 100%%;
            box-sizing: content-box;
        }
        .task-list-item {
            list-style-type: none;
        }
        .task-list-item input {
            margin: 0 0.2em 0.25em -1.6em;
            vertical-align: middle;
        }
    </style>
</head>
<body>
%s
</body>
</html>`
	
	return fmt.Sprintf(htmlTemplate, buf.String())
}