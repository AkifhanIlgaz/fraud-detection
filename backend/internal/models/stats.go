package models

type UserTransactionStats struct {
	UserID     string `bson:"_id"`
	Total      int64  `bson:"total"`
	FraudCount int64  `bson:"fraud_count"`
}
