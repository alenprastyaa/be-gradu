package utils

import (
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ImportRow struct {
	RowNumber      int
	Name           string
	ClassName      string
	Major          string
	WhatsappNumber string
	Email          string
}

type TeacherImportRow struct {
	RowNumber      int
	Name           string
	Position       string
	WhatsappNumber string
	Email          string
}

func ParseImportFile(file multipart.File, filename string) ([]ImportRow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".csv" {
		return parseCSV(file)
	}
	if ext == ".xlsx" || ext == ".xls" {
		return parseExcel(file)
	}
	return nil, errors.New("file harus berformat CSV atau Excel")
}

func ParseTeacherImportFile(file multipart.File, filename string) ([]TeacherImportRow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".csv" {
		return parseTeacherCSV(file)
	}
	if ext == ".xlsx" || ext == ".xls" {
		return parseTeacherExcel(file)
	}
	return nil, errors.New("file harus berformat CSV atau Excel")
}

func parseCSV(reader io.Reader) ([]ImportRow, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return rowsFromRecords(records)
}

func parseExcel(reader io.Reader) ([]ImportRow, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("file Excel tidak memiliki sheet")
	}
	records, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	return rowsFromRecords(records)
}

func parseTeacherCSV(reader io.Reader) ([]TeacherImportRow, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return teacherRowsFromRecords(records)
}

func parseTeacherExcel(reader io.Reader) ([]TeacherImportRow, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("file Excel tidak memiliki sheet")
	}
	records, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	return teacherRowsFromRecords(records)
}

func rowsFromRecords(records [][]string) ([]ImportRow, error) {
	if len(records) < 2 {
		return nil, errors.New("file import harus memiliki header dan minimal satu data")
	}
	headers := map[string]int{}
	for i, h := range records[0] {
		headers[strings.ToLower(strings.TrimSpace(h))] = i
	}
	required := []string{"name", "class_name", "major", "whatsapp_number"}
	for _, key := range required {
		if _, ok := headers[key]; !ok {
			return nil, errors.New("kolom wajib tidak ditemukan: " + key)
		}
	}
	var rows []ImportRow
	for i, record := range records[1:] {
		if isEmptyImportRecord(record) || isImportHelperRecord(record, headers) {
			continue
		}
		rows = append(rows, ImportRow{
			RowNumber:      i + 2,
			Name:           getCell(record, headers["name"]),
			ClassName:      getCell(record, headers["class_name"]),
			Major:          getCell(record, headers["major"]),
			WhatsappNumber: getCell(record, headers["whatsapp_number"]),
			Email:          getOptionalCell(record, headers, "email"),
		})
	}
	return rows, nil
}

func teacherRowsFromRecords(records [][]string) ([]TeacherImportRow, error) {
	if len(records) < 2 {
		return nil, errors.New("file import harus memiliki header dan minimal satu data")
	}
	headers := map[string]int{}
	for i, h := range records[0] {
		headers[strings.ToLower(strings.TrimSpace(h))] = i
	}
	required := []string{"name", "position"}
	for _, key := range required {
		if _, ok := headers[key]; !ok {
			return nil, errors.New("kolom wajib tidak ditemukan: " + key)
		}
	}
	var rows []TeacherImportRow
	for i, record := range records[1:] {
		if isEmptyImportRecord(record) || isTeacherImportHelperRecord(record, headers) {
			continue
		}
		rows = append(rows, TeacherImportRow{
			RowNumber:      i + 2,
			Name:           getCell(record, headers["name"]),
			Position:       getCell(record, headers["position"]),
			WhatsappNumber: getOptionalCell(record, headers, "whatsapp_number"),
			Email:          getOptionalCell(record, headers, "email"),
		})
	}
	return rows, nil
}

func isEmptyImportRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func isImportHelperRecord(record []string, headers map[string]int) bool {
	name := strings.ToLower(getCell(record, headers["name"]))
	className := strings.ToLower(getCell(record, headers["class_name"]))
	major := strings.ToLower(getCell(record, headers["major"]))
	whatsapp := strings.ToLower(getCell(record, headers["whatsapp_number"]))
	return name == "nama lengkap siswa" &&
		strings.HasPrefix(className, "kelas") &&
		strings.HasPrefix(major, "jurusan") &&
		strings.Contains(whatsapp, "whatsapp")
}

func isTeacherImportHelperRecord(record []string, headers map[string]int) bool {
	name := strings.ToLower(getCell(record, headers["name"]))
	position := strings.ToLower(getCell(record, headers["position"]))
	whatsapp := strings.ToLower(getOptionalCell(record, headers, "whatsapp_number"))
	return name == "nama lengkap guru" &&
		strings.HasPrefix(position, "jabatan") &&
		(strings.Contains(whatsapp, "whatsapp") || whatsapp == "")
}

func getCell(record []string, index int) string {
	if index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func getOptionalCell(record []string, headers map[string]int, key string) string {
	index, ok := headers[key]
	if !ok {
		return ""
	}
	return getCell(record, index)
}
