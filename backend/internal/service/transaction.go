package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"fraud-detection/internal/api/dto"
	"fraud-detection/internal/models"
	"fraud-detection/internal/store"
)

type TransactionService struct {
	repo store.TransactionRepository
}

func NewTransactionService(repo store.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Create(ctx context.Context, req dto.CreateTransactionRequest) (dto.TransactionResponse, error) {
	tx := models.Transaction{
		UserID:    req.UserID,
		Amount:    req.Amount,
		Lat:       req.Lat,
		Lon:       req.Lon,
		Status:    models.StatusPending,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Insert(ctx, &tx); err != nil {
		return dto.TransactionResponse{}, fmt.Errorf("create transaction: %w", err)
	}
	return toResponse(tx), nil
}

func (s *TransactionService) GetByUserID(ctx context.Context, userID string) ([]dto.TransactionResponse, error) {
	txs, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get by user_id: %w", err)
	}
	return toResponses(txs), nil
}

func (s *TransactionService) GetFraudsBetween(ctx context.Context, from, to time.Time) ([]dto.TransactionResponse, error) {
	txs, err := s.repo.FindFraudsBetween(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get frauds between: %w", err)
	}
	return toResponses(txs), nil
}

func (s *TransactionService) UpdateStatus(ctx context.Context, rawID string, req dto.UpdateStatusRequest) error {
	status := models.TransactionStatus(req.Status)
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	id, err := bson.ObjectIDFromHex(rawID)
	if err != nil {
		return fmt.Errorf("invalid transaction id: %w", err)
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func toResponse(tx models.Transaction) dto.TransactionResponse {
	return dto.TransactionResponse{
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

func toResponses(txs []models.Transaction) []dto.TransactionResponse {
	out := make([]dto.TransactionResponse, len(txs))
	for i, tx := range txs {
		out[i] = toResponse(tx)
	}
	return out
}
