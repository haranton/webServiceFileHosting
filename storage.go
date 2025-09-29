package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Storage struct {
	BasePath        string
	TaskDirPath     string
	DownloadDirPath string
}

func NewStorage(basePath string) *Storage {
	//todo get path from config
	return &Storage{
		BasePath:        basePath,
		TaskDirPath:     basePath + "/tasks",
		DownloadDirPath: basePath + "/downloads",
	}
}

func (s *Storage) SaveTask(task *Task) error {
	taskJson, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	resultPath := filepath.Join(s.TaskDirPath, task.ID+".json")

	if err := os.WriteFile(resultPath, taskJson, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (s *Storage) ChangeStatusTask(task *Task, status StatusTask) error {
	task.Status = status
	return s.SaveTask(task)
}

func (s *Storage) GetTaskByID(id string) (*Task, error) {

	resultPath := filepath.Join(s.TaskDirPath, id+".json")
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil

}

func (st *Storage) DownloadFile(url string, dirTask string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, fileName := filepath.Split(url)
	if fileName == "" {
		fileName = uuid.New().String() // fallback если URL заканчивается на /
	}

	fileName += "_" + uuid.New().String()
	filePath := filepath.Join(dirTask, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		resp.Body.Close()
		return err
	}

	// Скопировать содержимое
	_, err = io.Copy(out, resp.Body)
	resp.Body.Close()
	out.Close()

	if err != nil {
		return err
	}

	log.Printf("Downloaded %s -> %s\n", url, filePath)
	return nil

}
