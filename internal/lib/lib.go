package lib

import (
	"encoding/csv"
	"fmt"
	"lab_db/internal/record"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func Contains(value int, src []int) bool {
	for _, val := range src {
		if value == val {
			return true
		}
	}

	return false
}

func RecordToString(rec record.Record) []string {
	return []string{
		strconv.Itoa(int(rec.ID)),
		string(rec.Name[:]),
		fmt.Sprintf("%.2f", rec.Salary),
		rec.WorkingSince.Format("2006-01-02"), // Корректный формат даты
		strconv.FormatBool(rec.IsOfficial),
	}
}

func ExportToXLSX(filepath string) error {
	filename := "Backup_" + time.Now().Format("2006-01-02_15-04-05") + ".xlsx"

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	recordRows, err := record.GetRows(filepath)
	if err != nil {
		return err
	}

	recordsToUpload := make([][]string, 0)
	recordsToUpload = append(recordsToUpload, []string{"id", "name", "salary", "working_since", "is_official"})

	for _, val := range recordRows {
		recordsToUpload = append(recordsToUpload, RecordToString(val))
	}

	for i, row := range recordsToUpload {
		for j, cell := range row {
			cellName, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue("Sheet1", cellName, cell)
		}
	}

	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil

}

func CreateBackup(filepath string) error {
	name := strings.Split(filepath, "/")

	filename := "/home/stepa/lab_db/backup/" + name[len(name)-1] + ".csv"

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	tableRows, err := record.GetRows(filepath)
	if err != nil {
		return err
	}

	recordsToUpload := make([][]string, 0)
	recordsToUpload = append(recordsToUpload, []string{"id", "name", "salary", "working_since", "is_official"})

	for _, val := range tableRows {
		recordsToUpload = append(recordsToUpload, RecordToString(val))
	}

	for _, row := range recordsToUpload {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()

	return nil
}

func CreateFromBackup(filename string) (string, error) {
	filepath := "/home/stepa/lab_db/storage/backup_" + strings.Replace(filename, ".db.csv", "", 1) + ".db"
	fmt.Println(filepath)

	dbFile, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	dbFile.Sync()

	defer dbFile.Close()

	csvFile, err := os.Open("/home/stepa/lab_db/backup/" + filename)
	if err != nil {
		return "sdad", err
	}

	reader := csv.NewReader(csvFile)

	rows, err := reader.ReadAll()
	if err != nil {
		return "", nil
	}

	for _, row := range rows[1:] {
		id, _ := strconv.Atoi(row[0])
		salary, _ := strconv.ParseFloat(row[2], 64)
		wokingSince, _ := time.Parse("2006-01-02", row[3])
		isOfficial, _ := strconv.ParseBool(row[4])

		newRec := record.Record{
			ID:           int32(id),
			Name:         row[1],
			Salary:       salary,
			WorkingSince: wokingSince,
			IsOfficial:   isOfficial,
		}
		fmt.Println(newRec)
		buf, err := record.SerializeRecord(&newRec)
		if err != nil {
			return "", err
		}

		if _, err := dbFile.Write(buf); err != nil {
			return "", err
		}
	}

	dbFile.Sync()

	return filepath, err
}
