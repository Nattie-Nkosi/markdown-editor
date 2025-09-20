package main

import (
	"bytes"
	"fmt"
	"strings"

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
	content         *widget.RichText
	container       *fyne.Container
	scrollContainer *container.Scroll
	placeholder     fyne.CanvasObject
	visible         bool
	rawMarkdown     string
	rendered        string
	md              goldmark.Markdown
}

// NewPreview creates a new preview instance
func NewPreview() *Preview {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
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
	p.scrollContainer = container.NewScroll(
		container.NewPadded(p.content),
	)
	p.scrollContainer.Hide()

	placeholderLabel := widget.NewLabelWithStyle(
		"Live preview will appear here as you type.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	p.placeholder = container.NewCenter(placeholderLabel)

	title := widget.NewLabelWithStyle("Preview", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleContainer := container.NewBorder(nil, widget.NewSeparator(), nil, nil, container.NewPadded(title))

	body := container.NewMax(p.scrollContainer, p.placeholder)

	p.container = container.NewBorder(
		titleContainer,
		nil,
		nil,
		nil,
		body,
	)

	p.UpdateContent(p.rawMarkdown)
	return p.container
}

// UpdateContent updates the preview with new markdown content
func (p *Preview) UpdateContent(markdown string) {
	p.rawMarkdown = markdown

	if p.scrollContainer == nil || p.placeholder == nil {
		return
	}

	trimmed := strings.TrimSpace(markdown)
	if trimmed == "" {
		p.rendered = ""
		p.content.ParseMarkdown("")
		p.scrollContainer.Hide()
		p.placeholder.Show()
		return
	}

	if markdown == p.rendered {
		p.placeholder.Hide()
		p.scrollContainer.Show()
		return
	}

	p.placeholder.Hide()
	p.scrollContainer.Show()

	// Use Fyne's built-in markdown parsing for the preview
	p.content.ParseMarkdown(markdown)

	// Refresh the scroll container to ensure proper rendering
	p.scrollContainer.Refresh()
	p.rendered = markdown
}

// ToggleVisibility toggles the preview pane visibility
func (p *Preview) ToggleVisibility() {
	if p.container == nil {
		return
	}

	if p.visible {
		p.container.Hide()
	} else {
		p.container.Show()
	}
	p.visible = !p.visible
}

// GetHTML returns the markdown converted to HTML
func (p *Preview) GetHTML() string {
	var buf bytes.Buffer

	if err := p.md.Convert([]byte(p.rawMarkdown), &buf); err != nil {
		return fmt.Sprintf("<p>Error converting markdown: %v</p>", err)
	}

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
