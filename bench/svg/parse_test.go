package svg

import "testing"

func TestParse(t *testing.T) {
	s, err := Parse([]byte(`
	<?xml version="1.0" standalone="no"?>
	<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
	<svg xmlns="http://www.w3.org/2000/svg" version="1.1" baseProfile="full" width="1028" height="768" style="width:1028px;height:768px;background-color:white;" viewBox="0 0 1028 768">
	<polyline stroke="rgba(128,128,128,0.7)" stroke-width="20" stroke-linecap="round" stroke-linejoin="round" fill="none" points="105,204 105,202 106,193 114,179 129,162 152,143 175,128 206,116 232,111 250,111 272,111 290,113 304,122 324,145 332,165 343,191 351,217 354,240 355,262 354,281 343,303 322,333 300,354 271,376 252,387 237,394 223,399 216,401 213,402 212,405"></polyline>
	<polyline id="123" stroke="rgba(128,128,128,0.7)" stroke-width="20" stroke-linecap="round" stroke-linejoin="round" fill="none" points="105,204 105,202 106,193 114,179 129,162 152,143 175,128 206,116 232,111 250,111 272,111 290,113 304,122 324,145 332,165 343,191 351,217 354,240 355,262 354,281 343,303 322,333 300,354 271,376 252,387 237,394 223,399 216,401 213,402 212,405"></polyline>
	</svg>`))

	if err != nil {
		t.Errorf("%v", err)
	}

	if len(s.PolyLines) != 2 {
		t.Errorf("want %d, got %d", 2, len(s.PolyLines))
	}

	if len(s.PolyLines[0].Points) != 30 {
		t.Errorf("want %d, got %d", 30, len(s.PolyLines[0].Points))
	}

	if s.PolyLines[1].ID != "123" {
		t.Errorf("want %s, got %s", "123", s.PolyLines[1].ID)
	}

	if s.PolyLines[0].Points[0].X != 105 {
		t.Errorf("want %d, got %d", 105, s.PolyLines[0].Points[0].X)
	}

	if s.PolyLines[0].Points[0].Y != 204 {
		t.Errorf("want %d, got %d", 204, s.PolyLines[0].Points[0].Y)
	}

}
