package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var st = NewStorage("storage")

func main() {

	chUrls := make(chan string 
	)

	for range 4 {
		go worker()
	}

	r := gin.Default()
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "ping")
	})
	r.POST("/tasks", LoaadTasks)
	r.GET("/tasks/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		task, err := st.GetTaskByID(id)
		if err != nil {
			ctx.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		ctx.JSON(200, task)
	})
	r.Run(":8080")
}

func LoaadTasks(ctx *gin.Context) {

	var request struct {
		Urls []string `json:"urls"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	createTasks(request.Urls)

	ctx.JSON(200, gin.H{"urls": request.Urls})
}

func createTasks(urls []string) {

	task := Task{
		ID:     uuid.New().String(),
		Urls:   urls,
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



	for _, url := range urls {
		go st.DownloadFile(url, dir)
	}

	// Здесь можно обновить статус задачи на "completed" или "failed" в зависимости от результата
	if err := st.ChangeStatusTask(&task, StatusCompleted); err != nil {
		log.Printf("Error updating task status to completed: %v\n", err)
		setStatusFailed(&task)
	} else {
		log.Printf("Task %s completed successfully.\n", task.ID)
	}
}

func setStatusFailed(task *Task) error {
	task.Status = StatusFailed
	taskJson, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		log.Println("Error marshalling task to JSON:", err)
		return err
	}
	dir := filepath.Join("storage", "tasks")
	resultPath := filepath.Join(dir, task.ID+".json")
	if err := os.WriteFile(resultPath, taskJson, os.ModePerm); err != nil {
		log.Println("Error writing task to file:", err)
		return err
	}

	return nil
}

func worker(chUrls chan string, dirTask string) {
	for url := range chUrls {
		st.DownloadFile(url, dirTask)
	}
}
