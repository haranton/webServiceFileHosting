package main

type StatusTask string

const (
	StatusPending   StatusTask = "pending"
	StatusInProcess StatusTask = "in_process"
	StatusCompleted StatusTask = "completed"
	StatusFailed    StatusTask = "failed"
)

type Task struct {
	ID     string
	Urls   []string
	Status StatusTask
}
