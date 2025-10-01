package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fileHostingCopy/internal/config"
	"fileHostingCopy/internal/downloader"
	"fileHostingCopy/internal/queue"
	"fileHostingCopy/internal/server"
	"fileHostingCopy/internal/storage"
)

func main() {
	cfg := config.LoadConfig()

	st := storage.NewStorage(cfg.Downloader.TasksDir)
	q := queue.NewQueue()
	d := downloader.NewDownloader(cfg, st, q)
	d.Start()

	srv := server.NewServer(cfg, st, q)

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
