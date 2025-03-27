package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)


func getUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	id := c.Params(paramName)
	if id == "" {
		return uuid.Nil, fiber.ErrBadRequest
	}
	return uuid.Parse(id)
}
