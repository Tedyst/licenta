package worker

import "github.com/tedyst/licenta/models"

type TaskType string

const (
	TaskTypePostgresScan TaskType = "postgres_scan"
)

type PostgresScan struct {
	Scan     models.PostgresScan      `json:"scan"`
	Database models.PostgresDatabases `json:"database"`
}

type Task struct {
	Type         TaskType     `json:"type"`
	PostgresScan PostgresScan `json:"postgres_scan"`
}
