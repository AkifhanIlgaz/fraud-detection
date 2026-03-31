package response

import "github.com/gofiber/fiber/v3"

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func OK(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(Response{Success: true, Data: data})
}

func Error(c fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(Response{Error: err.Error()})
}
