package main

import (
	"context"
	"log"
	"time"

	"fraud-detection/config"
	"fraud-detection/internal/mcp"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/service"
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

	svc := service.NewTransactionService(repo, q)
	server := mcp.NewServer(svc)

	log.Println("fraud-detection MCP server started (stdio)")
	if err := server.Run(); err != nil {
		log.Fatalf("mcp server: %v", err)
	}
}
