package model

type FileStatus string

const (
    Pending    FileStatus = "pending"
    InProgress FileStatus = "in_progress"
    Done       FileStatus = "done"
    Failed     FileStatus = "failed"
)

type File struct {
    URL    string     `json:"url"`
    Name   string     `json:"name"`
    Status FileStatus `json:"status"`
}

type Task struct {
    ID    string `json:"id"`
    Files []File `json:"files"`
}
