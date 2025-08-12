package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// AppController manages the application state and coordinates between components
type AppController struct {
	window       fyne.Window
	editor       *Editor
	preview      *Preview
	statusBar    *StatusBar
	currentFile  fyne.URI
	modified     bool
	saveMenuItem *fyne.MenuItem
}

// NewAppController creates a new application controller
func NewAppController(window fyne.Window) *AppController {
	return &AppController{
		window:   window,
		modified: false,
	}
}

// SetEditor sets the editor component
func (c *AppController) SetEditor(editor *Editor) {
	c.editor = editor
}

// SetPreview sets the preview component
func (c *AppController) SetPreview(preview *Preview) {
	c.preview = preview
}

// SetStatusBar sets the status bar component
func (c *AppController) SetStatusBar(statusBar *StatusBar) {
	c.statusBar = statusBar
}

// SetSaveMenuItem sets the save menu item for enabling/disabling
func (c *AppController) SetSaveMenuItem(item *fyne.MenuItem) {
	c.saveMenuItem = item
}

// OnTextChanged handles text changes in the editor
func (c *AppController) OnTextChanged(content string) {
	// Update preview
	if c.preview != nil {
		c.preview.UpdateContent(content)
	}
	
	// Mark as modified
	if !c.modified {
		c.modified = true
		c.updateTitle()
	}
	
	// Enable save menu item
	if c.saveMenuItem != nil {
		c.saveMenuItem.Disabled = false
	}
	
	// Update status
	c.updateStatus()
}

// NewFile creates a new file
func (c *AppController) NewFile() {
	if c.modified {
		dialog.ShowConfirm("Unsaved Changes", 
			"Do you want to save your changes before creating a new file?",
			func(save bool) {
				if save {
					c.Save()
				}
				c.createNewFile()
			}, c.window)
	} else {
		c.createNewFile()
	}
}

func (c *AppController) createNewFile() {
	c.editor.SetContent("")
	c.currentFile = nil
	c.modified = false
	c.updateTitle()
	c.updateStatus()
	if c.saveMenuItem != nil {
		c.saveMenuItem.Disabled = true
	}
}

// Open opens a file
func (c *AppController) Open() {
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}

		c.editor.SetContent(string(data))
		c.currentFile = reader.URI()
		c.modified = false
		c.updateTitle()
		c.updateStatus()
		
		if c.saveMenuItem != nil {
			c.saveMenuItem.Disabled = true
		}
	}, c.window)

	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown", ".txt"}))
	openDialog.Show()
}

// Save saves the current file
func (c *AppController) Save() {
	if c.currentFile != nil {
		c.saveToFile(c.currentFile)
	} else {
		c.SaveAs()
	}
}

// SaveAs saves the file with a new name
func (c *AppController) SaveAs() {
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if writer == nil {
			return
		}
		
		c.currentFile = writer.URI()
		writer.Close()
		c.saveToFile(c.currentFile)
	}, c.window)

	saveDialog.SetFileName("untitled.md")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown"}))
	saveDialog.Show()
}

func (c *AppController) saveToFile(uri fyne.URI) {
	writer, err := storage.Writer(uri)
	if err != nil {
		dialog.ShowError(err, c.window)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte(c.editor.GetContent()))
	if err != nil {
		dialog.ShowError(err, c.window)
		return
	}

	c.modified = false
	c.updateTitle()
	c.updateStatus()
	
	if c.saveMenuItem != nil {
		c.saveMenuItem.Disabled = true
	}
	
	if c.statusBar != nil {
		c.statusBar.SetText(fmt.Sprintf("Saved: %s", uri.Name()))
	}
}

// ExportHTML exports the markdown to HTML
func (c *AppController) ExportHTML() {
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		html := c.preview.GetHTML()
		_, err = writer.Write([]byte(html))
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		
		if c.statusBar != nil {
			c.statusBar.SetText(fmt.Sprintf("Exported to: %s", writer.URI().Name()))
		}
	}, c.window)

	saveDialog.SetFileName("export.html")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".html", ".htm"}))
	saveDialog.Show()
}

// HandleClose handles window close event
func (c *AppController) HandleClose() {
	if c.modified {
		dialog.ShowConfirm("Unsaved Changes",
			"Do you want to save your changes before closing?",
			func(save bool) {
				if save {
					c.Save()
				}
				c.window.Close()
			}, c.window)
	} else {
		c.window.Close()
	}
}

// InsertMarkdown inserts markdown syntax
func (c *AppController) InsertMarkdown(before, after string, placeholder string) {
	if c.editor != nil {
		c.editor.InsertMarkdown(before, after, placeholder)
	}
}

// InsertAtLineStart inserts text at the beginning of the current line
func (c *AppController) InsertAtLineStart(prefix string) {
	if c.editor != nil {
		c.editor.InsertAtLineStart(prefix)
	}
}

// TogglePreview toggles the preview pane visibility
func (c *AppController) TogglePreview() {
	if c.preview != nil {
		c.preview.ToggleVisibility()
	}
}

// ShowFind shows the find dialog
func (c *AppController) ShowFind() {
	if c.editor != nil {
		c.editor.ShowFindDialog()
	}
}

// ShowReplace shows the replace dialog
func (c *AppController) ShowReplace() {
	if c.editor != nil {
		c.editor.ShowReplaceDialog()
	}
}

func (c *AppController) updateTitle() {
	title := "Markdown Editor"
	if c.currentFile != nil {
		title = fmt.Sprintf("%s - %s", title, c.currentFile.Name())
	} else {
		title = fmt.Sprintf("%s - Untitled", title)
	}
	if c.modified {
		title = fmt.Sprintf("%s *", title)
	}
	c.window.SetTitle(title)
}

func (c *AppController) updateStatus() {
	if c.statusBar != nil && c.editor != nil {
		content := c.editor.GetContent()
		lines := strings.Count(content, "\n") + 1
		words := len(strings.Fields(content))
		chars := len(content)
		
		status := fmt.Sprintf("Lines: %d | Words: %d | Characters: %d", lines, words, chars)
		c.statusBar.SetText(status)
	}
}