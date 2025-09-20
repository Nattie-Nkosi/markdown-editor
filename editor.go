package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

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
	selection := e.entry.SelectedText()
	if selection != "" {
		wrapped := before + selection + after
		e.pasteText(wrapped)

		if after != "" {
			current := e.cursorIndex()
			e.setCursorAtIndex(current - runeCount(after))
		}
		return
	}

	insertText := before + placeholder + after
	cursorIndex := e.cursorIndex()

	e.pasteText(insertText)

	if placeholder != "" {
		e.setCursorAtIndex(cursorIndex + runeCount(before))
	} else {
		e.setCursorAtIndex(cursorIndex + runeCount(insertText))
	}
}

// InsertAtLineStart inserts text at the beginning of the current line
func (e *Editor) InsertAtLineStart(prefix string) {
	if prefix == "" {
		return
	}

	if selection := e.entry.SelectedText(); selection != "" {
		lines := strings.Split(selection, "\n")
		for i, line := range lines {
			lines[i] = prefix + line
		}

		e.pasteText(strings.Join(lines, "\n"))
		return
	}

	cursor := e.cursorIndex()
	textRunes := []rune(e.entry.Text)

	lineStart := cursor
	for lineStart > 0 && textRunes[lineStart-1] != '\n' {
		lineStart--
	}

	relative := cursor - lineStart

	e.setCursorAtIndex(lineStart)
	e.pasteText(prefix)
	e.setCursorAtIndex(lineStart + runeCount(prefix) + relative)
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

	resetSearch := func(string) {
		lastIndex = 0
	}

	findEntry.OnChanged = resetSearch

	findNext := func() {
		searchText := findEntry.Text
		if searchText == "" {
			resultLabel.SetText("Enter search text")
			return
		}

		content := e.entry.Text
		runes := []rune(content)
		searchRunes := []rune(searchText)
		if len(searchRunes) == 0 {
			resultLabel.SetText("Enter search text")
			return
		}

		wrapped := false
		index := findFoldIndex(runes, searchRunes, lastIndex)
		if index < 0 && lastIndex > 0 {
			wrapped = true
			index = findFoldIndex(runes, searchRunes, 0)
		}

		if index < 0 {
			resultLabel.SetText("Not found")
			return
		}

		prefix := string(runes[:index])
		lineNum := strings.Count(prefix, "\n") + 1
		if wrapped {
			resultLabel.SetText(fmt.Sprintf("Found at line %d (wrapped)", lineNum))
		} else {
			resultLabel.SetText(fmt.Sprintf("Found at line %d", lineNum))
		}

		e.setCursorAtIndex(index)

		lastIndex = index + len(searchRunes)
		if lastIndex > len(runes) {
			lastIndex = len(runes)
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
			resultLabel.SetText(fmt.Sprintf("Replaced %d occurrences", count))
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

func (e *Editor) pasteText(text string) {
	clipboard := &staticClipboard{content: text}
	e.entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
}

func (e *Editor) cursorIndex() int {
	runes := []rune(e.entry.Text)
	if len(runes) == 0 {
		return 0
	}

	row := 0
	index := 0
	for index < len(runes) && row < e.entry.CursorRow {
		if runes[index] == '\n' {
			row++
		}
		index++
	}

	column := 0
	for index < len(runes) && column < e.entry.CursorColumn {
		if runes[index] == '\n' {
			break
		}
		column++
		index++
	}

	return index
}

func (e *Editor) setCursorAtIndex(idx int) {
	runes := []rune(e.entry.Text)
	if idx < 0 {
		idx = 0
	}
	if idx > len(runes) {
		idx = len(runes)
	}

	row := 0
	column := 0
	for i := 0; i < idx; i++ {
		if runes[i] == '\n' {
			row++
			column = 0
			continue
		}
		column++
	}

	e.entry.CursorRow = row
	e.entry.CursorColumn = column
	e.entry.Refresh()
}

func findFoldIndex(content []rune, search []rune, start int) int {
	if len(search) == 0 {
		return -1
	}
	if start < 0 {
		start = 0
	}
	if start > len(content) {
		start = len(content)
	}

	target := string(search)
	for i := start; i <= len(content)-len(search); i++ {
		segment := string(content[i : i+len(search)])
		if strings.EqualFold(segment, target) {
			return i
		}
	}
	return -1
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

type staticClipboard struct {
	content string
}

func (c *staticClipboard) Content() string {
	return c.content
}

func (c *staticClipboard) SetContent(string) {}
