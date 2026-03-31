package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"fraud-detection/internal/api/handler"
	"fraud-detection/internal/queue"
	"fraud-detection/internal/service"
	"fraud-detection/internal/store"
)

func NewRouter(repo store.TransactionRepository, q *queue.Client) *fiber.App {
	svc := service.NewTransactionService(repo, q)
	txHandler := handler.NewTransactionHandler(svc)

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	v1 := app.Group("/api/v1")
	tx := v1.Group("/transactions")
	tx.Post("", txHandler.Create)
	tx.Get("/user/:userID", txHandler.GetByUserID)
	tx.Get("/user/:userID/trust-score", txHandler.GetTrustScore)
	tx.Get("/frauds", txHandler.GetFraudsBetween)
	tx.Patch("/:id/status", txHandler.UpdateStatus)

	return app
}
