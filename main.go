package main

import (
	"ebook-meta-editor/dialogs"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

func main() {

	app := app.New()
	app.Settings().SetTheme(NewTheme())
	window := app.NewWindow(" ")
	drv := fyne.CurrentApp().Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		fmt.Println("Driver is desktop", drv)
		// window = drv.CreateSplashWindow()
	}
	window.Resize(fyne.NewSize(800, 600))
	// borderless
	window.CenterOnScreen()

	gui := &GUI{
		window: window,
	}

	wizard := dialogs.NewWizard("Welcome to the Epub Editor", dialogs.WelcomeDialog())
	wizard.Show(window)

	window.SetContent(gui.makeGUI())
	window.SetMainMenu(gui.makeMenu())

	window.ShowAndRun()

}
