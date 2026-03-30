package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"fraud-detection/internal/api/dto"
	"fraud-detection/internal/service"
	"fraud-detection/pkg/response"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// POST /api/v1/transactions
func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, err)
	}
	if err := validateCreate(req); err != nil {
		return response.BadRequest(c, err)
	}

	res, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return response.InternalError(c, err)
	}
	return response.Created(c, res)
}

// GET /api/v1/transactions/user/:userID
func (h *TransactionHandler) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("userID")
	if userID == "" {
		return response.BadRequest(c, errors.New("userID is required"))
	}

	res, err := h.svc.GetByUserID(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err)
	}
	return response.OK(c, res)
}

// GET /api/v1/transactions/frauds?from=&to=
func (h *TransactionHandler) GetFraudsBetween(c *fiber.Ctx) error {
	var req dto.TransactionsBetweenRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequest(c, err)
	}

	from, to, err := parseDateRange(req.From, req.To)
	if err != nil {
		return response.BadRequest(c, err)
	}

	res, err := h.svc.GetFraudsBetween(c.Context(), from, to)
	if err != nil {
		return response.InternalError(c, err)
	}
	return response.OK(c, res)
}

// PATCH /api/v1/transactions/:id/status
func (h *TransactionHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, err)
	}
	if req.Status == "" {
		return response.BadRequest(c, errors.New("status is required"))
	}

	if err := h.svc.UpdateStatus(c.Context(), id, req); err != nil {
		return response.BadRequest(c, err)
	}
	return response.OK(c, nil)
}

func validateCreate(req dto.CreateTransactionRequest) error {
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if req.Lat < -90 || req.Lat > 90 {
		return errors.New("lat must be between -90 and 90")
	}
	if req.Lon < -180 || req.Lon > 180 {
		return errors.New("lon must be between -180 and 180")
	}
	return nil
}

func parseDateRange(from, to string) (time.Time, time.Time, error) {
	if from == "" || to == "" {
		return time.Time{}, time.Time{}, errors.New("from and to are required")
	}
	f, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from must be RFC3339 (e.g. 2024-01-01T00:00:00Z)")
	}
	t, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to must be RFC3339 (e.g. 2024-01-01T00:00:00Z)")
	}
	if !f.Before(t) {
		return time.Time{}, time.Time{}, errors.New("from must be before to")
	}
	return f, t, nil
}
