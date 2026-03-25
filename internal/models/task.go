package models

import "time"

type TaskStatus string

const (
	StatusInProgress TaskStatus = "IN_PROGRESS"
	StatusPending    TaskStatus = "PENDING"
	StatusCompleted  TaskStatus = "COMPLETED"
	StatusError      TaskStatus = "ERROR"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusInProgress, StatusPending, StatusCompleted, StatusError:
		return true
	}
	return false
}

type Task struct {
	ID        string     `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Status    TaskStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateTaskInput struct {
	Title  *string
	Status *TaskStatus
}
