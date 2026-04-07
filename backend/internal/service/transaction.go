package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"fraud-detection/internal/api/dto"
	"fraud-detection/internal/models"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/store"
)

type TransactionService struct {
	repo  store.TransactionRepository
	queue *queue.Client
}

func NewTransactionService(repo store.TransactionRepository, q *queue.Client) *TransactionService {
	return &TransactionService{repo: repo, queue: q}
}

func (s *TransactionService) Create(ctx context.Context, req dto.CreateTransactionRequest) (dto.TransactionResponse, error) {
	createdAt := time.Now().UTC()
	if req.CreatedAt != nil {
		createdAt = *req.CreatedAt
	}

	tx := models.Transaction{
		UserID:    req.UserID,
		Amount:    req.Amount,
		Lat:       req.Lat,
		Lon:       req.Lon,
		Status:    models.StatusPending,
		CreatedAt: createdAt,
	}
	if err := s.repo.Insert(ctx, &tx); err != nil {
		return dto.TransactionResponse{}, fmt.Errorf("create transaction: %w", err)
	}

	// Fraud worker'ın işleyebilmesi için transaction'ı kuyruğa yaz.
	// Publish hatası transaction'ı iptal etmez — DB'ye yazıldı, kuyruk best-effort.
	if err := s.queue.PublishTransaction(ctx, tx); err != nil {
		fmt.Printf("publish transaction: %v\n", err)
	}

	return dto.NewTransactionResponse(tx), nil
}

func (s *TransactionService) GetByUserID(ctx context.Context, userID string, page dto.PageRequest) (dto.PaginatedResponse[dto.TransactionResponse], error) {
	page.Normalize()

	txs, total, err := s.repo.FindByUserID(ctx, userID, page.Skip(), page.Limit)
	if err != nil {
		return dto.PaginatedResponse[dto.TransactionResponse]{}, fmt.Errorf("get by user_id: %w", err)
	}

	return dto.PaginatedResponse[dto.TransactionResponse]{
		Items: dto.NewTransactionResponses(txs),
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
		Items: dto.NewTransactionResponses(txs),
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
// Fraud işlemler skoru tam, suspicious işlemler yarı ağırlıkla düşürür.
func calculateTrustScore(total, fraudCount int64) float64 {
	if total == 0 {
		return 100
	}

	score := (float64(total) - float64(fraudCount)) / float64(total) * 100
	if score < 0 {
		score = 0
	}

	return score
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
