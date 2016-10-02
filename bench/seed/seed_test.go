package seed

import "testing"

func TestGetStroke(t *testing.T) {
	s := GetStroke("main001")
	if len(s) != 41 {
		t.Errorf("main001: length is not 41")
	}
}
