package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/yourusername/vectorchat/internal/db"
)

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	store     *session.Store
	userStore *db.UserStore
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(store *session.Store, userStore *db.UserStore) *AuthMiddleware {
	return &AuthMiddleware{
		store:     store,
		userStore: userStore,
	}
}

// RequireAuth requires authentication for a route
func (m *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	// Check for API key in header
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		// Validate API key
		user, err := m.userStore.FindUserByAPIKey(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}
		
		// Set user in context
		c.Locals("user", user)
		return c.Next()
	}
	
	// Check for session
	sess, err := m.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}
	
	userID := sess.Get("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}
	
	user, err := m.userStore.FindUserByID(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to find user",
		})
	}
	
	// Set user in context
	c.Locals("user", user)
	return c.Next()
} 