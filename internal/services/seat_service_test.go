package services

import (
	"reflect"
	"testing"

	"graduation-invitation/internal/models"

	"github.com/google/uuid"
)

func TestParseSeatLayoutPairOrderReadsLeftSectionBeforeRight(t *testing.T) {
	layout := `[
		["a-student", "a-companion", "b-student", "b-companion", "e-student", "e-companion", "f-student", "f-companion"],
		["c-student", "c-companion", "d-student", "d-companion", "g-student", "g-companion", "h-student", "h-companion"]
	]`

	got := parseSeatLayoutPairOrder(layout, 8)
	want := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected pair order: got %v, want %v", got, want)
	}
}

func TestSortStudentsByMajorAndNameUsesAlphabeticalNames(t *testing.T) {
	students := []models.Student{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "TRI EKA NURCAHYANI", ClassName: "12 AKL 1", Major: "AKL"},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "LIVIA AZKA SAHIRA", ClassName: "12 AKL 1", Major: "AKL"},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Name: "ANINDA PUTRI", ClassName: "12 AKL 1", Major: "AKL"},
	}
	pairOrder := []string{students[0].ID.String(), students[1].ID.String(), students[2].ID.String()}

	sortStudentsByMajorAndName(students, pairOrder)

	got := []string{students[0].Name, students[1].Name, students[2].Name}
	want := []string{"ANINDA PUTRI", "LIVIA AZKA SAHIRA", "TRI EKA NURCAHYANI"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected student order: got %v, want %v", got, want)
	}
}
