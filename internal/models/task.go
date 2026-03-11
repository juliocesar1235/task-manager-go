package models

import "time"

// const (
// 	InProgress TaskStatus = iota
// 	Completed
// 	Failed
// )

type Task struct {
	ID        string     `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Status    string     `json:"status" db:"Status"`
	CreatedAt time.Time  `json:"created_at" db:"CreatedAt"`
	UpdatedAt *time.Time `json:"updated_at" db:"UpdatedAt"`
}
