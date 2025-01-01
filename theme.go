package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Theme struct {
	fyne.Theme
}

func NewTheme() fyne.Theme {
	return &Theme{
		Theme: theme.DefaultTheme(),
	}
}

var PALLETTE = map[string]color.Color{
	// #101217
	"background": color.RGBA{16, 18, 23, 255},
	"sidebar_info": color.RGBA{0, 0, 0, 255},
	// #262A36
	"sidebar_hover": color.RGBA{38, 42, 54, 255},
	//rgb(0, 15, 73)
	"sidebar_selected": color.RGBA{0, 15, 73, 255},
	// 3B4254
	"input_border": color.RGBA{59, 66, 84, 255},
}

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {

	if name == theme.ColorNameBackground {
		return PALLETTE["background"]
	}

	if name == theme.ColorNameSelection {
		return PALLETTE["sidebar_selected"]
	}

	if name == theme.ColorNameHover {
		return PALLETTE["sidebar_hover"]
	}

	// track bar color
	if name == theme.ColorNameScrollBar {
		return PALLETTE["sidebar_info"]
	}

	if name == theme.ColorNameInputBackground {
		return PALLETTE["background"]
	}

	if name == theme.ColorNameInputBorder {
		return PALLETTE["input_border"]
	}

	if name == theme.ColorNameFocus {
		return PALLETTE["sidebar_selected"]
	}

	return t.Theme.Color(name, variant)
}


func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 12
	}
	return t.Theme.Size(name)
}

// func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {