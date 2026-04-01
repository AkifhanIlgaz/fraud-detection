package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"fraud-detection/config"
	"fraud-detection/internal/models"
	"fraud-detection/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := store.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(ctx)

	repo, err := store.NewTransactionRepository(ctx, client)
	if err != nil {
		log.Fatalf("repo init: %v", err)
	}

	now := time.Now().UTC()

	transactions := []models.Transaction{
		// --- user-001 (8 işlem) ---
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    120.50,
			Lat:       41.0082,
			Lon:       28.9784,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -2, -14),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    4875.00,
			Lat:       41.0136,
			Lon:       28.9550,
			Status:    models.StatusFraud,
			CreatedAt: now.AddDate(0, -2, -10),
			FraudReasons: []string{
				"amount_exceeds_daily_limit",
				"unusual_location",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    38.99,
			Lat:       41.0195,
			Lon:       29.0047,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -1, -22),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    980.00,
			Lat:       40.9874,
			Lon:       28.8456,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -1, -15),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    3200.75,
			Lat:       41.0451,
			Lon:       29.0123,
			Status:    models.StatusFraud,
			CreatedAt: now.AddDate(0, -1, -8),
			FraudReasons: []string{
				"multiple_transactions_short_interval",
				"high_risk_merchant",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    55.20,
			Lat:       41.0072,
			Lon:       28.9741,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, 0, -18),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    215.00,
			Lat:       41.0310,
			Lon:       28.9801,
			Status:    models.StatusPending,
			CreatedAt: now.AddDate(0, 0, -5),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    1450.00,
			Lat:       41.0890,
			Lon:       29.0512,
			Status:    models.StatusSuspicious,
			CreatedAt: now.AddDate(0, 0, -3),
			FraudReasons: []string{
				"unusual_location",
				"high_risk_merchant",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-001",
			Amount:    760.40,
			Lat:       40.9923,
			Lon:       29.1023,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, 0, -1),
		},

		// --- user-002 (7 işlem) ---
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    89.90,
			Lat:       39.9208,
			Lon:       32.8541,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -3, -5),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    14500.00,
			Lat:       39.8765,
			Lon:       32.7412,
			Status:    models.StatusFraud,
			CreatedAt: now.AddDate(0, -2, -28),
			FraudReasons: []string{
				"amount_exceeds_daily_limit",
				"card_not_present",
				"ip_country_mismatch",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    430.00,
			Lat:       39.9334,
			Lon:       32.8597,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -2, -3),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    17.50,
			Lat:       39.9104,
			Lon:       32.8022,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -1, -19),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    6750.00,
			Lat:       39.8901,
			Lon:       32.7788,
			Status:    models.StatusFraud,
			CreatedAt: now.AddDate(0, -1, -7),
			FraudReasons: []string{
				"velocity_check_failed",
				"suspicious_device_fingerprint",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    320.00,
			Lat:       39.9450,
			Lon:       32.8745,
			Status:    models.StatusPending,
			CreatedAt: now.AddDate(0, 0, -12),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    2890.00,
			Lat:       39.8542,
			Lon:       32.7103,
			Status:    models.StatusSuspicious,
			CreatedAt: now.AddDate(0, 0, -8),
			FraudReasons: []string{
				"velocity_check_failed",
				"card_not_present",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    530.75,
			Lat:       39.9012,
			Lon:       32.8301,
			Status:    models.StatusSuspicious,
			CreatedAt: now.AddDate(0, 0, -4),
			FraudReasons: []string{
				"suspicious_device_fingerprint",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-002",
			Amount:    1100.25,
			Lat:       39.9271,
			Lon:       32.8643,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, 0, -3),
		},

		// --- user-003 (5 işlem) ---
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-003",
			Amount:    250.00,
			Lat:       38.4189,
			Lon:       27.1287,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -3, -20),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-003",
			Amount:    67.30,
			Lat:       38.4341,
			Lon:       27.1402,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, -2, -11),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-003",
			Amount:    9200.00,
			Lat:       38.3982,
			Lon:       27.0991,
			Status:    models.StatusFraud,
			CreatedAt: now.AddDate(0, -1, -4),
			FraudReasons: []string{
				"foreign_currency_mismatch",
				"amount_exceeds_daily_limit",
			},
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-003",
			Amount:    510.80,
			Lat:       38.4512,
			Lon:       27.1653,
			Status:    models.StatusApproved,
			CreatedAt: now.AddDate(0, 0, -9),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-003",
			Amount:    135.00,
			Lat:       38.4078,
			Lon:       27.1101,
			Status:    models.StatusPending,
			CreatedAt: now.AddDate(0, 0, -2),
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    "user-003",
			Amount:    3750.00,
			Lat:       38.3701,
			Lon:       27.0644,
			Status:    models.StatusSuspicious,
			CreatedAt: now.AddDate(0, 0, -1),
			FraudReasons: []string{
				"foreign_currency_mismatch",
				"multiple_transactions_short_interval",
			},
		},
	}

	inserted := 0
	for i := range transactions {
		if err := repo.Insert(ctx, &transactions[i]); err != nil {
			log.Printf("  SKIP [%d] %s: %v", i+1, transactions[i].UserID, err)
			continue
		}
		fmt.Printf("  [%2d] %-10s %8.2f₺  %-10s  %s\n",
			i+1,
			transactions[i].UserID,
			transactions[i].Amount,
			transactions[i].Status,
			transactions[i].CreatedAt.Format("2006-01-02"),
		)
		inserted++
	}

	fmt.Printf("\n%d/%d transaction inserted.\n", inserted, len(transactions))
}
