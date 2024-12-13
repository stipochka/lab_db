package record

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Record struct {
	ID           int32
	Name         string
	Salary       float64
	WorkingSince time.Time
	IsOfficial   bool
}

func (r *Record) Print() {
	nameStr := string(r.Name[:])
	nameStr = strings.TrimSpace(nameStr)

	fmt.Printf("ID: %d, Name: %s, Salary: %.2f, Work Experience: %s, Official: %v\n",
		r.ID, nameStr, r.Salary, r.WorkingSince.Format("2006-01-02"), r.IsOfficial)
}

const recordSize = 53

func CreateDatabase(filename string) (string, error) {
	filenameWithPath := "/home/stepa/lab_db/storage/" + filename

	if !strings.Contains(filenameWithPath, ".db") {
		filenameWithPath += ".db"
	}

	if _, err := os.Create(filenameWithPath); err != nil {
		return "", err
	}

	return filenameWithPath, nil
}

func GetRows(filepath string) ([]Record, error) {
	recordsList := make([]Record, 0)
	file, err := os.Open(filepath)
	if err != nil {
		return recordsList[:0], err
	}
	defer file.Close()

	buf := make([]byte, recordSize)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return recordsList[:0], err
		}
		if n < recordSize {
			break
		}

		record, err := DeserializeRecord(buf)
		if err != nil {
			return recordsList[:0], err
		}

		recordsList = append(recordsList, record)
	}
	return recordsList, nil
}

func SerializeRecord(rec *Record) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Запись ID
	if err := binary.Write(buf, binary.LittleEndian, rec.ID); err != nil {
		return nil, fmt.Errorf("error writing ID: %w", err)
	}

	// Имя фиксированной длины (32 символа)
	for len(rec.Name) < 32 {
		rec.Name += " "
	}

	name := fmt.Sprintf("%-32s", []byte(rec.Name))
	if err := binary.Write(buf, binary.LittleEndian, []byte(name)); err != nil {
		return nil, fmt.Errorf("error writing Name: %w", err)
	}

	// Зарплата
	if err := binary.Write(buf, binary.LittleEndian, rec.Salary); err != nil {
		return nil, fmt.Errorf("error writing Salary: %w", err)
	}

	// Время
	if err := binary.Write(buf, binary.LittleEndian, rec.WorkingSince.Unix()); err != nil {
		return nil, fmt.Errorf("error writing WorkingSince: %w", err)
	}

	// Официальный статус
	if err := binary.Write(buf, binary.LittleEndian, rec.IsOfficial); err != nil {
		return nil, fmt.Errorf("error writing IsOfficial: %w", err)
	}

	return buf.Bytes(), nil
}

func DeserializeRecord(data []byte) (Record, error) {
	buf := bytes.NewReader(data)
	var record Record

	// Чтение ID
	if err := binary.Read(buf, binary.LittleEndian, &record.ID); err != nil {
		return Record{}, fmt.Errorf("error reading ID: %w", err)
	}

	// Чтение имени
	nameBytes := make([]byte, 32)
	if err := binary.Read(buf, binary.LittleEndian, &nameBytes); err != nil {
		return Record{}, fmt.Errorf("error reading Name: %w", err)
	}
	record.Name = strings.TrimRight(string(nameBytes), "\x00")

	// Чтение зарплаты
	if err := binary.Read(buf, binary.LittleEndian, &record.Salary); err != nil {
		return Record{}, fmt.Errorf("error reading Salary: %w", err)
	}

	// Чтение даты
	var unixTime int64
	if err := binary.Read(buf, binary.LittleEndian, &unixTime); err != nil {
		return Record{}, fmt.Errorf("error reading WorkingSince: %w", err)
	}
	record.WorkingSince = time.Unix(unixTime, 0)

	// Чтение статуса
	if err := binary.Read(buf, binary.LittleEndian, &record.IsOfficial); err != nil {
		return Record{}, fmt.Errorf("error reading IsOfficial: %w", err)
	}

	return record, nil
}

func FindRecordByID(id int32, filepath string) (int, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return -1, fmt.Errorf("error getting file info: %w", err)
	}

	totalRecords := int(fileInfo.Size() / recordSize)
	low, high := 0, totalRecords-1

	buf := make([]byte, recordSize)
	for low <= high {
		mid := low + (high-low)/2

		offset := int64(mid * recordSize)
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return -1, fmt.Errorf("error seeking file: %w", err)
		}
		if _, err := file.Read(buf); err != nil {
			return -1, fmt.Errorf("error reading record: %w", err)
		}

		midRecord, err := DeserializeRecord(buf)
		if err != nil {
			return -1, fmt.Errorf("error deserializing record")
		}

		if midRecord.ID == id {
			return mid + 1, nil
		} else if midRecord.ID < id {
			low = mid + 1
		} else {
			high = mid - 1
		}

	}
	return -1, errors.New("record not found")
}

