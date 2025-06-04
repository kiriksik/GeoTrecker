package models

import "time"

type Location struct {
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"lat"`
	Longitude float64   `json:"lon"`
	UpdatedAt time.Time `json:"updated_at"`
}
