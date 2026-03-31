package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fraud-detection/config"
	"fraud-detection/internal/cache"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := queue.New(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer q.Close()

	rdb, err := cache.Connect(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	w := worker.New(q, rdb)
	if err := w.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}

	log.Println("worker başladı, mesajlar bekleniyor...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("worker durduruluyor...")
}
