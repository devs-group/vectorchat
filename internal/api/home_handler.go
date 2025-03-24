package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// HomeHandler handles home-related routes
type HomeHandler struct {
	store     *session.Store
	userStore *db.UserStore
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(store *session.Store, userStore *db.UserStore) *HomeHandler {
	return &HomeHandler{
		store:     store,
		userStore: userStore,
	}
}

// RegisterRoutes registers the home routes
func (h *HomeHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/", h.GET_Home)
}

// GET_Home handles the home route and returns user information if authenticated
func (h *HomeHandler) GET_Home(c *fiber.Ctx) error {
	// Get session
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}

	// Check if user is authenticated
	userID := sess.Get("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Get user information
	user, err := h.userStore.FindUserByID(c.Context(), userID.(string))
	if err != nil {
		if err == apperrors.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user information",
		})
	}

	// Return user information
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
		},
	})
} 