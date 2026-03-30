package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"fraud-detection/internal/api/handler"
	"fraud-detection/internal/service"
	"fraud-detection/internal/store"
)

func NewRouter(repo store.TransactionRepository) *fiber.App {
	svc := service.NewTransactionService(repo)
	txHandler := handler.NewTransactionHandler(svc)

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	v1 := app.Group("/api/v1")
	tx := v1.Group("/transactions")
	tx.Post("", txHandler.Create)
	tx.Get("/user/:userID", txHandler.GetByUserID)
	tx.Get("/frauds", txHandler.GetFraudsBetween)
	tx.Patch("/:id/status", txHandler.UpdateStatus)

	return app
}
