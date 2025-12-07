package entity

import "time"

const (
	TaskStatusProcess Status = "PROCESS"
	TaskStatusDone    Status = "DONE"
)

type (
	IdTask int
)
type IdFile int
type Url string
type Status string

type Task struct {
	Id      IdTask
	Timeout time.Duration
	Status  Status
	Files   []File
}

type File struct {
	Id    IdFile
	Data  []byte
	Url   Url
	Error error
}

// NewTask
func NewTask(timeout time.Duration, urls []Url) Task {
	files := make([]File, len(urls))
	for i, url := range urls {
		files[i] = File{Url: url}
	}
	return Task{
		Timeout: timeout,
		Status:  TaskStatusProcess,
		Files:   files,
	}
}
