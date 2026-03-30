package response

import "github.com/gofiber/fiber/v3"

type Response struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func OK(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Response{Success: true, Data: data})
}

func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Response{Success: true, Data: data})
}

func BadRequest(c fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
}

func NotFound(c fiber.Ctx, err error) error {
	return c.Status(fiber.StatusNotFound).JSON(Response{Error: err.Error()})
}

func InternalError(c fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(Response{Error: err.Error()})
}