func (rec *Record) AddRecord(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := SerializeRecord(rec)
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	totalRecords := int(fileInfo.Size() / recordSize)
	low, high := 0, totalRecords-1
	insertPos := 0

	buf := make([]byte, recordSize)
	for low <= high {
		mid := low + (high-low)/2

		offset := int64(mid * recordSize)
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("error seeking file: %w", err)
		}
		if _, err := file.Read(buf); err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}

		midRecord, err := DeserializeRecord(buf)
		if err != nil {
			return fmt.Errorf("error deserializing record")
		}

		if midRecord.ID == rec.ID {
			return fmt.Errorf("Row with that ID already exists")
		} else if midRecord.ID < rec.ID {
			low = mid + 1
		} else {
			high = mid - 1
		}

	}

	insertPos = low

	// Если вставка не в конец, сдвигаем последующие записи
	if insertPos < totalRecords {
		if err := shiftRecords(file, insertPos, totalRecords); err != nil {
			return fmt.Errorf("error shifting records: %w", err)
		}
	}

	// Позиционируем указатель записи
	if _, err := file.Seek(int64(insertPos*recordSize), io.SeekStart); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}

	// Пишем запись в файл
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("error writing record: %w", err)
	}

	// Синхронизируем буфер с диском
	if err := file.Sync(); err != nil {
		return fmt.Errorf("error syncing file: %w", err)
	}

	return nil
}

func shiftRecords(file *os.File, insertPos int, totalRecords int) error {
	buf := make([]byte, recordSize)

	for i := totalRecords - 1; i >= insertPos; i-- {
		srcOffset := int64(i * recordSize)
		destOffset := int64((i + 1) * recordSize)

		if _, err := file.Seek(srcOffset, io.SeekStart); err != nil {
			return fmt.Errorf("error seeking file: %w", err)
		}
		if _, err := file.Read(buf); err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}

		if _, err := file.Seek(destOffset, io.SeekStart); err != nil {
			return fmt.Errorf("error seeking file: %w", err)
		}

		if _, err := file.Write(buf); err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}

	// Синхронизация для гарантии записи на диск
	if err := file.Sync(); err != nil {
		return fmt.Errorf("error syncing file during shift: %w", err)
	}

	return nil
}

func DeleteByID(id int, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	rowInd, err := FindRecordByID(int32(id), filepath)
	if err != nil {
		return err
	}
	rowInd -= 1

	totalRecords := int(fileInfo.Size() / recordSize)

	buf := make([]byte, recordSize)

	for i := rowInd; i < totalRecords-1; i++ {
		srcOffset := int64((i + 1) * recordSize)
		dstOffset := int64(i * recordSize)

		if _, err := file.Seek(srcOffset, io.SeekStart); err != nil {
			return fmt.Errorf("error seeking source record: %w", err)
		}
		if _, err := file.Read(buf); err != nil {
			return fmt.Errorf("error reading source record: %w", err)
		}

		if _, err := file.Seek(dstOffset, io.SeekStart); err != nil {
			return fmt.Errorf("error seeking destination: %w", err)
		}
		if _, err := file.Write(buf); err != nil {
			return fmt.Errorf("error writing destination record: %w", err)
		}
	}

	newSize := int64((totalRecords - 1) * recordSize)
	if err := file.Truncate(newSize); err != nil {
		return fmt.Errorf("error truncating file: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("error syncing file: %w", err)
	}

	return nil
}

func DeleteByName(namesInd []int, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	for _, val := range namesInd {
		fileInfo, err := file.Stat()

		if err != nil {
			return err
		}

		totalRecords := int(fileInfo.Size() / recordSize)

		buf := make([]byte, recordSize)

		for i := val - 1; i < totalRecords-1; i++ {
			srcOffset := int64((i + 1) * recordSize)
			dstOffset := int64(i * recordSize)

			if _, err := file.Seek(srcOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking source record: %w", err)
			}
			if _, err := file.Read(buf); err != nil {
				return fmt.Errorf("error reading source record: %w", err)
			}

			if _, err := file.Seek(dstOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking destination: %w", err)
			}
			if _, err := file.Write(buf); err != nil {
				return fmt.Errorf("error writing destination record: %w", err)
			}
		}

		newSize := int64((totalRecords - 1) * recordSize)
		if err := file.Truncate(newSize); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}

		if err := file.Sync(); err != nil {
			return fmt.Errorf("error syncing file: %w", err)
		}

	}
	return nil
}

