package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"fraud-detection/internal/models"
)

type TransactionRepository interface {
	Insert(ctx context.Context, tx *models.Transaction) error
	FindByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Transaction, int64, error)
	FindFraudsBetween(ctx context.Context, from, to time.Time, skip, limit int64) ([]models.Transaction, int64, error)
	MarkLastTransactionsAsFraud(ctx context.Context, userID string, count int) error
	UpdateStatus(ctx context.Context, id bson.ObjectID, status models.TransactionStatus) error
	GetUserStats(ctx context.Context, userID string) (models.UserTransactionStats, error)
}

type transactionRepo struct {
	col *mongo.Collection
}

func NewTransactionRepository(ctx context.Context, client *mongo.Client) (TransactionRepository, error) {
	col := client.Database(dbName).Collection(colTransactions)

	// Compound index: user_id (asc) prefix olarak seçildi çünkü en seçici alan —
	// bir kullanıcının işlemleri toplam verinin küçük bir kesiti.
	// created_at (desc) ikinci alan olarak eklendi; FindByUserID sorgusunda
	// MongoDB hem filtreyi hem de sıralamayı index üzerinden karşılar, ekstra sort adımı gerekmez.
	// FindTransactionsBetween'deki created_at range sorgusu da bu index'ten yararlanır.
	// Index key sırası önemli olduğundan bson.D kullanılmalı;
	// bson.M map olduğu için key sırasını garanti etmez, MongoDB reddeder.
	indexKeys := bson.D{
		{Key: "user_id", Value: 1},
		{Key: "created_at", Value: -1},
	}

	indexOpts := options.Index().SetName("user_id_created_at")
	index := mongo.IndexModel{
		Keys:    indexKeys,
		Options: indexOpts,
	}
	if _, err := col.Indexes().CreateOne(ctx, index); err != nil {
		return nil, fmt.Errorf("create index: %w", err)
	}

	return &transactionRepo{col: col}, nil
}

func (r *transactionRepo) Insert(ctx context.Context, tx *models.Transaction) error {
	res, err := r.col.InsertOne(ctx, tx)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	tx.ID = res.InsertedID.(bson.ObjectID)

	return nil
}

func (r *transactionRepo) FindByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Transaction, int64, error) {
	filter := bson.M{"user_id": userID}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count by user_id: %w", err)
	}

	findOpts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(skip).
		SetLimit(limit)

	cur, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("find by user_id: %w", err)
	}
	defer cur.Close(ctx)

	var txs []models.Transaction
	if err := cur.All(ctx, &txs); err != nil {
		return nil, 0, fmt.Errorf("decode transactions: %w", err)
	}

	return txs, total, nil
}

func (r *transactionRepo) FindFraudsBetween(ctx context.Context, from, to time.Time, skip, limit int64) ([]models.Transaction, int64, error) {
	filter := bson.M{
		"status": models.StatusFraud,
		"created_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count frauds between: %w", err)
	}

	findOpts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(skip).
		SetLimit(limit)

	cur, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("find frauds between: %w", err)
	}
	defer cur.Close(ctx)

	var txs []models.Transaction
	if err := cur.All(ctx, &txs); err != nil {
		return nil, 0, fmt.Errorf("decode frauds: %w", err)
	}

	return txs, total, nil
}

func (r *transactionRepo) GetUserStats(ctx context.Context, userID string) (models.UserTransactionStats, error) {
	pipeline := mongo.Pipeline{
		// Sadece bu kullanıcının işlemlerini al
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		// Toplam işlem sayısını, fraud ve suspicious olanları say
		{{Key: "$group", Value: bson.M{
			"_id":   "$user_id",
			"total": bson.M{"$sum": 1},
			"fraud_count": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$status", models.StatusFraud}},
						1,
						0,
					},
				},
			},
			"suspicious_count": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$status", models.StatusSuspicious}},
						1,
						0,
					},
				},
			},
		}}},
	}

	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return models.UserTransactionStats{}, fmt.Errorf("aggregate user stats: %w", err)
	}
	defer cur.Close(ctx)

	var results []models.UserTransactionStats
	if err := cur.All(ctx, &results); err != nil {
		return models.UserTransactionStats{}, fmt.Errorf("decode user stats: %w", err)
	}

	if len(results) == 0 {
		return models.UserTransactionStats{UserID: userID}, nil
	}

	return results[0], nil
}

func (r *transactionRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, status models.TransactionStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status}}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("transaction %s not found", id.Hex())
	}

	return nil
}

func (r *transactionRepo) MarkLastTransactionsAsFraud(ctx context.Context, userID string, count int) error {
	cur, err := r.col.Find(ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(count)),
	)
	if err != nil {
		return fmt.Errorf("find last transactions: %w", err)
	}
	defer cur.Close(ctx)

	var ids []bson.ObjectID
	for cur.Next(ctx) {
		var tx models.Transaction
		if err := cur.Decode(&tx); err != nil {
			return fmt.Errorf("decode transaction: %w", err)
		}
		ids = append(ids, tx.ID)
	}

	if len(ids) == 0 {
		return nil
	}

	// Bulunan tüm ID'leri tek sorguda fraud olarak işaretle.
	_, err = r.col.UpdateMany(ctx,
		bson.M{"_id": bson.M{"$in": ids}},
		bson.M{"$set": bson.M{"status": models.StatusFraud}},
	)
	if err != nil {
		return fmt.Errorf("mark as fraud: %w", err)
	}

	return nil
}
