package middleware

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/services"
)

// AuthMiddleware trusts Oathkeeper to authenticate requests and forwards user context downstream.
type AuthMiddleware struct {
	authService *services.AuthService
}

// NewAuthMiddleware creates a new auth middleware instance.
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth validates the Oathkeeper headers and ensures a user context is available.
func (m *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	traitsHeader := c.Get("X-User-Traits")

	var user *db.User
	var err error

	if traitsHeader != "" {
		user, err = m.hydrateFromTraits(c, userID, traitsHeader)
	} else {
		user, err = m.hydrateFromStore(c, userID)
	}
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Locals("user", user)
	if traitsHeader != "" {
		traits := make(map[string]any)
		if err := json.Unmarshal([]byte(traitsHeader), &traits); err == nil {
			c.Locals("identity_traits", traits)
		}
	}

	return c.Next()
}

func (m *AuthMiddleware) hydrateFromTraits(c *fiber.Ctx, userID, traitsHeader string) (*db.User, error) {
	user, err := m.authService.SyncIdentity(c.Context(), userID, traitsHeader)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *AuthMiddleware) hydrateFromStore(c *fiber.Ctx, userID string) (*db.User, error) {
	user, err := m.authService.FindUserByID(c.Context(), userID)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, "failed to load user")
	}
	return user, nil
}
