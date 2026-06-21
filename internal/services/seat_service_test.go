package services

import (
	"reflect"
	"testing"
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
