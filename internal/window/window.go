package window

import (
	"fmt"
	"lab_db/internal/lib"
	"lab_db/internal/record"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Window struct {
	App               fyne.App
	MainWindow        fyne.Window
	DbFile            string
	IDEntry           *widget.Entry
	NameEntry         *widget.Entry
	SalaryEntry       *widget.Entry
	WorkingSinceEntry *widget.Entry
	IsOfficialEntry   *widget.Entry
	ErrorLabel        *widget.Label
	Table             *widget.Table
	TableRows         [][]string
	HighlightedRow    []int
}

func New(app fyne.App, filepath string) *Window {
	win := &Window{
		App:            app,
		DbFile:         filepath,
		HighlightedRow: []int{},
	}

	win.SetWidgets()

	return win
}

func (w *Window) ClearFields() {
	w.IDEntry.SetText("")
	w.NameEntry.SetText("")
	w.SalaryEntry.SetText("")
	w.WorkingSinceEntry.SetText("")
	w.IsOfficialEntry.SetText("")
}

func (w *Window) SetColWidth() {
	w.Table.SetColumnWidth(0, 100) // ID
	w.Table.SetColumnWidth(1, 200) // Name
	w.Table.SetColumnWidth(2, 100) // Salary
	w.Table.SetColumnWidth(3, 150) // Work Since
	w.Table.SetColumnWidth(4, 120) // Is official
}

func (w *Window) SetWidgets() {

	w.IDEntry = widget.NewEntry()
	w.IDEntry.SetPlaceHolder("Enter employee's id")

	w.NameEntry = widget.NewEntry()
	w.NameEntry.SetPlaceHolder("Enter employee's name")

	w.SalaryEntry = widget.NewEntry()
	w.SalaryEntry.SetPlaceHolder("Enter employee's salary")

	w.WorkingSinceEntry = widget.NewEntry()
	w.WorkingSinceEntry.SetPlaceHolder("Enter employee's date of working format: yyyy-mm-dd")

	w.IsOfficialEntry = widget.NewEntry()
	w.IsOfficialEntry.SetPlaceHolder("Enter true/false if employee is officially hired")

	w.ErrorLabel = widget.NewLabel("")
	w.ErrorLabel.TextStyle = fyne.TextStyle{Bold: true}
	w.ErrorLabel.Alignment = fyne.TextAlignCenter
	w.ErrorLabel.Hide()

	w.IDEntry.Resize(fyne.NewSize(400, 30))
	w.NameEntry.Resize(fyne.NewSize(400, 30))
	w.SalaryEntry.Resize(fyne.NewSize(400, 30))
	w.WorkingSinceEntry.Resize(fyne.NewSize(400, 30))
	w.IsOfficialEntry.Resize(fyne.NewSize(400, 30))

	w.addExistingRows()

	w.Table = widget.NewTable(
		func() (int, int) {
			return len(w.TableRows), len(w.TableRows[0])
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignCenter // Центрирование текста
			label.Resize(fyne.NewSize(150, 30))    // Ширина и высота ячейки
			return label
		},
		func(cell widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText(w.TableRows[cell.Row][cell.Col])
			label.Alignment = fyne.TextAlignCenter

			if cell.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true} // Заголовок таблицы
			} else if lib.Contains(cell.Row, w.HighlightedRow) {
				label.TextStyle = fyne.TextStyle{Bold: true} // Выделенная строка
			} else {
				label.TextStyle = fyne.TextStyle{} // Обычный стиль
			}
		},
	)

	w.SetColWidth()
}

func (w *Window) addExistingRows() {
	w.TableRows = [][]string{}
	w.TableRows = append(w.TableRows, []string{"id", "name", "salary", "work_since", "is_official"})
	rowsRecord, err := record.GetRows(w.DbFile)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or ") {
			log.Fatal(fmt.Sprintf("Failed to load rows %v", err))
		}
	}

	for _, val := range rowsRecord {
		w.TableRows = append(w.TableRows, lib.RecordToString(val))
	}
}

