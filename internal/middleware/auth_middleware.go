package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/postgres"
	"github.com/yourusername/vectorchat/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	store     *postgres.Storage
	userStore *db.UserStore
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(store *postgres.Storage, userStore *db.UserStore) *AuthMiddleware {
	return &AuthMiddleware{
		store:     store,
		userStore: userStore,
	}
}

// RequireAuth requires authentication for a route
func (m *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	// Check for API key in header first
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		// Validate API key
		apiKeyRecord, err := m.userStore.FindAPIKey(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Check expiration
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

		// Verify key hash
		if err := bcrypt.CompareHashAndPassword([]byte(apiKeyRecord.Key), []byte(apiKey)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Get user associated with API key
		user, err := m.userStore.FindUserByID(c.Context(), apiKeyRecord.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Set user in context
		c.Locals("user", user)
		return c.Next()
	}

	// Check for session cookie
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Match the session key format from OAuthHandler
	sessionKey := "session_" + sessionID
	userIDBytes, err := m.store.Get(sessionKey)
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
	if err := m.store.Set(sessionKey, userIDBytes, 8*time.Hour); err != nil {
		// Log error but don't fail request
		slog.Error("Failed to refresh session", "error", err)
	}

	// Set user in context
	c.Locals("user", user)
	return c.Next()
}
