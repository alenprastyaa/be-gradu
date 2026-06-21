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

	students = s.orderStudentsByActiveSeatLayout(ctx, students)

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

func (s *SeatService) orderStudentsByActiveSeatLayout(ctx context.Context, students []models.Student) []models.Student {
	if len(students) == 0 || s.eventSettingsRepo == nil {
		return students
	}

	settings, err := s.eventSettingsRepo.GetActive(ctx)
	if err != nil || settings == nil || strings.TrimSpace(settings.SeatMapLayout) == "" {
		return students
	}

	pairOrder := parseSeatLayoutPairOrder(settings.SeatMapLayout, settings.SeatMapColumns)
	if len(pairOrder) == 0 {
		return students
	}

	studentByID := make(map[string]models.Student, len(students))
	for _, student := range students {
		studentByID[student.ID.String()] = student
	}

	ordered := make([]models.Student, 0, len(students))
	used := make(map[string]struct{}, len(students))
	for _, pairID := range pairOrder {
		student, ok := studentByID[pairID]
		if !ok {
			continue
		}
		if _, exists := used[pairID]; exists {
			continue
		}
		ordered = append(ordered, student)
		used[pairID] = struct{}{}
	}

	for _, student := range students {
		id := student.ID.String()
		if _, exists := used[id]; exists {
			continue
		}
		ordered = append(ordered, student)
	}

	return ordered
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
