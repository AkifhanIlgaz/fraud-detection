package worker

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"

	"fraud-detection/internal/queue"
)

type Worker struct {
	queue *queue.Client
	redis *redis.Client
}

func New(q *queue.Client, rdb *redis.Client) *Worker {
	return &Worker{queue: q, redis: rdb}
}

// Run, transactions queue'sunu dinlemeye başlar.
// Sinyal yönetimi ve bağlantı kurma cmd/worker/main.go'da kalır.
func (w *Worker) Run(ctx context.Context) error {
	return w.queue.ConsumeTransactions(ctx, w.handleTransaction)
}

func (w *Worker) handleTransaction(msg queue.TransactionMessage) error {
	log.Printf("[worker] transaction alındı — id=%s user=%s amount=%.2f status=%s",
		msg.ID, msg.UserID, msg.Amount, msg.Status)

	// TODO: internal/fraud paketindeki fraud detection logic buraya gelecek.
	// Fraud tespit edilirse alerts exchange'ine publish edilecek.

	return nil
}
