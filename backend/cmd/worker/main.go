package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fraud-detection/config"
	"fraud-detection/internal/cache"
	"fraud-detection/internal/fraud"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/store"
	"fraud-detection/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoClient, err := store.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	repo, err := store.NewTransactionRepository(ctx, mongoClient)
	if err != nil {
		log.Fatalf("repo: %v", err)
	}

	rdb, err := cache.Connect(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	q, err := queue.New(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer q.Close()

	analyzer := fraud.NewAnalyzer(repo, cache.NewFraudCache(rdb), q)
	w := worker.New(q, analyzer)

	if err := w.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}

	log.Println("worker başladı, mesajlar bekleniyor...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("worker durduruluyor...")
}
