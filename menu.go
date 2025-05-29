package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

// Menu represents the application menu
type Menu struct {
	controller *AppController
}

// NewMenu creates a new menu instance
func NewMenu(controller *AppController) *Menu {
	return &Menu{
		controller: controller,
	}
}

// CreateMainMenu creates the main menu
func (m *Menu) CreateMainMenu() *fyne.MainMenu {
	// File menu
	newItem := fyne.NewMenuItem("New", m.controller.NewFile)
	newItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}
	
	openItem := fyne.NewMenuItem("Open...", m.controller.Open)
	openItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	
	saveItem := fyne.NewMenuItem("Save", m.controller.Save)
	saveItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	saveItem.Disabled = true
	m.controller.SetSaveMenuItem(saveItem)
	
	saveAsItem := fyne.NewMenuItem("Save As...", m.controller.SaveAs)
	saveAsItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}
	
	exportHTMLItem := fyne.NewMenuItem("Export as HTML...", m.controller.ExportHTML)
	
	fileMenu := fyne.NewMenu("File",
		newItem,
		openItem,
		fyne.NewMenuItemSeparator(),
		saveItem,
		saveAsItem,
		fyne.NewMenuItemSeparator(),
		exportHTMLItem,
	)
	
	// Edit menu
	undoItem := fyne.NewMenuItem("Undo", func() {
		m.controller.editor.entry.TypedShortcut(&fyne.ShortcutUndo{})
	})
	
	redoItem := fyne.NewMenuItem("Redo", func() {
		m.controller.editor.entry.TypedShortcut(&fyne.ShortcutRedo{})
	})
	
	cutItem := fyne.NewMenuItem("Cut", func() {
		m.controller.editor.entry.TypedShortcut(&fyne.ShortcutCut{})
	})
	
	copyItem := fyne.NewMenuItem("Copy", func() {
		m.controller.editor.entry.TypedShortcut(&fyne.ShortcutCopy{})
	})
	
	pasteItem := fyne.NewMenuItem("Paste", func() {
		m.controller.editor.entry.TypedShortcut(&fyne.ShortcutPaste{})
	})
	
	selectAllItem := fyne.NewMenuItem("Select All", func() {
		m.controller.editor.entry.TypedShortcut(&fyne.ShortcutSelectAll{})
	})
	
	findItem := fyne.NewMenuItem("Find...", m.controller.ShowFind)
	findItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierControl}
	
	replaceItem := fyne.NewMenuItem("Replace...", m.controller.ShowReplace)
	replaceItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyH, Modifier: fyne.KeyModifierControl}
	
	editMenu := fyne.NewMenu("Edit",
		undoItem,
		redoItem,
		fyne.NewMenuItemSeparator(),
		cutItem,
		copyItem,
		pasteItem,
		selectAllItem,
		fyne.NewMenuItemSeparator(),
		findItem,
		replaceItem,
	)
	
	// View menu
	togglePreviewItem := fyne.NewMenuItem("Toggle Preview", m.controller.TogglePreview)
	togglePreviewItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: fyne.KeyModifierControl}
	
	viewMenu := fyne.NewMenu("View",
		togglePreviewItem,
	)
	
	// Insert menu
	insertMenu := fyne.NewMenu("Insert",
		fyne.NewMenuItem("Bold", func() {
			m.controller.InsertMarkdown("**", "**", "bold text")
		}),
		fyne.NewMenuItem("Italic", func() {
			m.controller.InsertMarkdown("*", "*", "italic text")
		}),
		fyne.NewMenuItem("Code", func() {
			m.controller.InsertMarkdown("`", "`", "code")
		}),
		fyne.NewMenuItem("Strikethrough", func() {
			m.controller.InsertMarkdown("~~", "~~", "strikethrough")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Link", func() {
			m.controller.InsertMarkdown("[", "](url)", "link text")
		}),
		fyne.NewMenuItem("Image", func() {
			m.controller.InsertMarkdown("![", "](url)", "alt text")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Heading 1", func() {
			m.controller.InsertAtLineStart("# ")
		}),
		fyne.NewMenuItem("Heading 2", func() {
			m.controller.InsertAtLineStart("## ")
		}),
		fyne.NewMenuItem("Heading 3", func() {
			m.controller.InsertAtLineStart("### ")
		}),
		fyne.NewMenuItem("Heading 4", func() {
			m.controller.InsertAtLineStart("#### ")
		}),
		fyne.NewMenuItem("Heading 5", func() {
			m.controller.InsertAtLineStart("##### ")
		}),
		fyne.NewMenuItem("Heading 6", func() {
			m.controller.InsertAtLineStart("###### ")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Unordered List", func() {
			m.controller.InsertAtLineStart("- ")
		}),
		fyne.NewMenuItem("Ordered List", func() {
			m.controller.InsertAtLineStart("1. ")
		}),
		fyne.NewMenuItem("Task List", func() {
			m.controller.InsertAtLineStart("- [ ] ")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Blockquote", func() {
			m.controller.InsertAtLineStart("> ")
		}),
		fyne.NewMenuItem("Code Block", func() {
			m.controller.InsertMarkdown("```\n", "\n```", "language")
		}),
		fyne.NewMenuItem("Horizontal Rule", func() {
			m.controller.InsertMarkdown("\n---\n", "", "")
		}),
		fyne.NewMenuItem("Table", func() {
			table := "\n| Header 1 | Header 2 | Header 3 |\n|----------|----------|----------|\n| Cell 1   | Cell 2   | Cell 3   |\n| Cell 4   | Cell 5   | Cell 6   |\n"
			m.controller.InsertMarkdown(table, "", "")
		}),
	)
	
	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Markdown Cheatsheet", func() {
			m.showMarkdownCheatsheet()
		}),
		fyne.NewMenuItem("Keyboard Shortcuts", func() {
			m.showKeyboardShortcuts()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About", 
				"Markdown Editor\nVersion 1.0\n\nA simple yet powerful markdown editor built with Fyne.\n\n© 2024", 
				m.controller.window)
		}),
	)
	
	return fyne.NewMainMenu(
		fileMenu,
		editMenu,
		viewMenu,
		insertMenu,
		helpMenu,
	)
}

