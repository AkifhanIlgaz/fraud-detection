package dto

type CreateTransactionRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// query params — from/to parsed as RFC3339 strings, validated in handler
type TransactionsBetweenRequest struct {
	From string `query:"from"`
	To   string `query:"to"`
}

type UserTrustScoreResponse struct {
	UserID     string  `json:"user_id"`
	Score      float64 `json:"score"`       // 0-100
	RiskLevel  string  `json:"risk_level"`  // low | medium | high
	Total      int64   `json:"total"`
	FraudCount int64   `json:"fraud_count"`
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
