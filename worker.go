package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type TaskJob struct {
	Urls []string
}

type Dispatcher struct {
	chStop     chan struct{}
	chTasks    chan TaskJob
	maxWorkers int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewDispatcher(ctx context.Context, cancel context.CancelFunc, countWorkers int) *Dispatcher {

	return &Dispatcher{
		chStop:     make(chan struct{}),
		chTasks:    make(chan TaskJob, 100),
		maxWorkers: countWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (dis *Dispatcher) Start() {

	for i := 0; i < dis.maxWorkers; i++ {
		go dis.worker(i)
	}

	go func() {
		<-dis.ctx.Done()
		log.Println("Dispatcher shutting down...")
		close(dis.chTasks) // закрываем канал, чтобы воркеры завершились
	}()
}

func (dis *Dispatcher) worker(numberWorker int) {

	for task := range dis.chTasks {
		fmt.Println("Worker ", numberWorker, "работает над задачей")
		dis.executeTask(task)
	}
}

func (dis *Dispatcher) executeTask(taskJob TaskJob) {

	task := Task{
		ID:     uuid.New().String(),
		Urls:   taskJob.Urls,
		Status: StatusPending,
	}

	if err := st.SaveTask(&task); err != nil {
		log.Println("Error saving task:", err)
		return
	}
	log.Printf("Task %s created with URLs: %v\n", task.ID, task.Urls)

	dir := filepath.Join("storage", "downloads", task.ID)
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		log.Println("Error creating directory:", err)
		return
	}

	// Устанавливаем статус задачи в "in_process"
	if err := st.ChangeStatusTask(&task, StatusInProcess); err != nil {
		log.Println("Error updating task status to in_process:", err)
		return
	}

	resultDonwload := true
	for _, url := range task.Urls {
		if err := st.DownloadFile(url, dir); err != nil {
			resultDonwload = false
			break
		}
	}

	var status StatusTask
	if resultDonwload {
		status = StatusCompleted
	} else {
		status = StatusFailed
	}

	// Здесь можно обновить статус задачи на "completed" или "failed" в зависимости от результата
	if err := st.ChangeStatusTask(&task, status); err != nil {
		log.Printf("Error updating task status to completed: %v\n", err)
		setStatusFailed(&task)
	} else {
		log.Printf("Task %s completed successfully.\n", task.ID)
	}

}

func (dis *Dispatcher) stop() {
	dis.cancel()
}
