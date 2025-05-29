package main

import (
	"io/ioutil"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type config struct {
	EditWidget    *widget.Entry
	PreviewWidget *widget.RichText
	CurrentFile   fyne.URI
	SaveMenuItem  *fyne.MenuItem
}

var cfg config

func main() {
	// create a fyne app
	a := app.New()

	// create a window for the app
	win := a.NewWindow("Markdown Editor")

	// get the user interface
	edit, preview := cfg.makeUI()

	// create the main menu
	cfg.createMenuItems(win)

	// set the content of the window
	win.SetContent(container.NewHSplit(edit, preview))

	// show window and run app
	win.Resize(fyne.Size{Width: 800, Height: 500})
	win.CenterOnScreen()
	win.ShowAndRun()
}

func (app *config) makeUI() (*widget.Entry, *widget.RichText) {
	edit := widget.NewMultiLineEntry()
	preview := widget.NewRichTextFromMarkdown("")
	app.EditWidget = edit
	app.PreviewWidget = preview

	edit.OnChanged = preview.ParseMarkdown

	return edit, preview
}

func (app *config) createMenuItems(win fyne.Window) {
	// File menu
	openMenuItem := fyne.NewMenuItem("Open...", app.openFunc(win))
	saveMenuItem := fyne.NewMenuItem("Save", app.saveFunc(win))
	app.SaveMenuItem = saveMenuItem
	app.SaveMenuItem.Disabled = true
	saveAsMenuItem := fyne.NewMenuItem("Save As...", app.saveAsFunc(win))

	fileMenu := fyne.NewMenu("File",
		openMenuItem,
		saveMenuItem,
		saveAsMenuItem,
	)

	// Edit menu
	cutMenuItem := fyne.NewMenuItem("Cut", func() {
		app.EditWidget.TypedShortcut(&fyne.ShortcutCut{})
	})
	copyMenuItem := fyne.NewMenuItem("Copy", func() {
		app.EditWidget.TypedShortcut(&fyne.ShortcutCopy{})
	})
	pasteMenuItem := fyne.NewMenuItem("Paste", func() {
		app.EditWidget.TypedShortcut(&fyne.ShortcutPaste{})
	})

	editMenu := fyne.NewMenu("Edit",
		cutMenuItem,
		copyMenuItem,
		pasteMenuItem,
	)

	// View menu
	togglePreviewMenuItem := fyne.NewMenuItem("Toggle Preview", func() {
		if app.PreviewWidget.Visible() {
			app.PreviewWidget.Hide()
		} else {
			app.PreviewWidget.Show()
		}
	})

	viewMenu := fyne.NewMenu("View",
		togglePreviewMenuItem,
	)

	// Insert menu
	insertMenu := fyne.NewMenu("Insert",
		fyne.NewMenuItem("Bold", func() {
			app.insertMarkdown("**", "**")
		}),
		fyne.NewMenuItem("Italic", func() {
			app.insertMarkdown("*", "*")
		}),
		fyne.NewMenuItem("Code", func() {
			app.insertMarkdown("`", "`")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Link", func() {
			app.insertMarkdown("[", "](url)")
		}),
		fyne.NewMenuItem("Image", func() {
			app.insertMarkdown("![alt text](", ")")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Heading 1", func() {
			app.insertAtLineStart("# ")
		}),
		fyne.NewMenuItem("Heading 2", func() {
			app.insertAtLineStart("## ")
		}),
		fyne.NewMenuItem("Heading 3", func() {
			app.insertAtLineStart("### ")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Bullet List", func() {
			app.insertAtLineStart("- ")
		}),
		fyne.NewMenuItem("Numbered List", func() {
			app.insertAtLineStart("1. ")
		}),
		fyne.NewMenuItem("Quote", func() {
			app.insertAtLineStart("> ")
		}),
		fyne.NewMenuItem("Code Block", func() {
			app.insertMarkdown("```\n", "\n```")
		}),
	)

	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About", "Markdown Editor\nBuilt with Fyne\nVersion 1.0", win)
		}),
	)

	mainMenu := fyne.NewMainMenu(
		fileMenu,
		editMenu,
		viewMenu,
		insertMenu,
		helpMenu,
	)

	win.SetMainMenu(mainMenu)
}

// File operations
func (app *config) openFunc(win fyne.Window) func() {
	return func() {
		openDialog := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if read == nil {
				return
			}
			defer read.Close()

			data, err := ioutil.ReadAll(read)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			app.EditWidget.SetText(string(data))
			app.CurrentFile = read.URI()
			app.SaveMenuItem.Disabled = false
			win.SetTitle("Markdown Editor - " + read.URI().Name())
		}, win)

		openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown"}))
		openDialog.Show()
	}
}

func (app *config) saveFunc(win fyne.Window) func() {
	return func() {
		if app.CurrentFile != nil {
			write, err := storage.Writer(app.CurrentFile)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			defer write.Close()

			write.Write([]byte(app.EditWidget.Text))
		}
	}
}

func (app *config) saveAsFunc(win fyne.Window) func() {
	return func() {
		saveDialog := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if write == nil {
				return
			}
			defer write.Close()

			write.Write([]byte(app.EditWidget.Text))
			app.CurrentFile = write.URI()
			app.SaveMenuItem.Disabled = false
			win.SetTitle("Markdown Editor - " + write.URI().Name())
		}, win)

		saveDialog.SetFileName("untitled.md")
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown"}))
		saveDialog.Show()
	}
}

// Helper functions for markdown insertion
func (app *config) insertMarkdown(before, after string) {
	if app.EditWidget.SelectedText() == "" {
		// No selection, insert at cursor
		pos := app.EditWidget.CursorColumn
		text := app.EditWidget.Text
		newText := text[:pos] + before + "text" + after + text[pos:]
		app.EditWidget.SetText(newText)
		app.EditWidget.CursorColumn = pos + len(before)
	} else {
		// Wrap selection
		selectedText := app.EditWidget.SelectedText()
		app.EditWidget.TypedRune(0) // This clears the selection
		pos := app.EditWidget.CursorColumn
		text := app.EditWidget.Text
		newText := text[:pos] + before + selectedText + after + text[pos:]
		app.EditWidget.SetText(newText)
	}
}

func (app *config) insertAtLineStart(prefix string) {
	// Calculate cursor position from the start of text
	lines := strings.Split(app.EditWidget.Text[:app.EditWidget.CursorRow], "\n")
	pos := 0
	for _, line := range lines[:len(lines)-1] {
		pos += len(line) + 1 // +1 for newline
	}
	pos += app.EditWidget.CursorColumn
	
	text := app.EditWidget.Text
	
	// Find start of current line
	lineStart := strings.LastIndex(text[:pos], "\n") + 1
	
	// Insert prefix at line start
	newText := text[:lineStart] + prefix + text[lineStart:]
	app.EditWidget.SetText(newText)
	app.EditWidget.CursorRow = app.EditWidget.CursorRow
	app.EditWidget.CursorColumn = app.EditWidget.CursorColumn + len(prefix)
}