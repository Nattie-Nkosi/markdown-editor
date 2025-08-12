package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Editor represents the text editor component
type Editor struct {
	controller *AppController
	entry      *widget.Entry
	container  *fyne.Container
}

// NewEditor creates a new editor instance
func NewEditor(controller *AppController) *Editor {
	e := &Editor{
		controller: controller,
		entry:      widget.NewMultiLineEntry(),
	}

	e.entry.PlaceHolder = "Start typing your markdown here..."
	e.entry.OnChanged = func(content string) {
		controller.OnTextChanged(content)
	}

	// Enable word wrap
	e.entry.Wrapping = fyne.TextWrapWord

	return e
}

// Create creates the editor UI component
func (e *Editor) Create() fyne.CanvasObject {
	scrollContainer := container.NewScroll(e.entry)
	e.container = container.NewBorder(nil, nil, nil, nil, scrollContainer)
	return e.container
}

// SetContent sets the editor content
func (e *Editor) SetContent(content string) {
	e.entry.SetText(content)
}

// GetContent returns the editor content
func (e *Editor) GetContent() string {
	return e.entry.Text
}

// InsertMarkdown inserts markdown syntax around selected text or at cursor
func (e *Editor) InsertMarkdown(before, after string, placeholder string) {
	// Get current position (approximate using selection)
	if e.entry.SelectedText() != "" {
		// Wrap selected text
		newText := before + e.entry.SelectedText() + after
		e.entry.TypedRune([]rune(newText)[0])
		for _, r := range newText[1:] {
			e.entry.TypedRune(r)
		}
	} else {
		// Insert with placeholder
		insertText := before + placeholder + after
		// Get current content and append
		currentText := e.entry.Text
		e.entry.SetText(currentText + insertText)
		// Try to position cursor between markers
		if placeholder != "" {
			// Select the placeholder text for easy replacement
			startPos := len(currentText) + len(before)
			// Note: Fyne doesn't have direct selection API, but we can approximate
			e.entry.CursorRow = strings.Count(currentText[:startPos], "\n")
		}
	}
}

// InsertAtLineStart inserts text at the beginning of the current line
func (e *Editor) InsertAtLineStart(prefix string) {
	text := e.entry.Text
	
	// Simple approach: just insert the prefix
	// Since we can't get exact cursor position, append to current text
	if text != "" && !strings.HasSuffix(text, "\n") {
		e.entry.SetText(text + "\n" + prefix)
	} else {
		e.entry.SetText(text + prefix)
	}
}

// ShowFindDialog shows the find dialog
func (e *Editor) ShowFindDialog() {
	findEntry := widget.NewEntry()
	findEntry.PlaceHolder = "Find text..."
	
	resultLabel := widget.NewLabel("")
	
	content := container.NewVBox(
		widget.NewLabel("Find:"),
		findEntry,
		resultLabel,
	)
	
	d := dialog.NewCustom("Find", "Close", content, e.controller.window)
	
	lastIndex := 0
	
	findNext := func() {
		searchText := findEntry.Text
		if searchText == "" {
			resultLabel.SetText("Enter search text")
			return
		}
		
		text := strings.ToLower(e.entry.Text)
		search := strings.ToLower(searchText)
		
		index := strings.Index(text[lastIndex:], search)
		if index >= 0 {
			foundAt := lastIndex + index
			lastIndex = foundAt + 1
			
			// Count line number
			lineNum := strings.Count(text[:foundAt], "\n") + 1
			resultLabel.SetText("Found at line " + string(rune(lineNum+'0')))
		} else if lastIndex > 0 {
			// Try from beginning
			lastIndex = 0
			index = strings.Index(text, search)
			if index >= 0 {
				lineNum := strings.Count(text[:index], "\n") + 1
				resultLabel.SetText("Found at line " + string(rune(lineNum+'0')) + " (wrapped)")
				lastIndex = index + 1
			} else {
				resultLabel.SetText("Not found")
			}
		} else {
			resultLabel.SetText("Not found")
		}
	}
	
	findEntry.OnSubmitted = func(s string) { findNext() }
	
	d.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Find Next", findNext),
		widget.NewButton("Close", d.Hide),
	})
	
	d.Resize(fyne.NewSize(300, 150))
	d.Show()
}

// ShowReplaceDialog shows the find and replace dialog
func (e *Editor) ShowReplaceDialog() {
	findEntry := widget.NewEntry()
	findEntry.PlaceHolder = "Find text..."
	
	replaceEntry := widget.NewEntry()
	replaceEntry.PlaceHolder = "Replace with..."
	
	resultLabel := widget.NewLabel("")
	
	content := container.NewVBox(
		widget.NewLabel("Find:"),
		findEntry,
		widget.NewLabel("Replace:"),
		replaceEntry,
		resultLabel,
	)
	
	d := dialog.NewCustom("Find and Replace", "Close", content, e.controller.window)
	
	replaceOne := func() {
		find := findEntry.Text
		replace := replaceEntry.Text
		if find == "" {
			resultLabel.SetText("Enter search text")
			return
		}
		
		text := e.entry.Text
		if strings.Contains(text, find) {
			newText := strings.Replace(text, find, replace, 1)
			e.entry.SetText(newText)
			resultLabel.SetText("Replaced 1 occurrence")
		} else {
			resultLabel.SetText("Not found")
		}
	}
	
	replaceAll := func() {
		find := findEntry.Text
		replace := replaceEntry.Text
		if find == "" {
			resultLabel.SetText("Enter search text")
			return
		}
		
		text := e.entry.Text
		count := strings.Count(text, find)
		if count > 0 {
			newText := strings.ReplaceAll(text, find, replace)
			e.entry.SetText(newText)
			resultLabel.SetText("Replaced " + string(rune(count+'0')) + " occurrences")
		} else {
			resultLabel.SetText("Not found")
		}
	}
	
	d.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Replace", replaceOne),
		widget.NewButton("Replace All", replaceAll),
		widget.NewButton("Close", d.Hide),
	})
	
	d.Resize(fyne.NewSize(350, 200))
	d.Show()
}

// Focus sets focus to the editor
func (e *Editor) Focus() {
	e.entry.FocusGained()
}