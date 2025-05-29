package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
)

func main() {
	// Create application
	myApp := app.New()
	myApp.Settings().SetTheme(&myTheme{})

	// Create main window
	window := myApp.NewWindow("Markdown Editor")
	window.Resize(fyne.NewSize(1000, 600))
	window.CenterOnScreen()

	// Create application controller
	appController := NewAppController(window)

	// Create UI components
	editor := NewEditor(appController)
	preview := NewPreview()

	// Set up the controller
	appController.SetEditor(editor)
	appController.SetPreview(preview)

	// Create menu
	menu := NewMenu(appController)
	window.SetMainMenu(menu.CreateMainMenu())

	// Create toolbar
	toolbar := NewToolbar(appController)

	// Create status bar
	statusBar := NewStatusBar()
	appController.SetStatusBar(statusBar)

	// Layout the application
	content := container.NewBorder(
		toolbar.Create(),
		statusBar.Create(),
		nil,
		nil,
		container.NewHSplit(
			editor.Create(),
			preview.Create(),
		),
	)

	window.SetContent(content)

	// Set up keyboard shortcuts
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		appController.NewFile()
	})
	
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		appController.Open()
	})
	
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		appController.Save()
	})
	
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
	}, func(shortcut fyne.Shortcut) {
		appController.SaveAs()
	})
	
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyF, Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		appController.ShowFind()
	})
	
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyH, Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		appController.ShowReplace()
	})
	
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyP, Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		appController.TogglePreview()
	})

	// Handle window close
	window.SetCloseIntercept(func() {
		appController.HandleClose()
	})

	// Show and run
	window.ShowAndRun()
}