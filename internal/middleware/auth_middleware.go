package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/postgres"
	"github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/store"
)

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	sessionStore  *postgres.Storage
	userStore     *store.UserStore
	APIKeyService *services.APIKeyService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(sessionStore *postgres.Storage, userStore *store.UserStore, apiKeyService *services.APIKeyService) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore:  sessionStore,
		userStore:     userStore,
		APIKeyService: apiKeyService,
	}
}

// RequireAuth requires authentication for a route
func (m *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		apiKeyRecord, err := m.userStore.FindAPIKey(c.Context(), func(hashedKey string) (bool, error) {
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

		if apiKeyRecord.ExpiresAt.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key expired",
			})
		}

		if apiKeyRecord.RevokedAt != nil && apiKeyRecord.RevokedAt.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key revoked",
			})
		}

		// Get user associated with API key
		user, err := m.userStore.FindUserByID(c.Context(), apiKeyRecord.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
		c.Locals("user", user)
		return c.Next()
	}

	// Check for session cookie
	sessionID := c.Cookies("session_id")
	slog.Error("session_id", "session_id", sessionID)
	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Match the session key format from OAuthHandler
	sessionKey := "session_" + sessionID
	userIDBytes, err := m.sessionStore.Get(sessionKey)
	if err != nil || userIDBytes == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired session",
		})
	}

	// Get user
	user, err := m.userStore.FindUserByID(c.Context(), string(userIDBytes))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Refresh session expiration
	if err := m.sessionStore.Set(sessionKey, userIDBytes, 8*time.Hour); err != nil {
		// Log error but don't fail request
		slog.Error("Failed to refresh session", "error", err)
	}

	// Set user in context
	c.Locals("user", user)
	return c.Next()
}
