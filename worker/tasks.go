package worker

import "github.com/tedyst/licenta/db/queries"

type TaskType string

const (
	TaskTypePostgresScan TaskType = "postgres_scan"
)

type PostgresScan struct {
	Scan     queries.PostgresScan     `json:"scan"`
	Database queries.PostgresDatabase `json:"database"`
}

type Task struct {
	Type         TaskType     `json:"type"`
	PostgresScan PostgresScan `json:"postgres_scan"`
}
