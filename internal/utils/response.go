package utils

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(APIResponse{Success: true, Message: message, Data: data})
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{Success: true, Message: message, Data: data})
}

func Error(c *fiber.Ctx, status int, message string, errors interface{}) error {
	return c.Status(status).JSON(APIResponse{Success: false, Message: message, Errors: errors})
}