func (m *Menu) showMarkdownCheatsheet() {
	content := `# Markdown Cheatsheet

## Headers
# H1
## H2
### H3
#### H4
##### H5
###### H6

## Emphasis
*italic* or _italic_
**bold** or __bold__
**_bold italic_**
~~strikethrough~~

## Lists
Unordered:
- Item 1
- Item 2
  - Subitem 2.1
  - Subitem 2.2

Ordered:
1. First item
2. Second item
   1. Subitem 2.1
   2. Subitem 2.2

Task List:
- [x] Completed task
- [ ] Incomplete task

## Links & Images
[Link text](https://example.com)
![Alt text](image.jpg)

## Code
Inline code: ` + "`code`" + `

Code block:
` + "```" + `
function example() {
    return "Hello, World!";
}
` + "```" + `

## Blockquotes
> This is a blockquote
> 
> With multiple paragraphs

## Tables
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |

## Horizontal Rule
---
or
***`

	dialog.ShowInformation("Markdown Cheatsheet", content, m.controller.window)
}

func (m *Menu) showKeyboardShortcuts() {
	shortcuts := `File Operations:
Ctrl+N - New File
Ctrl+O - Open File
Ctrl+S - Save
Ctrl+Shift+S - Save As

Edit Operations:
Ctrl+Z - Undo
Ctrl+Y - Redo
Ctrl+X - Cut
Ctrl+C - Copy
Ctrl+V - Paste
Ctrl+A - Select All
Ctrl+F - Find
Ctrl+H - Replace

View:
Ctrl+P - Toggle Preview`

	dialog.ShowInformation("Keyboard Shortcuts", shortcuts, m.controller.window)
}