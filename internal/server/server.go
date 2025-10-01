package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"fileHostingCopy/internal/config"
	"fileHostingCopy/internal/downloader"
	"fileHostingCopy/internal/model"
	"fileHostingCopy/internal/queue"
	"fileHostingCopy/internal/storage"
)

type Server struct {
	cfg *config.Config
	st  *storage.Storage
	q   *queue.Queue
	eng *gin.Engine
	srv *http.Server
}

func NewServer(cfg *config.Config, st *storage.Storage, q *queue.Queue) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	s := &Server{
		cfg: cfg,
		st:  st,
		q:   q,
		eng: r,
	}

	s.registerRoutes()

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	return s
}

func (s *Server) registerRoutes() {
	s.eng.POST("/tasks", s.createTask)
	s.eng.GET("/tasks/:id", s.getTask)
}

func (s *Server) Run() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

type createTaskRequest struct {
	URLs []string `json:"urls"`
}

func (s *Server) createTask(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	task := s.st.NewTask(req.URLs)
	if err := s.st.Save(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save task"})
		return
	}

	log.Printf("[SERVER] New task created: %s, files: %d\n", task.ID, len(task.Files))

	for i := range task.Files {
		if task.Files[i].Status == model.Pending {
			log.Printf("[SERVER] Queued file: %s (task %s)\n", task.Files[i].Name, task.ID)
			s.q.Push(downloader.TaskItem{Task: task, FileIndex: i})
		}
	}

	c.JSON(http.StatusOK, task)
}

func (s *Server) getTask(c *gin.Context) {
	id := c.Param("id")
	path := s.st.Dir() + "/" + id + ".json"
	data, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}
