package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var st = NewStorage("storage")
var dis = NewDispatcher(5)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dis.ctx = ctx
	dis.Start()

	r := gin.Default()
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "ping")
	})
	r.POST("/tasks", LoadTasks)
	r.GET("/tasks/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		task, err := st.GetTaskByID(id)
		if err != nil {
			ctx.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		ctx.JSON(200, task)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Ловим сигналы ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Сигнал на завершение, выключаем сервер и диспетчер...")

	// Останавливаем диспетчер
	dis.stop()

	// Контекст с таймаутом для корректного завершения сервера
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Ошибка остановки сервера: %v\n", err)
	}

	log.Println("Сервер и диспетчер завершили работу")

}
func LoadTasks(ctx *gin.Context) {

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

	// to do отправляем задачу в очередь

	dis.chTasks <- TaskJob{
		Urls: urls,
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
