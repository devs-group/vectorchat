package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/middleware"
)

type WidgetHandler struct {
	authMiddleware *middleware.AuthMiddleware
	widgetsPath    string
}

func NewWidgetHandler(authMiddleware *middleware.AuthMiddleware) *WidgetHandler {
	return &WidgetHandler{
		authMiddleware: authMiddleware,
		widgetsPath:    "./widgets",
	}
}

func (h *WidgetHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/widgets/chats/:chatID/:widget", h.GET_Widget)
}

// @Summary Get widget JavaScript file
// @Description Serves the JavaScript widget file for embedding in external applications
// @Tags widget
// @Accept json
// @Produce application/javascript
// @Param chatID path string true "Chat/Chatbot ID"
// @Param widget path string true "Widget filename (e.g., vectorchat-plex-widget.js)"
// @Success 200 {string} string "JavaScript widget content"
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /widgets/chats/{chatID}/{widget} [get]
func (h *WidgetHandler) GET_Widget(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	widgetFile := c.Params("widget")

	if chatID == "" {
		return ErrorResponse(c, "Chat ID is required", nil, http.StatusBadRequest)
	}

	if widgetFile == "" || !strings.HasSuffix(widgetFile, ".js") {
		return ErrorResponse(c, "Valid widget filename is required", nil, http.StatusBadRequest)
	}

	widgetName := strings.TrimSuffix(widgetFile, ".js")
	widgetDir := filepath.Join(h.widgetsPath, widgetName)
	filePath := filepath.Join(widgetDir, widgetFile)

	if !h.isValidWidget(widgetName) {
		return ErrorResponse(c, "Widget not found", nil, http.StatusNotFound)
	}

	c.Set("Content-Type", "application/javascript")
	c.Set("Cache-Control", "public, max-age=3600")

	return c.SendFile(filePath)
}

func (h *WidgetHandler) isValidWidget(widgetName string) bool {
	validWidgets := []string{
		"vectorchat-glass-widget",
		"vectorchat-neon-widget",
		"vectorchat-plex-widget",
	}

	for _, valid := range validWidgets {
		if valid == widgetName {
			return true
		}
	}
	return false
}
