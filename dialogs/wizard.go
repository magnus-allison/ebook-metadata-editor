package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Wizard struct {
	title string
	stack []fyne.CanvasObject
	content *fyne.Container

	dialog dialog.Dialog
}

func NewWizard(title string, content fyne.CanvasObject) *Wizard {
	wizard := &Wizard{
		title: title,
		stack: []fyne.CanvasObject{content},
	}
	wizard.content = container.NewStack(content)
	return wizard
}

func (w *Wizard) Show(window fyne.Window) {
	w.dialog = dialog.NewCustomWithoutButtons(w.title, w.content, window)
}

func (w *Wizard) Hide(content fyne.CanvasObject) {
	w.dialog.Hide()
}

func WelcomeDialog() *fyne.Container {
	label := widget.NewLabel("Welcome to the Epub Editor")

	return container.NewVBox(label)
}