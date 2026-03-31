package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fraud-detection/config"
	"fraud-detection/internal/api"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := store.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}

	repo, err := store.NewTransactionRepository(ctx, client)
	if err != nil {
		log.Fatalf("repo: %v", err)
	}

	q, err := queue.New(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer q.Close()

	app := api.NewRouter(repo, q)

	// Fiber'ı ayrı goroutine'de başlat; ana goroutine sinyal bekler.
	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received, draining connections...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("fiber shutdown: %v", err)
	}

	if err := client.Disconnect(shutdownCtx); err != nil {
		log.Printf("mongo disconnect: %v", err)
	}

	log.Println("server stopped.")
}
