package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Transaction struct {
	ID           bson.ObjectID     `bson:"_id,omitempty"`
	UserID       string            `bson:"user_id"`
	Amount       float64           `bson:"amount"`
	Lat          float64           `bson:"lat"`
	Lon          float64           `bson:"lon"`
	Status       TransactionStatus `bson:"status"`
	CreatedAt    time.Time         `bson:"created_at"`
	FraudReasons []FraudReason     `bson:"fraud_reasons,omitempty"`
}
