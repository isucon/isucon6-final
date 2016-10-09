package seed

import "testing"

func TestGetStroke(t *testing.T) {
	s := GetStrokes("isu")
	if len(s) != 41 {
		t.Errorf("isu: length is not 41")
	}
}
