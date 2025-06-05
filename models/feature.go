package models

type Feature struct {
	Type       string `json:"type"`
	Properties struct {
		Timestamp string `json:"timestamp"`
	} `json:"properties"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
}
