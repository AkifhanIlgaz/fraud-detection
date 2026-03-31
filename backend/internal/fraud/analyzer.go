package fraud

import (
	"context"
	"fmt"
	"log"
	"slices"

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
	violations := a.runRules(ctx, msg)
	if len(violations) < 2 {
		if !slices.ContainsFunc(violations, func(v violation) bool { return v.reason == ReasonAmountAnomaly }) {
			a.cache.UpdateAmountAverage(ctx, msg.UserID, msg.Amount, amountHistoryTTL)
		}
		// TODO: fraud degilse last location guncelle
		return nil
	}

	for _, v := range violations {
		if err := a.handleViolation(ctx, msg, v); err != nil {
			return err
		}
	}

	// TODO: alerts exchange'e publish et → WebSocket ile frontend'e bildir
	return nil
}

func (a *Analyzer) handleViolation(ctx context.Context, msg queue.TransactionMessage, v violation) error {
	switch v.reason {
	case ReasonVelocity:
		if err := a.repo.MarkLastTransactionsAsFraud(ctx, msg.UserID, int(v.count)); err != nil {
			return fmt.Errorf("mark last transactions as fraud: %w", err)
		}
		log.Printf("[fraud] velocity ihlali — kullanıcı %s'in son %d işlemi fraud olarak işaretlendi",
			msg.UserID, v.count)

	default:
		id, err := bson.ObjectIDFromHex(msg.ID)
		if err != nil {
			return fmt.Errorf("invalid transaction id: %w", err)
		}
		if err := a.repo.UpdateStatus(ctx, id, models.StatusFraud); err != nil {
			return fmt.Errorf("update fraud status: %w", err)
		}
		log.Printf("[fraud] transaction %s fraud olarak işaretlendi — sebep: %s", msg.ID, v.reason)
	}

	return nil
}
