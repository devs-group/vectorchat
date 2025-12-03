package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

func GetUser(c *fiber.Ctx) (*db.User, error) {
	user, ok := c.Locals("user").(*db.User)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	return user, nil
}

func GetOrgContext(c *fiber.Ctx) *services.OrganizationContext {
	if org, ok := c.Locals("org").(*services.OrganizationContext); ok && org != nil {
		return org
	}
	return &services.OrganizationContext{Role: services.OrgRolePersonal}
}

func ErrorResponse(c *fiber.Ctx, msg string, err error, statusCode ...int) error {
	slog.Error(msg, "err", err)
	res := models.APIResponse{
		Error: msg,
	}
	if len(statusCode) > 0 {
		return c.Status(statusCode[0]).JSON(res)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(res)
}
