package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/services"
)

func GetUser(c *fiber.Ctx) (*services.User, error) {
	user, ok := c.Locals("user").(*services.User)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	return user, nil
}

func ErrorResponse(c *fiber.Ctx, msg string, err error, statusCode ...int) error {
	slog.Error(msg, "err", err)
	res := fiber.Map{
		"error": msg,
	}
	if len(statusCode) > 0 {
		return c.Status(statusCode[0]).JSON(res)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(res)
}
