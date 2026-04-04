package fraud

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"

	"fraud-detection/internal/cache"
	"fraud-detection/internal/models"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/store"
)

type Analyzer struct {
	repo  store.TransactionRepository
	cache *cache.FraudCache
	q     *queue.Client
}

func NewAnalyzer(repo store.TransactionRepository, fc *cache.FraudCache, q *queue.Client) *Analyzer {
	return &Analyzer{repo: repo, cache: fc, q: q}
}

func (a *Analyzer) Analyze(ctx context.Context, msg queue.TransactionMessage) error {
	msg.FraudReasons = a.runRules(ctx, msg)

	id, err := bson.ObjectIDFromHex(msg.ID)
	if err != nil {
		return fmt.Errorf("invalid transaction id: %w", err)
	}

	if len(msg.FraudReasons) < 2 {
		msg.Status = models.StatusApproved
		if err := a.repo.UpdateStatus(ctx, id, msg.Status); err != nil {
			return fmt.Errorf("update transaction status: %w", err)
		}

		return nil
	}

	msg.Status = models.StatusFraud
	if err := a.repo.UpdateStatus(ctx, id, msg.Status, msg.FraudReasons...); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}

	a.publishEvent(ctx, msg)

	return nil
}

// publishEvent, analiz sonucunu events exchange'e gönderir.
// Hata durumunda ana akış kesilmez — sadece loglanır.
func (a *Analyzer) publishEvent(ctx context.Context, msg queue.TransactionMessage) {
	event := queue.TransactionEvent{
		TransactionID: msg.ID,
		UserID:        msg.UserID,
		Status:        msg.Status,
		Amount:        msg.Amount,
		FraudReasons:  msg.FraudReasons,
		CreatedAt:     msg.CreatedAt,
	}

	if err := a.q.PublishEvent(ctx, event); err != nil {
		log.Printf("[fraud] event publish hatası: %v", err)
	}
}