func FindRecordByIDToModify(id int32, filepath string) (*Record, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return &Record{}, err
	}

	recordDst, err := FindRecordByID(id, filepath)
	if err != nil {
		return &Record{}, err
	}

	recordDst -= 1

	buf := make([]byte, recordSize)

	if _, err := file.Seek(int64(recordSize*recordDst), io.SeekStart); err != nil {
		return &Record{}, err
	}

	if _, err := file.Read(buf); err != nil {
		return &Record{}, err
	}

	rec, err := DeserializeRecord(buf)
	if err != nil {
		return &Record{}, err
	}

	return &rec, nil
}

func FindAllName(name string, filepath string) []int {
	names := make([]int, 0)

	records, err := GetRows(filepath)
	if err != nil {
		return names
	}

	for ind, val := range records {
		if strings.TrimSpace(val.Name) == strings.TrimSpace(name) {
			names = append(names, ind+1)
		}
	}

	return names
}

func FindAllSalary(salary float64, filepath string) []int {
	salaries := make([]int, 0)

	records, err := GetRows(filepath)
	if err != nil {
		return salaries
	}

	for ind, val := range records {
		if fmt.Sprintf("%0.2f", val.Salary) == fmt.Sprintf("%0.2f", salary) {
			salaries = append(salaries, ind+1)
		}
	}

	return salaries
}

func FindAllDate(date time.Time, filepath string) []int {
	dates := make([]int, 0)

	records, err := GetRows(filepath)
	if err != nil {
		return dates
	}

	for ind, val := range records {
		formatDate := strings.Split(val.WorkingSince.String(), " ")
		if strings.Contains(date.String(), formatDate[0]) {
			dates = append(dates, ind+1)
		}
	}

	return dates
}

func FindAllOfficial(isOfficial bool, filepath string) []int {
	found := make([]int, 0)

	records, err := GetRows(filepath)
	if err != nil {
		return found
	}

	for ind, val := range records {
		if val.IsOfficial == isOfficial {
			found = append(found, ind+1)
		}
	}

	return found

}

func DeleteBySalary(salaryInd []int, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	for _, val := range salaryInd {
		fileInfo, err := file.Stat()

		if err != nil {
			return err
		}

		totalRecords := int(fileInfo.Size() / recordSize)

		buf := make([]byte, recordSize)

		for i := val - 1; i < totalRecords-1; i++ {
			srcOffset := int64((i + 1) * recordSize)
			dstOffset := int64(i * recordSize)

			if _, err := file.Seek(srcOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking source record: %w", err)
			}
			if _, err := file.Read(buf); err != nil {
				return fmt.Errorf("error reading source record: %w", err)
			}

			if _, err := file.Seek(dstOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking destination: %w", err)
			}
			if _, err := file.Write(buf); err != nil {
				return fmt.Errorf("error writing destination record: %w", err)
			}
		}

		newSize := int64((totalRecords - 1) * recordSize)
		if err := file.Truncate(newSize); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}

		if err := file.Sync(); err != nil {
			return fmt.Errorf("error syncing file: %w", err)
		}

	}
	return nil
}

func DeleteByDate(dateInd []int, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	for _, val := range dateInd {
		fileInfo, err := file.Stat()

		if err != nil {
			return err
		}

		totalRecords := int(fileInfo.Size() / recordSize)

		buf := make([]byte, recordSize)

		for i := val - 1; i < totalRecords-1; i++ {
			srcOffset := int64((i + 1) * recordSize)
			dstOffset := int64(i * recordSize)

			if _, err := file.Seek(srcOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking source record: %w", err)
			}
			if _, err := file.Read(buf); err != nil {
				return fmt.Errorf("error reading source record: %w", err)
			}

			if _, err := file.Seek(dstOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking destination: %w", err)
			}
			if _, err := file.Write(buf); err != nil {
				return fmt.Errorf("error writing destination record: %w", err)
			}
		}

		newSize := int64((totalRecords - 1) * recordSize)
		if err := file.Truncate(newSize); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}

		if err := file.Sync(); err != nil {
			return fmt.Errorf("error syncing file: %w", err)
		}

	}
	return nil
}

