package model

type Activity struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
	Notes       string `json:"notes"`
}
