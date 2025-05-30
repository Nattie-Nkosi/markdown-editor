# Markdown Editor

<div align="center">
  <img src="icon.png" alt="Markdown Editor Icon" width="128" height="128">
  
  A cross-platform markdown editor built with Go and Fyne framework. This project was developed as part of my journey learning Go programming language while creating a practical and useful tool.

![Go Version](https://img.shields.io/badge/Go-1.24.3-00ADD8?style=flat&logo=go)
![Fyne Version](https://img.shields.io/badge/Fyne-2.6.1-00ADD8?style=flat)
![License](https://img.shields.io/badge/License-MIT-green.svg)

</div>

## 📸 Screenshot

<div align="center">
  <img src="app.png" alt="Markdown Editor Screenshot" width="800">
</div>

## 🎯 Project Motivation

As someone learning Go, I wanted to build something beyond simple tutorials - a real-world application that I could actually use. A markdown editor seemed perfect because:

- It's genuinely useful for daily work
- It covers many programming concepts (file I/O, GUI, text processing)
- It's complex enough to be challenging but achievable for a learning project
- Cross-platform desktop apps in Go showcase the language's versatility

## ✨ Features

### Core Functionality

- **Live Preview**: Real-time markdown rendering as you type
- **Syntax Support**: Full markdown syntax including headers, lists, links, images, code blocks, tables
- **File Operations**: Create, open, save, and save as functionality
- **Export to HTML**: Export your markdown as styled HTML with embedded CSS

### Editor Features

- **Smart Markdown Insertion**: Wrap selected text or insert with placeholders
- **Find & Replace**: Search and replace text within your documents
- **Line-based Operations**: Insert headers, lists, and quotes at line start
- **Status Bar**: Shows line count, word count, and character count
- **Unsaved Changes Protection**: Warns before closing or creating new files with unsaved changes

### User Interface

- **Modern Design**: Clean, intuitive interface with custom theme
- **Dual-pane Layout**: Editor and preview side by side
- **Toolbar**: Quick access to common formatting options
- **Comprehensive Menus**: Full menu system with keyboard shortcuts
- **Toggle Preview**: Hide/show preview pane for focused writing

### Keyboard Shortcuts

- `Ctrl+N` - New file
- `Ctrl+O` - Open file
- `Ctrl+S` - Save
- `Ctrl+Shift+S` - Save As
- `Ctrl+F` - Find
- `Ctrl+H` - Replace
- `Ctrl+P` - Toggle preview
- `Ctrl+Z/Y` - Undo/Redo
- `Ctrl+X/C/V` - Cut/Copy/Paste

## 🛠️ Technical Stack

- **Language**: Go 1.24.3
- **GUI Framework**: Fyne v2.6.1
- **Markdown Parser**: Goldmark
- **Architecture**: MVC-like pattern with separation of concerns
- **Platform Support**: Windows, macOS, Linux

## 📁 Project Structure

```
fynemd/
├── main.go          # Application entry point
├── controller.go    # Application state management
├── editor.go        # Text editor component
├── preview.go       # Markdown preview component
├── menu.go          # Menu system
├── toolbar.go       # Toolbar implementation
├── statusbar.go     # Status bar component
├── theme.go         # Custom theme definition
├── FyneApp.toml     # Application metadata
├── icon.png         # Application icon
├── image.png        # Screenshot
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── README.md        # This file
```

## 🚀 Getting Started

### Prerequisites

- Go 1.24.3 or higher
- Git

### Installation

1. Clone the repository:

```bash
git clone https://github.com/Nattie-Nkosi/markdown-editor.git
cd markdown-editor
```

2. Install dependencies:

```bash
go mod tidy
```

3. Run the application:

```bash
go run .
```

### Building

To build an executable:

```bash
# For current platform
go build -o markdown-editor

# For Windows
GOOS=windows GOARCH=amd64 go build -o markdown-editor.exe

# For macOS
GOOS=darwin GOARCH=amd64 go build -o markdown-editor

# For Linux
GOOS=linux GOARCH=amd64 go build -o markdown-editor
```

### Packaging with Fyne

To create a distributable package with icon:

```bash
# Install fyne tool
go install fyne.io/tools/cmd/fyne@latest

# Package for current platform
fyne package -release

# Or specify platform
fyne package -os windows -release
fyne package -os darwin -release
fyne package -os linux -release
```

## 💡 Learning Highlights

Building this project taught me several Go concepts:

1. **Goroutines & Concurrency**: Although this app is primarily event-driven, understanding Go's concurrency model was crucial
2. **Interfaces**: Fyne's widget system heavily uses interfaces, perfect for learning Go's interface philosophy
3. **Package Organization**: Structuring code with proper separation of concerns
4. **Error Handling**: Go's explicit error handling throughout file operations
5. **Cross-platform Development**: Building truly portable desktop applications

## 🔧 Architecture Decisions

### Why Fyne?

- Pure Go (no CGo dependencies)
- Cross-platform with native look
- Active development and community
- Good documentation for learners

### Design Patterns

- **MVC-like Architecture**: Controller manages state, components handle UI
- **Composition over Inheritance**: Go's approach with embedded structs
- **Interface-based Design**: Loose coupling between components

## 🚧 Future Enhancements

As I continue learning Go, I plan to add:

- [ ] Syntax highlighting in the editor
- [ ] Plugin system for extending functionality
- [ ] Themes and preferences
- [ ] Auto-save functionality
- [ ] Recent files menu
- [ ] Markdown extensions (Mermaid, Math)
- [ ] Split view for multiple files
- [ ] Vim/Emacs key bindings
- [ ] Spell check integration
- [ ] Git integration

## 🤝 Contributing

As this is a learning project, I welcome contributions, suggestions, and feedback! Feel free to:

- Report bugs
- Suggest features
- Submit pull requests
- Share Go best practices

## 🙏 Acknowledgments

- The Go team for creating such an elegant language
- The Fyne team for the excellent GUI framework
- The Goldmark team for the robust markdown parser
- The Go community for amazing learning resources
- Everyone who creates and shares markdown editors that inspired this project

## 📚 Learning Resources

If you're also learning Go, here are resources I found helpful:

- [The Go Programming Language](https://www.gopl.io/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Fyne Documentation](https://developer.fyne.io/)
- [Go by Example](https://gobyexample.com/)

---

<div align="center">
  <i>This project is part of my Go learning journey. If you're also learning Go, I hope this code helps you as much as building it helped me!</i>
</div>
