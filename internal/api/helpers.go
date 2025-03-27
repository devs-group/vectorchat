package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/store"
)

func GetUser(c *fiber.Ctx) (*store.User, error) {
	user, ok := c.Locals("user").(*store.User)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	return user, nil
}
