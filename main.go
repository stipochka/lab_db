package main

import (
	"lab_db/internal/entry_window"
	"lab_db/internal/window"

	"fyne.io/fyne/v2/app"
)

func main() {
	app := app.New()

	entryWin := entry_window.NewEntryWindow(app)

	entryWin.OnFileSelect = func(selectedFile string) {
		mainWin := window.New(app, selectedFile)

		mainWin.Show("Table " + selectedFile)
	}

	entryWin.Show()
}
