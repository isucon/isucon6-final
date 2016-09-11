package svg

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
)

type Point struct {
	X int
	Y int
}

type PolyLine struct {
	Stroke      string `xml:"stroke,attr"`
	StrokeWidth int    `xml:"stroke-width,attr"`
	PointsRaw   string `xml:"points,attr"`
	Points      []Point
}

type Svg struct {
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	Style   string `xml:"style,attr"`
	ViewBox string `xml:"viewBox,attr"`

	PolyLine PolyLine `xml:"polyline"`
}

func Parse(data []byte) Svg {
	v := Svg{}
	err := xml.Unmarshal([]byte(data), &v)

	if err != nil {
		fmt.Printf("error: %v", err)
		return Svg{}
	}

	points := make([]Point, 0, 100)

	for _, s := range strings.Split(v.PolyLine.PointsRaw, " ") {
		ps := strings.Split(s, ",")
		x, _ := strconv.Atoi(ps[0])
		y, _ := strconv.Atoi(ps[1])
		points = append(points, Point{x, y})
	}

	v.PolyLine.Points = points

	pp.Println(v)

	return v
}
