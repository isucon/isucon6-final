package seed
// ベンチマーカーから参照される
import "encoding/json"

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

func GetStroke(name string) []Stroke {
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
