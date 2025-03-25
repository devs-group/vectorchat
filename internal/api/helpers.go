package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
)

func getUser(c *fiber.Ctx) (*db.User, error) {
	if user, ok := c.Locals("user").(*db.User); ok {
		return user, nil
	}
	return nil, fiber.ErrUnauthorized
}

func getUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	id := c.Params(paramName)
	if id == "" {
		return uuid.Nil, fiber.ErrBadRequest
	}
	return uuid.Parse(id)
}

func getUUIDFormValue(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	id := c.FormValue(paramName)
	if id == "" {
		return uuid.Nil, fiber.ErrBadRequest
	}
	return uuid.Parse(id)
}
