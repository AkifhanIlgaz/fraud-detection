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
