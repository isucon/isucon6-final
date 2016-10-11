package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type stroke struct {
	Red    int     `json:"red"`
	Green  int     `json:"green"`
	Blue   int     `json:"blue"`
	Width  int     `json:"width"`
	Points []point `json:"points"`
	Alpha  float64 `json:"alpha"`
}

func prepareSeedData(seedDir string) ([]stroke, error) {
	if seedDir == "" {
		return nil, errors.New("seedDataディレクトリが指定されていません")
	}
	info, err := os.Stat(seedDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("seedDirがディレクトリではありません")
	}

	file, err := os.Open(seedDir + "/main001.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	strokes := []stroke{}

	byte, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(byte, &strokes)
	if err != nil {
		return nil, err
	}

	return strokes, nil
}