/*
	func (w *Window) CreateDatabase(filename string, rec record.Record) {
		err := record.CreateDatabase("home/stepa/lab_db/storage" + filename)
		if err != nil {
			log.Fatalf("Error creating database: %v", err)
		}
	}
*/
func (w *Window) AddRecord() {
	defer w.ClearFields()

	clearError := func() {
		w.ErrorLabel.SetText("")
		w.ErrorLabel.Hide()
	}

	showError := func(msg string) {
		w.ErrorLabel.SetText(msg)
		w.ErrorLabel.Show()
	}

	clearError()

	id, err := strconv.Atoi(w.IDEntry.Text)
	if err != nil {
		showError("Invalid ID")
		return
	}

	name := w.NameEntry.Text

	salary, err := strconv.ParseFloat(w.SalaryEntry.Text, 64)
	if err != nil || salary <= 0.0 {
		showError("Invalid value for Salary")
		return
	}

	workingSince, err := time.Parse("2006-01-02", w.WorkingSinceEntry.Text)
	if err != nil {
		showError("Invalid date format")
		return
	}

	isOfficial, err := strconv.ParseBool(w.IsOfficialEntry.Text)
	if err != nil || len(w.IsOfficialEntry.Text) == 0 {
		showError("Invalid data for official field")
		return
	}

	rec := record.Record{
		ID:           int32(id),
		Name:         name,
		Salary:       salary,
		WorkingSince: workingSince,
		IsOfficial:   isOfficial,
	}
	rec.Print()

	if err = rec.AddRecord(w.DbFile); err != nil {
		showError("Error adding record")
		fmt.Println(fmt.Sprintf("%v", err))
		return
	}

	showError("Record added successfully!")
	w.ErrorLabel.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
	w.addExistingRows()

}

