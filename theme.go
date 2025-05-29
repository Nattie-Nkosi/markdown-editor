package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// myTheme is a custom theme for the markdown editor
type myTheme struct{}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 250, G: 250, B: 250, A: 255}
		}
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 45, G: 45, B: 45, A: 255}
		}
		return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 122, B: 255, A: 255}
	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	return theme.DefaultTheme().Font(style)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNamePadding:
		return 6
	}
	return theme.DefaultTheme().Size(name)
}