package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"

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
func (h *TransactionHandler) Create(c fiber.Ctx) error {
	var req dto.CreateTransactionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := req.Validate(); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	res, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.OK(c, fiber.StatusCreated, res)
}

// GET /api/v1/transactions/user/:userID?page=1&limit=20
func (h *TransactionHandler) GetByUserID(c fiber.Ctx) error {
	userID := c.Params("userID")
	if userID == "" {
		return response.Error(c, fiber.StatusBadRequest, errors.New("userID is required"))
	}

	var page dto.PageRequest
	if err := c.Bind().Query(&page); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	res, err := h.svc.GetByUserID(c.Context(), userID, page)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.OK(c, fiber.StatusOK, res)
}

// GET /api/v1/transactions/frauds?from=&to=&page=1&limit=20
func (h *TransactionHandler) GetFraudsBetween(c fiber.Ctx) error {
	var req dto.TransactionsBetweenRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	from, to, err := req.Parse()
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var page dto.PageRequest
	if err := c.Bind().Query(&page); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	res, err := h.svc.GetFraudsBetween(c.Context(), from, to, page)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.OK(c, fiber.StatusOK, res)
}

// GET /api/v1/transactions/user/:userID/trust-score
func (h *TransactionHandler) GetTrustScore(c fiber.Ctx) error {
	userID := c.Params("userID")
	if userID == "" {
		return response.Error(c, fiber.StatusBadRequest, errors.New("userID is required"))
	}

	res, err := h.svc.GetUserTrustScore(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.OK(c, fiber.StatusOK, res)
}

// PATCH /api/v1/transactions/:id/status
func (h *TransactionHandler) UpdateStatus(c fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := req.Validate(); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.svc.UpdateStatus(c.Context(), id, req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	return response.OK(c, fiber.StatusOK, nil)
}
