package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/services"
)

type OrganizationMiddleware struct {
	orgService *services.OrganizationService
}

func NewOrganizationMiddleware(orgService *services.OrganizationService) *OrganizationMiddleware {
	return &OrganizationMiddleware{orgService: orgService}
}

// Attach resolves the organization context from the optional X-Organization-ID header.
func (m *OrganizationMiddleware) Attach(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*db.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}

	header := strings.TrimSpace(c.Get("X-Organization-ID"))
	var (
		orgCtx *services.OrganizationContext
		err    error
	)

	if header != "" {
		orgID, parseErr := uuid.Parse(header)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid organization id"})
		}
		orgCtx, err = m.orgService.EnsureMembership(c.Context(), &orgID, user.ID)
		if err != nil {
			status := fiber.StatusForbidden
			if apperrors.Is(err, apperrors.ErrOrganizationNotFound) {
				status = fiber.StatusNotFound
			}
			return c.Status(status).JSON(fiber.Map{"error": "organization access denied"})
		}
	} else {
		orgCtx = &services.OrganizationContext{Role: services.OrgRolePersonal}
	}

	c.Locals("org", orgCtx)
	return c.Next()
}
