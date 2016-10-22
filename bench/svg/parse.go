package svg

import (
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
)

type Point struct {
	X float32
	Y float32
}

type PolyLine struct {
	ID          string `xml:"id,attr"`
	Stroke      string `xml:"stroke,attr"`
	StrokeWidth int    `xml:"stroke-width,attr"`
	PointsRaw   string `xml:"points,attr"`
	Points      []Point
}

type SVG struct {
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	Style   string `xml:"style,attr"`
	ViewBox string `xml:"viewBox,attr"`

	PolyLines []PolyLine `xml:"polyline"`
}

func Parse(data []byte) (*SVG, error) {
	v := &SVG{}
	err := xml.Unmarshal(data, &v)

	if err != nil {
		return nil, err
	}

	for i, polyLine := range v.PolyLines {
		points := make([]Point, 0)

		for _, s := range strings.Split(polyLine.PointsRaw, " ") {
			ps := strings.Split(s, ",")
			if len(ps) < 2 {
				return nil, errors.New("polylineの形式が不正です")
			}

			x, err := strconv.ParseFloat(ps[0], 32)
			if err != nil {
				return nil, err
			}
			y, err := strconv.ParseFloat(ps[1], 32)
			if err != nil {
				return nil, err
			}
			points = append(points, Point{float32(x), float32(y)})
		}

		v.PolyLines[i].Points = points
	}

	return v, nil
}
