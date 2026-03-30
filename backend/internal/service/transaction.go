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

func (s *TransactionService) GetByUserID(ctx context.Context, userID string, page dto.PageRequest) (dto.PaginatedResponse[dto.TransactionResponse], error) {
	page.Normalize()

	txs, total, err := s.repo.FindByUserID(ctx, userID, page.Skip(), page.Limit)
	if err != nil {
		return dto.PaginatedResponse[dto.TransactionResponse]{}, fmt.Errorf("get by user_id: %w", err)
	}
	return dto.PaginatedResponse[dto.TransactionResponse]{
		Items: toResponses(txs),
		Meta:  dto.NewPageMeta(page.Page, page.Limit, total),
	}, nil
}

func (s *TransactionService) GetFraudsBetween(ctx context.Context, from, to time.Time, page dto.PageRequest) (dto.PaginatedResponse[dto.TransactionResponse], error) {
	page.Normalize()

	txs, total, err := s.repo.FindFraudsBetween(ctx, from, to, page.Skip(), page.Limit)
	if err != nil {
		return dto.PaginatedResponse[dto.TransactionResponse]{}, fmt.Errorf("get frauds between: %w", err)
	}
	return dto.PaginatedResponse[dto.TransactionResponse]{
		Items: toResponses(txs),
		Meta:  dto.NewPageMeta(page.Page, page.Limit, total),
	}, nil
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

func (s *TransactionService) GetUserTrustScore(ctx context.Context, userID string) (dto.UserTrustScoreResponse, error) {
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return dto.UserTrustScoreResponse{}, fmt.Errorf("get user stats: %w", err)
	}

	score := calculateTrustScore(stats.Total, stats.FraudCount)

	return dto.UserTrustScoreResponse{
		UserID:     userID,
		Score:      score,
		RiskLevel:  riskLevel(score),
		Total:      stats.Total,
		FraudCount: stats.FraudCount,
	}, nil
}

// calculateTrustScore: işlem geçmişi yoksa 100 (nötr başlangıç).
// Aksi hâlde fraud olmayan işlemlerin oranı 0-100 arasına ölçeklenir.
func calculateTrustScore(total, fraudCount int64) float64 {
	if total == 0 {
		return 100
	}
	return (float64(total-fraudCount) / float64(total)) * 100
}

// riskLevel: score eşikleri iş kuralına göre ayarlanabilir.
func riskLevel(score float64) string {
	switch {
	case score >= 80:
		return "low"
	case score >= 50:
		return "medium"
	default:
		return "high"
	}
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
