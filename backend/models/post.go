package models

import "time"

type Post struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}
