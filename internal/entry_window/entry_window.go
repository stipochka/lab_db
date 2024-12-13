package entry_window

import (
	"fmt"
	"io/ioutil"
	"lab_db/internal/lib"
	"lab_db/internal/record"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type EntryWindow struct {
	App               fyne.App
	MainWindow        fyne.Window
	SelectedFile      string
	OnFileSelect      func(string)
	fileListContainer *fyne.Container
}

func NewEntryWindow(app fyne.App) *EntryWindow {
	return &EntryWindow{
		App: app,
	}
}

func CreateDatabase(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func (ew *EntryWindow) scanBackupStorage(folder string) ([]string, error) {
	var backupFiles []string

	items, err := ioutil.ReadDir(folder)
	if err != nil {
		return backupFiles, err
	}

	for _, file := range items {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			backupFiles = append(backupFiles, file.Name())
		}
	}

	return backupFiles, nil
}

func (ew *EntryWindow) Show() {
	ew.fileListContainer = container.NewVBox()

	ew.MainWindow = ew.App.NewWindow("Select Database File")

	files, err := ew.scanStorageFolder("/home/stepa/lab_db/storage/")
	if err != nil {
		dialog := widget.NewLabel("Error reading files: " + err.Error())
		ew.MainWindow.SetContent(container.NewVBox(dialog))
		ew.MainWindow.ShowAndRun()
		return
	}

	fileListContainer := container.NewVBox()

	for _, fileName := range files {
		fileButton := widget.NewButton(fileName, func(name string) func() {
			return func() {
				ew.SelectedFile = "/home/stepa/lab_db/storage/" + name
				fmt.Printf("Selected file: %s\n", ew.SelectedFile)

				if ew.OnFileSelect != nil {
					ew.OnFileSelect(ew.SelectedFile)
				}
			}
		}(fileName))

		fileButton.Resize(fyne.NewSize(400, 50))
		fileListContainer.Add(fileButton)
	}

	backupListContainer := container.NewVBox()

	backupFiles, err := ew.scanBackupStorage("/home/stepa/lab_db/backup")
	if err != nil {
		dialog := widget.NewLabel("Error reading files: " + err.Error())
		ew.MainWindow.SetContent(container.NewVBox(dialog))
		ew.MainWindow.ShowAndRun()
		return
	}

	for _, fileName := range backupFiles {
		fileButton := widget.NewButton(fileName, func(name string) func() {
			return func() {
				uploadedBackup, err := lib.CreateFromBackup(fileName)
				if err != nil {
					fmt.Printf("%v", err)
				}
				ew.SelectedFile = uploadedBackup
				fmt.Printf("Selected file: %s\n", ew.SelectedFile)

				if ew.OnFileSelect != nil {
					ew.OnFileSelect(ew.SelectedFile)
				}
			}
		}(fileName))

		fileButton.Resize(fyne.NewSize(400, 50))
		backupListContainer.Add(fileButton)
	}

	createDbFileEntry := widget.NewEntry()
	createDbFileEntry.SetPlaceHolder("Enter db file name")

	createButton := widget.NewButton("Create Database", func() {
		filename := createDbFileEntry.Text

		if len(filename) == 0 {
			return
		}

		filename, err := record.CreateDatabase(filename)
		if err != nil {
			fmt.Printf("failed to create database %v", err)
			return
		}

		ew.OnFileSelect(filename)

		ew.addFileButton(filename)

		fmt.Println("created db file")
	})

	ew.MainWindow.SetContent(container.NewVBox(
		widget.NewLabel("Select a Database File:"),
		fileListContainer,
		createDbFileEntry,
		createButton,
		widget.NewLabel("Select backup File:"),
		backupListContainer,
	))

	ew.MainWindow.Resize(fyne.NewSize(600, 500))
	ew.MainWindow.ShowAndRun()
}

func (ew *EntryWindow) scanStorageFolder(folder string) ([]string, error) {
	files := []string{}
	items, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if !item.IsDir() && strings.HasSuffix(item.Name(), ".db") {
			files = append(files, item.Name())
		}
	}
	return files, nil
}

func (ew *EntryWindow) populateFileButtons(files []string) {
	for _, fileName := range files {
		ew.addFileButton(fileName)
	}
}

func (ew *EntryWindow) addFileButton(fileName string) {
	fileButton := widget.NewButton(fileName, func(name string) func() {
		return func() {
			ew.SelectedFile = "/home/stepa/lab_db/storage/" + name
			fmt.Printf("Selected file: %s\n", ew.SelectedFile)

			if ew.OnFileSelect != nil {
				ew.OnFileSelect(ew.SelectedFile)
			}
		}
	}(fileName))

	fileButton.Resize(fyne.NewSize(400, 50))
	ew.fileListContainer.Add(fileButton)
	ew.fileListContainer.Refresh()
}
