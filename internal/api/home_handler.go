package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/postgres"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/store"
)

// HomeHandler handles home-related routes
type HomeHandler struct {
	store          *postgres.Storage
	userStore      *store.UserStore
	authMiddleware *middleware.AuthMiddleware
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(store *postgres.Storage, userStore *store.UserStore, authMiddleware *middleware.AuthMiddleware) *HomeHandler {
	return &HomeHandler{
		store:          store,
		userStore:      userStore,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the home routes
func (h *HomeHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/", h.authMiddleware.RequireAuth, h.GET_Home)
}

// @Summary Get user information
// @Description Returns authenticated user information if logged in, otherwise redirects to swagger
// @Tags home
// @Accept json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 307 {string} string "Redirect to /swagger"
// @Failure 500 {object} APIResponse
// @Router / [get]
func (h *HomeHandler) GET_Home(c *fiber.Ctx) error {
	_, ok := c.Locals("user").(*store.User)
	if !ok {
		return c.JSON(fiber.Map{
			"message": "no user found in request context",
		})
	}
	return c.Redirect("/swagger")
}
