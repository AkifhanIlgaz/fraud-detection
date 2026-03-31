package worker

import (
	"context"

	"fraud-detection/internal/fraud"
	"fraud-detection/internal/queue"
)

type Worker struct {
	queue    *queue.Client
	analyzer *fraud.Analyzer
}

func New(q *queue.Client, a *fraud.Analyzer) *Worker {
	return &Worker{queue: q, analyzer: a}
}

// Run, transactions queue'sunu dinlemeye başlar.
// Sinyal yönetimi ve bağlantı kurma cmd/worker/main.go'da kalır.
func (w *Worker) Run(ctx context.Context) error {
	return w.queue.ConsumeTransactions(ctx, func(msg queue.TransactionMessage) error {
		return w.analyzer.Analyze(ctx, msg)
	})
}
