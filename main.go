package main

import (
	"ebook-meta-editor/dialogs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {

	app := app.New()
	app.Settings().SetTheme(NewTheme())
	window := app.NewWindow("Epub Editor")
	window.Resize(fyne.NewSize(800, 600))

	gui := &GUI{
		window: window,
	}

	wizard := dialogs.NewWizard("Welcome to the Epub Editor", dialogs.WelcomeDialog())
	wizard.Show(window)

	window.SetContent(gui.makeGUI())
	window.SetMainMenu(gui.makeMenu())

	window.ShowAndRun()

}
