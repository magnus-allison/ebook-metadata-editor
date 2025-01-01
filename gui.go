package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	gif "fyne.io/x/fyne/widget"
)

type GUI struct {
	window fyne.Window
	book  *EPub
	content *fyne.Container
	sidebar *fyne.Container
}

func (g *GUI) makeGUI() fyne.CanvasObject {
	g.content = container.NewVBox() // Initialize empty container
	books, err := CollectBooks("./EPUB")
	if err != nil {
		panic(err)
	}

	g.content = container.NewVBox()
	g.renderSideBar(books)
	g.updateContent()

	split := container.NewHSplit(g.sidebar, g.content)
	split.SetOffset(0.25)

	return container.New(layout.NewBorderLayout(nil, nil, nil, nil), split)
}

func (g *GUI) renderSideBar(books []*EPub) {

	bookList := widget.NewList(
		func() int { return len(books) },
		func() fyne.CanvasObject {
			title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
			return container.NewWithoutLayout(title)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			container := o.(*fyne.Container)
			title := container.Objects[0].(*widget.Label)
			title.SetText(books[i].OPFData.Metadata.Title)
		},
	)

	// Create the title and found text
	title := canvas.NewText("Your Library", color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}

	amt := strconv.Itoa(len(books))
	found := canvas.NewText("Found: " + amt, color.White)
	found.TextStyle = fyne.TextStyle{Italic: true}
	found.TextSize = 10

	background := canvas.NewRectangle(PALLETTE["sidebar_info"])
	content := container.NewVBox(
		title,
		found,
	)
	paddedContent := container.New(&paddingLayout{padding: 10}, content)
	topBar := container.NewStack(
		background,
		paddedContent,
	)
	topBar = container.NewBorder(nil, nil, nil, nil, topBar)
	sidebar := container.NewBorder(
		topBar,                  // Top widget
		nil,                      // Bottom widget
		nil,                      // Left widget
		nil,                      // Right widget
		bookList,                 // Center widget, takes remaining space
	)

	bookList.OnSelected = func(id widget.ListItemID) {
		g.book = books[id]
		g.updateContent()
	}

	g.sidebar = sidebar
}

func (g *GUI) updateContent() {

	g.content.Objects = nil
	// spinner := g.renderSpinner(50, 50)
	// g.content.Add(spinner)

	if g.book == nil || g.book.OPFData == nil {
		emptyText := canvas.NewText("Select a book from the sidebar", color.White)
		emptyText.Alignment = fyne.TextAlignCenter

		container := container.New(layout.NewVBoxLayout(), emptyText)
		g.content = container
		g.content.Refresh()
		return
	}


	title := canvas.NewText(g.book.OPFData.Metadata.Title, color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 22

	author := canvas.NewText(g.book.OPFData.Metadata.Creator, color.White)
	author.TextStyle = fyne.TextStyle{Bold: true}
	author.Alignment = fyne.TextAlignCenter
	author.TextSize = 13

	coverPath := canvas.NewText(g.book.CoverImagePath, color.White)
	coverPath.TextStyle = fyne.TextStyle{Italic:  true}
	coverPath.Alignment = fyne.TextAlignCenter
	coverPath.TextSize = 10

	topInfoContent := container.NewVBox(title, author, coverPath)
	topInfo := container.New(&paddingLayout{padding: 10}, topInfoContent)

	// Load image from URL
	imageContainer := renderBookCover(g.book.CoverImagePath)

	// Input Fields
	titleInput := renderInput("Title", g.book.OPFData.Metadata.Title)
	authorInput := renderInput("Author", g.book.OPFData.Metadata.Creator)
	identInput := renderInput("ISBN", g.book.OPFData.Metadata.Identifier)
	dateInput := renderInput("Date", g.book.OPFData.Metadata.Date)

	metaInfoContent := container.NewVBox(titleInput, authorInput, identInput, dateInput)
	metaInfo := container.New(&paddingLayout{padding: 60}, metaInfoContent)


	// Sticky Bottom Bar
	bottomBar := container.New(&paddingLayout{10}, widget.NewButton("Save", func() {
		// Save action
	}))

	// Main Content
	mainContent := container.NewVBox(
		topInfo,
		layout.NewSpacer(), // Push content down
		imageContainer,
		metaInfo,
		layout.NewSpacer(), // Push content up
	)

	content := container.New(layout.NewBorderLayout(nil, bottomBar, nil, nil), mainContent, bottomBar)
	// g.content.Remove(spinner)
	g.content.Add(content)
	g.content.Refresh()
}

func (g *GUI) openProject() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
	}, g.window);

}

func (g *GUI) renderSpinner(w,h int) *gif.AnimatedGif {

	// https://icons8.com/preloaders/
	spinner, _ := gif.NewAnimatedGif(storage.NewFileURI("assets/1490.gif"))
	spinner.SetMinSize(fyne.NewSize(float32(w), float32(h)))
	spinner.Start()
	return spinner
}

func (g *GUI) makeMenu() *fyne.MainMenu {
	file := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Project", g.openProject),
	)

	return fyne.NewMainMenu(file)
}

func (g *GUI) renderTopBar() fyne.CanvasObject {

	title := canvas.NewText("Epub Editor", color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(
		title,
		layout.NewSpacer(),
	)
}

func renderInput(label, value string) fyne.CanvasObject {

	labelText := canvas.NewText(label, color.White)

	entry := widget.NewEntry()
	entry.SetText(value)
	entry.SetPlaceHolder(label)

	return container.NewVBox(labelText, entry)
}

func renderBookCover(path string) fyne.CanvasObject {


	img, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		// fallback img
		img, _ = fyne.LoadResourceFromPath("assets/cover.jpg")
	}
	cover := canvas.NewImageFromResource(img)
	cover.FillMode = canvas.ImageFillContain
	cover.SetMinSize(fyne.NewSize(200, 300))

	// Add spacing around the image
	imageContainer := container.NewVBox(
		layout.NewSpacer(),
		cover,
		layout.NewSpacer(),
	)

	return imageContainer
}

type paddingLayout struct {
	padding float32
}

func (p *paddingLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	// Place the content with padding
	content := objects[0]
	content.Resize(size.Subtract(fyne.NewSize(p.padding*2, p.padding*2)))
	content.Move(fyne.NewPos(p.padding, p.padding))
}

func (p *paddingLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	content := objects[0]
	return content.MinSize().Add(fyne.NewSize(p.padding*2, p.padding*2))
}
