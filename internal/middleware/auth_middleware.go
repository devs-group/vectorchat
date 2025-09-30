package middleware

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/services"
)

// AuthMiddleware enforces that requests carry either a valid API key or a verified Kratos session.
type AuthMiddleware struct {
	authService   *services.AuthService
	APIKeyService *services.APIKeyService
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(authService *services.AuthService, apiKeyService *services.APIKeyService) *AuthMiddleware {
	return &AuthMiddleware{
		authService:   authService,
		APIKeyService: apiKeyService,
	}
}

// RequireAuth requires authentication for a route.
func (m *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	if apiKey := c.Get("X-API-Key"); apiKey != "" {
		return m.authenticateWithAPIKey(c, apiKey)
	}

	identityID := c.Get("X-User-Id")
	traitsHeader := c.Get("X-User-Traits")
	if identityID == "" || traitsHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	user, err := m.authService.SyncIdentity(c.Context(), identityID, traitsHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Store hydrated user and traits for downstream handlers.
	c.Locals("user", user)

	var traits map[string]any
	if err := json.Unmarshal([]byte(traitsHeader), &traits); err == nil {
		c.Locals("identity_traits", traits)
	}

	return c.Next()
}

func (m *AuthMiddleware) authenticateWithAPIKey(c *fiber.Ctx, apiKey string) error {
	apiKeyRecord, err := m.authService.FindAPIKeyByPlaintext(c.Context(), apiKey, func(hashedKey string) (bool, error) {
		isValid, err := m.APIKeyService.IsAPIKeyValid(hashedKey, apiKey)
		if err != nil {
			return false, errors.Wrap(err, "failed to verify api key")
		}
		if !isValid {
			return false, errors.Wrap(err, "provided api key is invalid")
		}
		return true, nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if apiKeyRecord.ExpiresAt != nil && apiKeyRecord.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "API key expired",
		})
	}

	if apiKeyRecord.RevokedAt != nil && apiKeyRecord.RevokedAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "API key revoked",
		})
	}

	user, err := m.authService.FindUserByID(c.Context(), apiKeyRecord.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	c.Locals("user", user)
	return c.Next()
}
