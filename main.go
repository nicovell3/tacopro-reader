package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/nicovell3/tacopro-reader/pkg/reader"
	"github.com/sf1/go-card/smartcard"
)

var cardContext *smartcard.Context

var application fyne.App
var mainWindow fyne.Window

var readers []*smartcard.Reader
var selectedReaderName string
var mainContent *fyne.Container
var generatedFilename string
var readingCard bool

func main() {
	var err error
	application = app.New()
	mainWindow = application.NewWindow("TacoPro Reader")

	cardContext, err = smartcard.EstablishContext()
	if err != nil {
		showError(err)
		application.Run()
		return
	}
	defer cardContext.Release()
	smartcardList()

	mainContent = container.New(layout.NewVBoxLayout())
	refreshMainContent()
	mainWindow.SetContent(mainContent)
	mainWindow.Resize(fyne.NewSize(300, 100))
	mainWindow.Show()
	application.Run()
	tidyUp()
}

func refreshMainContent() {
	mainContent.RemoveAll()
	mainContent.Add(widget.NewButton("Refresh readers list", func() {
		err := smartcardList()
		if err != nil {
			showError(err)
			return
		}
		refreshMainContent()
	}))
	mainContent.Add(widget.NewLabel("The following list shows the available card readers:"))
	mainContent.Add(showCardSelector())
	if selectedReaderName == "" {
		return
	}
	mainContent.Add(widget.NewButton("Read card", func() {
		readingCard = true
		refreshMainContent()
		var err error
		generatedFilename, err = readCard()
		if err != nil {
			showError(err)
			return
		}
		readingCard = false
		refreshMainContent()
	}))
	if readingCard {
		mainContent.Add(widget.NewLabel("Card is being read..."))
	} else if generatedFilename != "" {
		mainContent.Add(widget.NewLabel("Card read successfully and saved data to " + generatedFilename))
	}
}

func showCardSelector() fyne.CanvasObject {
	if len(readers) == 0 {
		selectedReaderName = ""
		return widget.NewLabel("No card readers available")
	} else if len(readers) == 1 {
		selectedReaderName = readers[0].Name()
		return widget.NewLabel(readers[0].Name())
	}
	readerNames := make([]string, len(readers))
	for i, reader := range readers {
		readerNames[i] = reader.Name()
	}
	return widget.NewRadioGroup(readerNames, func(value string) {
		log.Println("Value changed to", value)
		selectedReaderName = value
		refreshMainContent()
	})
}

func showError(err error) {
	errorWindow := application.NewWindow("TacoPro Reader - Error")
	errorWindow.SetContent(widget.NewLabel(fmt.Sprintf("Error: %v", err)))
	errorWindow.SetMaster()
	errorWindow.Show()
	if mainWindow != nil {
		log.Println("Closing main window")
		mainWindow.Close()
		mainWindow.Hide()
	}
}

func smartcardList() error {
	var err error
	readers, err = cardContext.ListReadersWithCard()
	if err != nil {
		return err
	}
	log.Println("Listed cards:", len(readers))
	return nil
}

func readCard() (string, error) {
	var currentReader *smartcard.Reader
	for _, reader := range readers {
		if reader.Name() == selectedReaderName {
			currentReader = reader
			break
		}
	}
	if currentReader == nil {
		return "", fmt.Errorf("no reader selected")
	}
	if !currentReader.IsCardPresent() {
		return "", fmt.Errorf("card is not inserted")
	}
	card, err := currentReader.Connect()
	if err != nil {
		return "", err
	}
	defer card.Disconnect()
	log.Printf("Card ATR: %s\n", card.ATR())
	filename, err := reader.ReadTGD("", card)
	if err != nil {
		return "", err
	}
	log.Println("File saved to", filename)
	return filename, nil
}

func tidyUp() {
	log.Println("Exit")
}
