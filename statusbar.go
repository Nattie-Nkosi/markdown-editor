package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// StatusBar represents the application status bar
type StatusBar struct {
	label *widget.Label
}

// NewStatusBar creates a new status bar instance
func NewStatusBar() *StatusBar {
	return &StatusBar{
		label: widget.NewLabel("Ready"),
	}
}

// Create creates the status bar UI component
func (s *StatusBar) Create() fyne.CanvasObject {
	return container.NewBorder(
		widget.NewSeparator(),
		nil,
		nil,
		nil,
		container.NewPadded(s.label),
	)
}

// SetText updates the status bar text
func (s *StatusBar) SetText(text string) {
	s.label.SetText(text)
}