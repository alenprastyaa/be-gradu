package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"
)

const seatsPerStudent = 2

type SeatService struct {
	studentRepo       *repositories.StudentRepository
	eventSettingsRepo *repositories.EventSettingsRepository
}

func NewSeatService(studentRepo *repositories.StudentRepository, eventSettingsRepo *repositories.EventSettingsRepository) *SeatService {
	return &SeatService{
		studentRepo:       studentRepo,
		eventSettingsRepo: eventSettingsRepo,
	}
}

func (s *SeatService) GenerateNextSeatNumber(ctx context.Context, className, major string) (string, string, error) {
	count, err := s.studentRepo.CountAll(ctx)
	if err != nil {
		return "", "", err
	}
	return "-", formatSeatRange(count*seatsPerStudent + 1), nil
}

func (s *SeatService) RegenerateAllSeatNumbers(ctx context.Context) error {
	students, err := s.studentRepo.ListAllOrdered(ctx)
	if err != nil {
		return err
	}
	sort.SliceStable(students, func(i, j int) bool {
		a := students[i]
		b := students[j]
		if a.ClassName != b.ClassName {
			return a.ClassName < b.ClassName
		}
		if a.Major != b.Major {
			return a.Major < b.Major
		}
		return a.Name < b.Name
	})

	students, err = s.orderStudentsByActiveSeatLayout(ctx, students)
	if err != nil {
		return err
	}

	counters := map[string]int{}
	for _, student := range students {
		key := "global"
		counters[key] += seatsPerStudent
		if err := s.studentRepo.UpdateSeat(ctx, student.ID, "-", formatSeatRange(counters[key]-seatsPerStudent+1)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SeatService) orderStudentsByActiveSeatLayout(ctx context.Context, students []models.Student) ([]models.Student, error) {
	if len(students) == 0 || s.eventSettingsRepo == nil {
		return students, nil
	}

	settings, err := s.eventSettingsRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil || strings.TrimSpace(settings.SeatMapLayout) == "" {
		return students, nil
	}

	pairOrder := parseSeatLayoutPairOrder(settings.SeatMapLayout, settings.SeatMapColumns)
	if len(pairOrder) == 0 {
		return students, nil
	}

	sortStudentsByMajorAndName(students, pairOrder)

	nextLayout, err := buildAlphabeticalSeatLayout(settings.SeatMapLayout, settings.SeatMapColumns, students)
	if err != nil {
		return nil, err
	}
	if nextLayout != strings.TrimSpace(settings.SeatMapLayout) {
		if _, err := s.eventSettingsRepo.UpdateActiveSeatMap(ctx, settings.SeatMapColumns, settings.SeatMapColorMode, nextLayout); err != nil {
			return nil, err
		}
	}

	return students, nil
}

func normalizedSortValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func sortStudentsByMajorAndName(students []models.Student, pairOrder []string) {
	studentByID := make(map[string]models.Student, len(students))
	for _, student := range students {
		studentByID[student.ID.String()] = student
	}

	majorRank := make(map[string]int)
	for _, pairID := range pairOrder {
		student, ok := studentByID[pairID]
		if !ok {
			continue
		}
		major := normalizedSortValue(student.Major)
		if _, exists := majorRank[major]; !exists {
			majorRank[major] = len(majorRank)
		}
	}

	missingMajors := make([]string, 0)
	for _, student := range students {
		major := normalizedSortValue(student.Major)
		if _, exists := majorRank[major]; !exists {
			majorRank[major] = -1
			missingMajors = append(missingMajors, major)
		}
	}
	sort.Strings(missingMajors)
	firstMissingRank := len(majorRank) - len(missingMajors)
	for index, major := range missingMajors {
		majorRank[major] = firstMissingRank + index
	}

	sort.SliceStable(students, func(i, j int) bool {
		a := students[i]
		b := students[j]
		aMajor := normalizedSortValue(a.Major)
		bMajor := normalizedSortValue(b.Major)
		if majorRank[aMajor] != majorRank[bMajor] {
			return majorRank[aMajor] < majorRank[bMajor]
		}
		aClass := normalizedSortValue(a.ClassName)
		bClass := normalizedSortValue(b.ClassName)
		if aClass != bClass {
			return aClass < bClass
		}
		aName := normalizedSortValue(a.Name)
		bName := normalizedSortValue(b.Name)
		if aName != bName {
			return aName < bName
		}
		return a.ID.String() < b.ID.String()
	})
}

func buildAlphabeticalSeatLayout(layout string, columns int, students []models.Student) (string, error) {
	var currentRows [][]string
	if err := json.Unmarshal([]byte(layout), &currentRows); err != nil {
		return "", err
	}
	if columns <= 0 {
		columns = 20
	}

	trailingEmptyRows := 0
	for index := len(currentRows) - 1; index >= 0; index-- {
		hasSeat := false
		for _, value := range currentRows[index] {
			if strings.TrimSpace(value) != "" {
				hasSeat = true
				break
			}
		}
		if hasSeat {
			break
		}
		trailingEmptyRows++
	}

	keys := make([]string, 0, len(students)*seatsPerStudent)
	for _, student := range students {
		id := student.ID.String()
		keys = append(keys, id+"-student", id+"-companion")
	}
	baseRows := (len(keys) + columns - 1) / columns
	rows := make([][]string, baseRows+trailingEmptyRows)
	for index := range rows {
		rows[index] = make([]string, columns)
	}

	leftColumns := columns / 2
	rightColumns := columns - leftColumns
	leftCapacity := baseRows * leftColumns
	leftKeysEnd := leftCapacity
	if len(keys) < leftKeysEnd {
		leftKeysEnd = len(keys)
	}
	leftKeys := keys[:leftKeysEnd]
	rightKeys := keys[leftKeysEnd:]
	for row := 0; row < baseRows; row++ {
		for column := 0; column < leftColumns; column++ {
			index := row*leftColumns + column
			if index < len(leftKeys) {
				rows[row][column] = leftKeys[index]
			}
		}
		for column := 0; column < rightColumns; column++ {
			index := row*rightColumns + column
			if index < len(rightKeys) {
				rows[row][leftColumns+column] = rightKeys[index]
			}
		}
	}

	encoded, err := json.Marshal(rows)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func parseSeatLayoutPairOrder(layout string, columns int) []string {
	var rows [][]string
	if err := json.Unmarshal([]byte(layout), &rows); err != nil {
		return nil
	}
	if columns <= 0 {
		columns = 20
	}
	leftColumns := columns / 2

	order := make([]string, 0)
	seen := make(map[string]struct{})
	appendPair := func(value string) {
		key := strings.TrimSpace(value)
		if key == "" {
			return
		}
		pairID := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(key, "-student"), "-companion"))
		if pairID == "" {
			return
		}
		if _, exists := seen[pairID]; exists {
			return
		}
		seen[pairID] = struct{}{}
		order = append(order, pairID)
	}

	// Urutan visual dibaca per baris pada sisi kiri, lalu sisi kanan setelah lorong.
	for _, row := range rows {
		limit := leftColumns
		if len(row) < limit {
			limit = len(row)
		}
		for column := 0; column < limit; column++ {
			appendPair(row[column])
		}
	}
	for _, row := range rows {
		limit := columns
		if len(row) < limit {
			limit = len(row)
		}
		for column := leftColumns; column < limit; column++ {
			appendPair(row[column])
		}
	}
	return order
}

func formatSeatRange(start int) string {
	return fmt.Sprintf("%03d - %03d", start, start+seatsPerStudent-1)
}
