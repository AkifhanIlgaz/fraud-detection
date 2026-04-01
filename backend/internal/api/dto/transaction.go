package dto

import (
	"errors"
	"time"

	"fraud-detection/internal/models"
)

type CreateTransactionRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

func (r CreateTransactionRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if r.Lat < -90 || r.Lat > 90 {
		return errors.New("lat must be between -90 and 90")
	}
	if r.Lon < -180 || r.Lon > 180 {
		return errors.New("lon must be between -180 and 180")
	}
	return nil
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func (r UpdateStatusRequest) Validate() error {
	if r.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

// query params — from/to parsed as RFC3339 strings
type TransactionsBetweenRequest struct {
	From string `query:"from"`
	To   string `query:"to"`
}

func (r TransactionsBetweenRequest) Parse() (from, to time.Time, err error) {
	if r.From == "" || r.To == "" {
		return time.Time{}, time.Time{}, errors.New("from and to are required")
	}
	from, err = time.Parse(time.RFC3339, r.From)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from must be RFC3339 (e.g. 2024-01-01T00:00:00Z)")
	}
	to, err = time.Parse(time.RFC3339, r.To)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to must be RFC3339 (e.g. 2024-01-01T00:00:00Z)")
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, errors.New("from must be before to")
	}
	return from, to, nil
}

type UserTrustScoreResponse struct {
	UserID          string  `json:"user_id"`
	Score           float64 `json:"score"`            // 0-100
	RiskLevel       string  `json:"risk_level"`       // low | medium | high
	Total           int64   `json:"total"`
	FraudCount      int64   `json:"fraud_count"`
	SuspiciousCount int64   `json:"suspicious_count"`
}

type TransactionResponse struct {
	ID           string   `json:"id"`
	UserID       string   `json:"user_id"`
	Amount       float64  `json:"amount"`
	Lat          float64  `json:"lat"`
	Lon          float64  `json:"lon"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
	FraudReasons []string `json:"fraud_reasons,omitempty"`
}

func NewTransactionResponse(tx models.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:           tx.ID.Hex(),
		UserID:       tx.UserID,
		Amount:       tx.Amount,
		Lat:          tx.Lat,
		Lon:          tx.Lon,
		Status:       tx.Status.String(),
		CreatedAt:    tx.CreatedAt.Format(time.RFC3339),
		FraudReasons: tx.FraudReasons,
	}
}

func NewTransactionResponses(txs []models.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		out[i] = NewTransactionResponse(tx)
	}
	return out
}
