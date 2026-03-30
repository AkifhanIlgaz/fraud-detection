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
	FindByUserID(ctx context.Context, userID string) ([]models.Transaction, error)
	FindFraudsBetween(ctx context.Context, from, to time.Time) ([]models.Transaction, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status models.TransactionStatus) error
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
	indexKeys := bson.M{
		"user_id":    1,
		"created_at": -1,
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
	_, err := r.col.InsertOne(ctx, tx)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *transactionRepo) FindByUserID(ctx context.Context, userID string) ([]models.Transaction, error) {
	filter := bson.M{"user_id": userID}
	findOpts := options.Find().SetSort(bson.M{"created_at": -1})

	cur, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find by user_id: %w", err)
	}
	defer cur.Close(ctx)

	var txs []models.Transaction
	if err := cur.All(ctx, &txs); err != nil {
		return nil, fmt.Errorf("decode transactions: %w", err)
	}
	return txs, nil
}

func (r *transactionRepo) FindFraudsBetween(ctx context.Context, from, to time.Time) ([]models.Transaction, error) {
	filter := bson.M{
		"status": models.StatusFraud,
		"created_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}
	findOpts := options.Find().SetSort(bson.M{"created_at": -1})

	cur, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find frauds between: %w", err)
	}
	defer cur.Close(ctx)

	var txs []models.Transaction
	if err := cur.All(ctx, &txs); err != nil {
		return nil, fmt.Errorf("decode frauds: %w", err)
	}
	return txs, nil
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
