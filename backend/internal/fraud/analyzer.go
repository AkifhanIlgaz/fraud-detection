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
}

func NewAnalyzer(repo store.TransactionRepository, fc *cache.FraudCache) *Analyzer {
	return &Analyzer{repo: repo, cache: fc}
}

func (a *Analyzer) Analyze(ctx context.Context, msg queue.TransactionMessage) error {
	reasons := a.runRules(ctx, msg)
	if len(reasons) < 2 {
		a.cache.RecordApprovedAmount(ctx, msg.UserID, msg.ID, msg.Amount, amountHistoryTTL)
		return nil
	}

	for _, reason := range reasons {
		if err := a.handleReason(ctx, msg, reason); err != nil {
			return err
		}
	}

	// TODO: alerts exchange'e publish et → WebSocket ile frontend'e bildir
	return nil
}

func (a *Analyzer) handleReason(ctx context.Context, msg queue.TransactionMessage, reason FraudReason) error {
	switch reason {
	case ReasonVelocity:
		if err := a.repo.MarkLastTransactionsAsFraud(ctx, msg.UserID, maxTxPerMinute); err != nil {
			return fmt.Errorf("mark last transactions as fraud: %w", err)
		}
		log.Printf("[fraud] velocity ihlali — kullanıcı %s'in son %d işlemi fraud olarak işaretlendi",
			msg.UserID, maxTxPerMinute)

	default:
		id, err := bson.ObjectIDFromHex(msg.ID)
		if err != nil {
			return fmt.Errorf("invalid transaction id: %w", err)
		}
		if err := a.repo.UpdateStatus(ctx, id, models.StatusFraud); err != nil {
			return fmt.Errorf("update fraud status: %w", err)
		}
		log.Printf("[fraud] transaction %s fraud olarak işaretlendi — sebep: %s", msg.ID, reason)
	}

	return nil
}
