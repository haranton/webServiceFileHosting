package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fileHostingCopy/internal/config"
	"fileHostingCopy/internal/model"
	"fileHostingCopy/internal/queue"
	"fileHostingCopy/internal/storage"
)

type Downloader struct {
	cfg *config.Config
	st  *storage.Storage
	q   *queue.Queue
}

func NewDownloader(cfg *config.Config, st *storage.Storage, q *queue.Queue) *Downloader {
	os.MkdirAll(cfg.Downloader.DownloadsDir, 0755)
	return &Downloader{cfg: cfg, st: st, q: q}
}

func (d *Downloader) Start() {
	tasks, _ := d.st.LoadAll()
	for _, t := range tasks {
		for i := range t.Files {
			if t.Files[i].Status == model.Pending {
				d.q.Push(TaskItem{Task: t, FileIndex: i})
			}
		}
	}

	for i := 0; i < d.cfg.Downloader.Workers; i++ {
		go d.worker()
	}
}

type TaskItem struct {
	Task      *model.Task
	FileIndex int
}

func (d *Downloader) worker() {
	for item := range d.q.Pop() {
		if ti, ok := item.(TaskItem); ok {
			d.downloadFile(ti.Task, ti.FileIndex)
		}
	}
}

func (d *Downloader) downloadFile(task *model.Task, idx int) {
	file := &task.Files[idx]
	file.Status = model.InProgress
	d.st.Save(task)

	log.Printf("[DOWNLOADER] Start downloading: %s (task %s)\n", file.URL, task.ID)

	resp, err := http.Get(file.URL)
	if err != nil {
		log.Printf("[DOWNLOADER] FAILED request %s: %v\n", file.URL, err)
		file.Status = model.Failed
		d.st.Save(task)
		return
	}
	defer resp.Body.Close()

	taskDir := filepath.Join(d.cfg.Downloader.DownloadsDir, task.ID)
	os.MkdirAll(taskDir, 0755)

	path := d.uniqueFilename(taskDir, file.Name)
	out, err := os.Create(path)
	if err != nil {
		log.Printf("[DOWNLOADER] FAILED create file %s: %v\n", path, err)
		file.Status = model.Failed
		d.st.Save(task)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("[DOWNLOADER] FAILED save file %s: %v\n", path, err)
		file.Status = model.Failed
	} else {
		file.Status = model.Done
		file.Name = filepath.Base(path)
		log.Printf("[DOWNLOADER] Done: %s → %s (task %s)\n", file.URL, path, task.ID)
	}
	d.st.Save(task)
}

func (d *Downloader) uniqueFilename(dir, base string) string {
	path := filepath.Join(dir, base)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s_%d%s", name, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
