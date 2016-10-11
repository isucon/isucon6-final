package seed

// ベンチマーカーから参照される
import (
	"encoding/json"
	"math/rand"
)

type Stroke struct {
	Width  int     `json:"width"`
	Red    int     `json:"red"`
	Green  int     `json:"green"`
	Blue   int     `json:"blue"`
	Alpha  float32 `json:"alpha"`
	Points []Point `json:"points"`
}

type Point struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

func GetStrokes(name string) []Stroke {
	data, err := Asset("data/" + name + ".json")
	if err != nil {
		panic(err)
	}
	var s []Stroke
	err = json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func FluctuateStroke(s Stroke) Stroke {
	points := make([]Point, 0)
	for _, p := range s.Points {
		points = append(points, Point{
			X: p.X + 3.0*rand.Float32() - 1.5,
			Y: p.Y + 3.0*rand.Float32() - 1.5,
		})
	}
	return Stroke{
		Width:  bounded(s.Width+rand.Intn(20)-10, 1, 50),
		Red:    rand.Intn(100) + 100,
		Green:  rand.Intn(100) + 100,
		Blue:   rand.Intn(100) + 100,
		Alpha:  s.Alpha,
		Points: points,
	}
}

func bounded(n, min, max int) int {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}
