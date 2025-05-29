package main

import (
	"fmt"
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
	selectedText := e.entry.SelectedText()
	
	if selectedText == "" {
		// No selection, insert at cursor with placeholder
		e.entry.TypedRune([]rune(before + placeholder + after)[0])
		for _, r := range (before + placeholder + after)[1:] {
			e.entry.TypedRune(r)
		}
	} else {
		// Replace selection with wrapped text
		replacement := before + selectedText + after
		e.entry.TypedRune([]rune(replacement)[0])
		for _, r := range replacement[1:] {
			e.entry.TypedRune(r)
		}
	}
}

// InsertAtLineStart inserts text at the beginning of the current line
func (e *Editor) InsertAtLineStart(prefix string) {
	// Get current text and selection
	text := e.entry.Text
	
	// Since we can't get cursor position directly, we'll work with the current selection
	// or insert at the end if there's no selection
	selectedText := e.entry.SelectedText()
	
	if selectedText != "" {
		// Find the start of the line containing the selection
		beforeSelection := strings.Split(text, selectedText)[0]
		lineStart := strings.LastIndex(beforeSelection, "\n") + 1
		
		// Check if line already starts with this prefix
		afterLineStart := text[lineStart:]
		lineEnd := strings.Index(afterLineStart, "\n")
		if lineEnd == -1 {
			lineEnd = len(afterLineStart)
		}
		
		currentLine := afterLineStart[:lineEnd]
		if strings.HasPrefix(strings.TrimSpace(currentLine), strings.TrimSpace(prefix)) {
			// Line already has this prefix, don't add another
			return
		}
		
		// Insert prefix at line start
		newText := text[:lineStart] + prefix + text[lineStart:]
		e.entry.SetText(newText)
	} else {
		// No selection, insert at current line
		// This is a simplified approach - just add the prefix
		e.entry.TypedRune([]rune(prefix)[0])
		for _, r := range prefix[1:] {
			e.entry.TypedRune(r)
		}
	}
}

// ShowFindDialog shows the find dialog
func (e *Editor) ShowFindDialog() {
	findEntry := widget.NewEntry()
	findEntry.PlaceHolder = "Find text..."
	
	var lastFoundIndex int
	
	content := widget.NewCard("Find", "", findEntry)
	content.Resize(fyne.NewSize(300, 100))
	
	findFunc := func() {
		searchText := findEntry.Text
		if searchText == "" {
			return
		}
		
		text := e.entry.Text
		
		// Find next occurrence
		index := strings.Index(text[lastFoundIndex:], searchText)
		if index >= 0 {
			// Found a match
			foundAt := lastFoundIndex + index
			lastFoundIndex = foundAt + 1
			
			// Highlight the found text by selecting it
			// Note: Fyne doesn't provide direct selection API, 
			// so we'll show a notification of where it was found
			dialog.ShowInformation("Found", 
				fmt.Sprintf("Found at position %d", foundAt), 
				e.controller.window)
		} else {
			// No more matches, wrap around
			lastFoundIndex = 0
			index = strings.Index(text, searchText)
			if index >= 0 {
				dialog.ShowInformation("Found", 
					fmt.Sprintf("Found at position %d (wrapped)", index), 
					e.controller.window)
				lastFoundIndex = index + 1
			} else {
				dialog.ShowInformation("Not Found", 
					fmt.Sprintf("'%s' not found", searchText), 
					e.controller.window)
			}
		}
	}
	
	findEntry.OnSubmitted = func(s string) { findFunc() }
	
	d := dialog.NewCustom("Find", "Close", content, e.controller.window)
	
	// Add Find Next button
	d.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Find Next", findFunc),
		widget.NewButton("Close", d.Hide),
	})
	
	d.Show()
}

// ShowReplaceDialog shows the find and replace dialog
func (e *Editor) ShowReplaceDialog() {
	findEntry := widget.NewEntry()
	findEntry.PlaceHolder = "Find text..."
	
	replaceEntry := widget.NewEntry()
	replaceEntry.PlaceHolder = "Replace with..."
	
	form := container.NewVBox(
		widget.NewLabel("Find:"),
		findEntry,
		widget.NewLabel("Replace:"),
		replaceEntry,
	)
	
	content := widget.NewCard("Find and Replace", "", form)
	content.Resize(fyne.NewSize(350, 200))
	
	d := dialog.NewCustom("Find and Replace", "Close", content, e.controller.window)
	
	replaceFunc := func() {
		find := findEntry.Text
		replace := replaceEntry.Text
		if find == "" {
			return
		}
		
		text := e.entry.Text
		if strings.Contains(text, find) {
			newText := strings.Replace(text, find, replace, 1)
			e.entry.SetText(newText)
		}
	}
	
	replaceAllFunc := func() {
		find := findEntry.Text
		replace := replaceEntry.Text
		if find == "" {
			return
		}
		
		text := e.entry.Text
		newText := strings.ReplaceAll(text, find, replace)
		e.entry.SetText(newText)
	}
	
	d.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Replace", replaceFunc),
		widget.NewButton("Replace All", replaceAllFunc),
		widget.NewButton("Close", d.Hide),
	})
	
	d.Show()
}

// Focus sets focus to the editor
func (e *Editor) Focus() {
	e.entry.FocusGained()
}