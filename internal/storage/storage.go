package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"fileHostingCopy/internal/model"

	"github.com/google/uuid"
)

type Storage struct {
	dir string
	mu  sync.RWMutex
}

func NewStorage(dir string) *Storage {
	os.MkdirAll(dir, 0755)
	return &Storage{dir: dir}
}

func (s *Storage) Save(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(s.dir, task.ID+".json"), data, 0644)
}

func (s *Storage) LoadAll() ([]*model.Task, error) {
	var tasks []*model.Task
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return tasks, nil
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path := filepath.Join(s.dir, f.Name())
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("cannot read task file %s: %v", f.Name(), err)
			continue
		}
		var t model.Task
		if err := json.Unmarshal(data, &t); err != nil {
			log.Printf("cannot parse task file %s: %v", f.Name(), err)
			continue
		}
		for i := range t.Files {
			if t.Files[i].Status == model.InProgress {
				t.Files[i].Status = model.Pending
			}
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (s *Storage) NewTask(urls []string) *model.Task {
	id := uuid.New().String()
	task := &model.Task{ID: id}
	for _, u := range urls {
		task.Files = append(task.Files, model.File{
			URL:    u,
			Name:   filepath.Base(u),
			Status: model.Pending,
		})
	}
	return task
}

func (s *Storage) Dir() string {
	return s.dir
}
