package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Toolbar represents the application toolbar
type Toolbar struct {
	controller *AppController
}

// NewToolbar creates a new toolbar instance
func NewToolbar(controller *AppController) *Toolbar {
	return &Toolbar{
		controller: controller,
	}
}

// Create creates the toolbar UI component with both icon toolbar and button toolbar
func (t *Toolbar) Create() fyne.CanvasObject {
	// Main toolbar with icons
	iconToolbar := widget.NewToolbar(
		// File operations
		widget.NewToolbarAction(theme.DocumentCreateIcon(), t.controller.NewFile),
		widget.NewToolbarAction(theme.FolderOpenIcon(), t.controller.Open),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), t.controller.Save),
		widget.NewToolbarSeparator(),
		
		// Edit operations
		widget.NewToolbarAction(theme.ContentUndoIcon(), func() {
			t.controller.editor.entry.TypedShortcut(&fyne.ShortcutUndo{})
		}),
		widget.NewToolbarAction(theme.ContentRedoIcon(), func() {
			t.controller.editor.entry.TypedShortcut(&fyne.ShortcutRedo{})
		}),
		widget.NewToolbarSeparator(),
		
		// View and Export
		widget.NewToolbarAction(theme.VisibilityIcon(), t.controller.TogglePreview),
		widget.NewToolbarAction(theme.DocumentPrintIcon(), t.controller.ExportHTML),
	)
	
	// Text formatting buttons (since we don't have specific icons)
	formatButtons := container.NewHBox(
		widget.NewButton("B", func() {
			t.controller.InsertMarkdown("**", "**", "bold")
		}),
		widget.NewButton("I", func() {
			t.controller.InsertMarkdown("*", "*", "italic")
		}),
		widget.NewButton("S", func() {
			t.controller.InsertMarkdown("~~", "~~", "strikethrough")
		}),
		widget.NewButton("Code", func() {
			t.controller.InsertMarkdown("`", "`", "code")
		}),
		widget.NewSeparator(),
		widget.NewButton("Link", func() {
			t.controller.InsertMarkdown("[", "](url)", "link text")
		}),
		widget.NewButton("Image", func() {
			t.controller.InsertMarkdown("![", "](url)", "alt text")
		}),
		widget.NewSeparator(),
		widget.NewButton("H1", func() {
			t.controller.InsertAtLineStart("# ")
		}),
		widget.NewButton("H2", func() {
			t.controller.InsertAtLineStart("## ")
		}),
		widget.NewButton("H3", func() {
			t.controller.InsertAtLineStart("### ")
		}),
		widget.NewSeparator(),
		widget.NewButton("List", func() {
			t.controller.InsertAtLineStart("- ")
		}),
		widget.NewButton("Quote", func() {
			t.controller.InsertAtLineStart("> ")
		}),
		widget.NewButton("Code Block", func() {
			t.controller.InsertMarkdown("```\n", "\n```", "language")
		}),
	)
	
	// Combine both toolbars
	return container.NewVBox(
		iconToolbar,
		container.NewHScroll(formatButtons),
	)
}