func DeleteByOfficial(officialInd []int, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	for _, val := range officialInd {
		fileInfo, err := file.Stat()

		if err != nil {
			return err
		}

		totalRecords := int(fileInfo.Size() / recordSize)

		buf := make([]byte, recordSize)

		for i := val - 1; i < totalRecords-1; i++ {
			srcOffset := int64((i + 1) * recordSize)
			dstOffset := int64(i * recordSize)

			if _, err := file.Seek(srcOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking source record: %w", err)
			}
			if _, err := file.Read(buf); err != nil {
				return fmt.Errorf("error reading source record: %w", err)
			}

			if _, err := file.Seek(dstOffset, io.SeekStart); err != nil {
				return fmt.Errorf("error seeking destination: %w", err)
			}
			if _, err := file.Write(buf); err != nil {
				return fmt.Errorf("error writing destination record: %w", err)
			}
		}

		newSize := int64((totalRecords - 1) * recordSize)
		if err := file.Truncate(newSize); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}

		if err := file.Sync(); err != nil {
			return fmt.Errorf("error syncing file: %w", err)
		}

	}
	return nil
}

func ChangeRecord(rec *Record, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	recordDst, err := FindRecordByID(rec.ID, filepath)
	if err != nil {
		return err
	}

	buf, err := SerializeRecord(rec)
	if err != nil {
		return err
	}

	if _, err := file.Seek(int64(recordDst-1)*recordSize, io.SeekStart); err != nil {
		return err
	}

	if _, err := file.Write(buf); err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

func FindRecordsByNameToModify(name string, filepath string) ([]Record, error) {
	var records []Record

	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return records, err
	}

	recordsOccur := FindAllName(name, filepath)

	buf := make([]byte, recordSize)

	for _, val := range recordsOccur {

		if _, err := file.Seek(int64((val-1)*recordSize), io.SeekStart); err != nil {
			return records, err
		}

		if _, err := file.Read(buf); err != nil {
			return records, err
		}

		desRecord, err := DeserializeRecord(buf)
		if err != nil {
			return records, err
		}

		records = append(records, desRecord)
	}

	return records, nil
}

func ChangeRecords(records []Record, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	for _, val := range records {
		recordDst, err := FindRecordByID(val.ID, filepath)
		if err != nil {
			return err
		}
		recordDst -= 1

		if _, err = file.Seek(int64(recordDst*recordSize), io.SeekStart); err != nil {
			return err
		}

		buf, err := SerializeRecord(&val)
		if err != nil {
			return err
		}

		if _, err := file.Write(buf); err != nil {
			return err
		}
		file.Sync()
	}

	return nil
}

func FindRecordsBySalaryToModify(salary float64, filepath string) ([]Record, error) {
	var records []Record

	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return records, err
	}

	recordsOccur := FindAllSalary(salary, filepath)

	buf := make([]byte, recordSize)

	for _, val := range recordsOccur {

		if _, err := file.Seek(int64((val-1)*recordSize), io.SeekStart); err != nil {
			return records, err
		}

		if _, err := file.Read(buf); err != nil {
			return records, err
		}

		desRecord, err := DeserializeRecord(buf)
		if err != nil {
			return records, err
		}

		records = append(records, desRecord)
	}

	return records, nil
}

func FindRecordsByDateToModify(date time.Time, filepath string) ([]Record, error) {
	var records []Record

	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return records, err
	}

	recordsOccur := FindAllDate(date, filepath)

	buf := make([]byte, recordSize)

	for _, val := range recordsOccur {

		if _, err := file.Seek(int64((val-1)*recordSize), io.SeekStart); err != nil {
			return records, err
		}

		if _, err := file.Read(buf); err != nil {
			return records, err
		}

		desRecord, err := DeserializeRecord(buf)
		if err != nil {
			return records, err
		}

		records = append(records, desRecord)
	}

	return records, nil
}

func FindRecordsByIsOfficialToModify(isOfficial bool, filepath string) ([]Record, error) {
	var records []Record

	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return records, err
	}

	recordsOccur := FindAllOfficial(isOfficial, filepath)

	buf := make([]byte, recordSize)

	for _, val := range recordsOccur {

		if _, err := file.Seek(int64((val-1)*recordSize), io.SeekStart); err != nil {
			return records, err
		}

		if _, err := file.Read(buf); err != nil {
			return records, err
		}

		desRecord, err := DeserializeRecord(buf)
		if err != nil {
			return records, err
		}

		records = append(records, desRecord)
	}

	return records, nil
}