func (w *Window) Show(windowName string) {
	w.MainWindow = w.App.NewWindow(windowName)

	xlsxButton := widget.NewButton("Create XLSX file", func() {
		filename := "table" + time.Now().String() + ".xlsx"
		err := lib.ExportToXLSX(w.DbFile)
		if err != nil {
			dialog.ShowError(err, w.MainWindow)
			return
		}
		dialog.ShowInformation("Success", "Excel filed saved as "+filename, w.MainWindow)
	})

	backupButton := widget.NewButton("Create backup file", func() {
		err := lib.CreateBackup(w.DbFile)
		if err != nil {
			dialog.ShowError(err, w.MainWindow)
			return
		}
		dialog.ShowInformation("Success", "Backup file saved", w.MainWindow)

	})

	addButton := widget.NewButton("Add Record", func() {
		w.AddRecord()
	})

	truncateButton := widget.NewButton("Truncate table", func() {
		os.Create(w.DbFile)
		w.addExistingRows()
		w.Table.Refresh()
		w.Table.Refresh()
	})

	deleteTableButton := widget.NewButton("Delete Table", func() {
		os.Remove(w.DbFile)
		w.MainWindow.Close()
	})

	buttonsRow := container.NewHBox(
		xlsxButton,
		backupButton,
		truncateButton,
		deleteTableButton,
	)

	searchByIDEntry := widget.NewEntry()
	searchByIDEntry.SetPlaceHolder("Enter employee's ID")
	searchByIDEntry.Resize(fyne.NewSize(400, 30))

	searchByNameEntry := widget.NewEntry()
	searchByNameEntry.SetPlaceHolder("Entry employee's Name")
	searchByNameEntry.Resize(fyne.NewSize(400, 30))

	searchBySalaryEntry := widget.NewEntry()
	searchBySalaryEntry.SetPlaceHolder("Entry employee's salary")
	searchBySalaryEntry.Resize(fyne.NewSize(400, 30))

	searchByDateEntry := widget.NewEntry()
	searchByDateEntry.SetPlaceHolder("Entry employee's work_since date")
	searchByDateEntry.Resize(fyne.NewSize(400, 30))

	searchByIsOfficialEntry := widget.NewEntry()
	searchByIsOfficialEntry.SetPlaceHolder("Entry employee's is_official")
	searchByIsOfficialEntry.Resize(fyne.NewSize(400, 30))

	findButton := widget.NewButton("Find employee", func() {
		defer w.Table.Refresh()
		defer w.Table.Refresh()
		defer func() {
			searchByIDEntry.SetText("")
			searchByNameEntry.SetText("")
			searchBySalaryEntry.SetText("")
			searchByDateEntry.SetText("")
			searchByIsOfficialEntry.SetText("")
		}()
		searchID, err := strconv.Atoi(searchByIDEntry.Text)
		if err == nil {
			ind, err := record.FindRecordByID(int32(searchID), w.DbFile)
			if err != nil {
				fmt.Printf("Failed to find employee: %v", err)

				w.HighlightedRow = []int{}

				return
			}

			w.HighlightedRow = []int{ind}

			return
		}

		searchName := searchByNameEntry.Text
		nameIdx := record.FindAllName(searchName, w.DbFile)

		w.HighlightedRow = nameIdx

		if len(w.HighlightedRow) != 0 {
			return
		}

		searchSalary, err := strconv.ParseFloat(searchBySalaryEntry.Text, 64)
		if err == nil {
			salaryFound := record.FindAllSalary(searchSalary, w.DbFile)

			w.HighlightedRow = salaryFound

			return
		}

		searchDate, err := time.Parse("2006-01-02", searchByDateEntry.Text)
		if err == nil {
			datesFound := record.FindAllDate(searchDate, w.DbFile)

			w.HighlightedRow = datesFound

			return
		}

		isOfficial, err := strconv.ParseBool(searchByIsOfficialEntry.Text)
		if err == nil {
			isOfficialFound := record.FindAllOfficial(isOfficial, w.DbFile)

			w.HighlightedRow = isOfficialFound
			return
		}

	})

	deleteByIDEntry := widget.NewEntry()
	deleteByIDEntry.SetPlaceHolder("Enter employee ID to delete")

	deleteByNameEntry := widget.NewEntry()
	deleteByNameEntry.SetPlaceHolder("Enter employee Name to delete")

	deleteBySalaryEntry := widget.NewEntry()
	deleteBySalaryEntry.SetPlaceHolder("Enter employee Salary to delete")

	deleteByDateEntry := widget.NewEntry()
	deleteByDateEntry.SetPlaceHolder("Enter employee work_since to delete")

	deleteByOfficialEntry := widget.NewEntry()
	deleteByOfficialEntry.SetPlaceHolder("Enter employee is_official to delete")

	deleteButton := widget.NewButton("Delete employee", func() {
		defer w.Table.Refresh()
		defer w.addExistingRows()
		defer func() {
			deleteByIDEntry.SetText("")
			deleteByNameEntry.SetText("")
			deleteBySalaryEntry.SetText("")
			deleteByDateEntry.SetText("")
			deleteByOfficialEntry.SetText("")
		}()

		id, err := strconv.Atoi(deleteByIDEntry.Text)
		if err == nil {
			err = record.DeleteByID(id, w.DbFile)
			if err == nil {
				return
			}
			return
		}

		name := deleteByNameEntry.Text
		if len(name) != 0 {
			nameIndexes := record.FindAllName(name, w.DbFile)
			record.DeleteByName(nameIndexes, w.DbFile)
			return
		}

		salary, err := strconv.ParseFloat(deleteBySalaryEntry.Text, 64)
		if err == nil {
			salaryIndexes := record.FindAllSalary(salary, w.DbFile)
			record.DeleteBySalary(salaryIndexes, w.DbFile)
			return
		}

		date, err := time.Parse("2006-01-02", deleteByDateEntry.Text)
		if err == nil {
			dateIndexes := record.FindAllDate(date, w.DbFile)
			record.DeleteByDate(dateIndexes, w.DbFile)
			return
		}

		isOfficial, err := strconv.ParseBool(deleteByOfficialEntry.Text)
		if err == nil {
			dateIndexes := record.FindAllOfficial(isOfficial, w.DbFile)
			record.DeleteByOfficial(dateIndexes, w.DbFile)
			return
		}

	})

	alterByID := widget.NewEntry()
	alterByID.SetPlaceHolder("Enter ID to modify")

	alterByName := widget.NewEntry()
	alterByName.SetPlaceHolder("Enter name to modify")

	alterBySalary := widget.NewEntry()
	alterBySalary.SetPlaceHolder("Enter salary to modify")

	alterByDate := widget.NewEntry()
	alterByDate.SetPlaceHolder("Enter date to modify")

	alterByIsOfficial := widget.NewEntry()
	alterByIsOfficial.SetPlaceHolder("Enter is_official to modify")

	enterNameToChange := widget.NewEntry()
	enterNameToChange.SetPlaceHolder("Enter name value")

	enterSalaryToChange := widget.NewEntry()
	enterSalaryToChange.SetPlaceHolder("Enter salary value")

	enterDateToChange := widget.NewEntry()
	enterDateToChange.SetPlaceHolder("Enter date value")

	enterIsOfficialToChange := widget.NewEntry()
	enterIsOfficialToChange.SetPlaceHolder("Enter is_official value")

	changeButton := widget.NewButton("Change data", func() {
		defer w.Table.Refresh()
		defer w.Table.Refresh()
		defer w.addExistingRows()

		changeFunc := func(recordsToChange []record.Record) {
			isEntered := 1

			newRecordName := enterNameToChange.Text

			newRecordSalary, err := strconv.ParseFloat(enterSalaryToChange.Text, 64)
			if err != nil {
				newRecordSalary = -0.1
			}

			newRecordDate, err := time.Parse("2006-01-02", enterDateToChange.Text)
			if err != nil {
				newRecordDate = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			newRecordIsOfficial, err := strconv.ParseBool(enterIsOfficialToChange.Text)
			if err != nil {
				newRecordIsOfficial = false
				isEntered = 0
			}

			for i := 0; i < len(recordsToChange); i++ {
				if len(newRecordName) != 0 {
					recordsToChange[i].Name = newRecordName
				}

				if newRecordSalary > 0.0 {
					recordsToChange[i].Salary = newRecordSalary
				}

				if !strings.Contains(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).String(), newRecordDate.String()) {
					recordsToChange[i].WorkingSince = newRecordDate
				}
				if isEntered == 1 {
					recordsToChange[i].IsOfficial = newRecordIsOfficial
				}
			}

			err = record.ChangeRecords(recordsToChange, w.DbFile)

			if err != nil {
				fmt.Printf("%v", err)
			}

			return
		}

		id, err := strconv.Atoi(alterByID.Text)
		if err == nil && id > 0 {
			newRecord, err := record.FindRecordByIDToModify(int32(id), w.DbFile)
			if err != nil {
				return
			}

			if len(enterNameToChange.Text) != 0 {
				newRecord.Name = enterNameToChange.Text
			}

			err = record.ChangeRecord(newRecord, w.DbFile)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}

		}

		name := alterByName.Text

		if len(name) != 0 {
			recordsToChange, err := record.FindRecordsByNameToModify(name, w.DbFile)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}
			changeFunc(recordsToChange)
		}

		salary, err := strconv.ParseFloat(alterBySalary.Text, 64)
		if err == nil {
			recordsToChange, err := record.FindRecordsBySalaryToModify(salary, w.DbFile)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}

			changeFunc(recordsToChange)
		}

		date, err := time.Parse("2006-01-02", alterByDate.Text)
		if err == nil {
			recordsToChange, err := record.FindRecordsByDateToModify(date, w.DbFile)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}

			changeFunc(recordsToChange)
		}

		isOffic, err := strconv.ParseBool(alterByIsOfficial.Text)
		if err == nil {
			recordsToChange, err := record.FindRecordsByIsOfficialToModify(isOffic, w.DbFile)
			if err != nil {
				fmt.Printf("%v", err)
				return
			}

			changeFunc(recordsToChange)
		}

	})

	//addButton.Resize(fyne.NewSize(300, 30))
	alterByContainer := container.NewVBox(
		alterByID,
		alterByName,
		alterBySalary,
		alterByDate,
		alterByIsOfficial,
	)

	alterValueContainer := container.NewVBox(
		enterNameToChange,
		enterSalaryToChange,
		enterDateToChange,
		enterIsOfficialToChange,
	)

	inputRow := container.NewVBox(
		w.IDEntry,
		w.NameEntry,
		w.SalaryEntry,
		w.WorkingSinceEntry,
		w.IsOfficialEntry,
		addButton,
	)

	findContainer := container.NewVBox(
		searchByIDEntry,
		searchByNameEntry,
		searchBySalaryEntry,
		searchByDateEntry,
		searchByIsOfficialEntry,
		findButton,
	)

	deleteContainer := container.NewVBox(
		deleteByIDEntry,
		deleteByNameEntry,
		deleteBySalaryEntry,
		deleteByDateEntry,
		deleteByOfficialEntry,
		deleteButton,
	)

	findAddRecordContainer := container.New(
		layout.NewGridLayout(3),

		inputRow,
		findContainer,
		deleteContainer,
	)

	alterRecordContainer := container.New(
		layout.NewGridLayout(2),

		alterByContainer,
		alterValueContainer,
	)

	scrollableTable := container.NewVScroll(w.Table)
	scrollableTable.SetMinSize(fyne.NewSize(400, 200))

	content := container.NewVBox(
		widget.NewLabel("File Database Management"),
		buttonsRow,
		w.ErrorLabel,
		findAddRecordContainer,
		//inputRow,
		//widget.NewLabel("Find Employee by ID:"),
		//findContainer,
		widget.NewLabel("Database Records:"),
		container.NewStack(scrollableTable),
		alterRecordContainer,
		changeButton,
	)

	w.MainWindow.SetContent(content)
	w.MainWindow.Resize(fyne.NewSize(1200, 800))
	w.MainWindow.Show()
}
