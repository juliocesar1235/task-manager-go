package models

import "time"

// const (
// 	InProgress TaskStatus = iota
// 	Completed
// 	Failed
// )

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
