package services

import (
	"context"
	"fmt"
	"sort"

	"graduation-invitation/internal/repositories"
)

const seatsPerStudent = 2

type SeatService struct {
	studentRepo *repositories.StudentRepository
}

func NewSeatService(studentRepo *repositories.StudentRepository) *SeatService {
	return &SeatService{studentRepo: studentRepo}
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

func formatSeatRange(start int) string {
	return fmt.Sprintf("%03d - %03d", start, start+seatsPerStudent-1)
}
