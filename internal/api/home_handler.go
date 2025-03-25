package api

import (
	"log/slog"

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

// @Summary Get user information
// @Description Returns authenticated user information
// @Tags home
// @Accept json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router / [get]
func (h *HomeHandler) GET_Home(c *fiber.Ctx) error {
	u := c.Locals("user")
	slog.Info("c.Locals", "user", u)
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Error: "Failed to get session",
		})
	}

	userID := sess.Get("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Error: "Not authenticated",
		})
	}

	user, err := h.userStore.FindUserByID(c.Context(), userID.(string))
	if err != nil {
		if err == apperrors.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(APIResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Error: "Failed to get user information",
		})
	}

	return c.JSON(UserResponse{
		User: User{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	})
} 