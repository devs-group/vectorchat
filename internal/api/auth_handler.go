package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

// AuthConfig captures dependencies for Kratos-backed authentication flows.
type AuthConfig struct {
	KratosPublicURL string
	KratosAdminURL  string
	SessionCookie   string
}

// AuthHandler exposes authentication-related endpoints backed by Ory Kratos.
type AuthHandler struct {
	authService    *services.AuthService
	authMiddleware *middleware.AuthMiddleware
	httpClient     *http.Client
	cfg            AuthConfig
}

// NewAuthHandler wires a Kratos-backed authentication handler.
func NewAuthHandler(
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
	cfg AuthConfig,
) *AuthHandler {
	client := &http.Client{Timeout: 10 * time.Second}
	return &AuthHandler{
		authService:    authService,
		authMiddleware: authMiddleware,
		httpClient:     client,
		cfg:            cfg,
	}
}

// RegisterRoutes registers auth endpoints.
func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Get("/session", h.authMiddleware.RequireAuth, h.getSession)
	auth.Post("/logout", h.authMiddleware.RequireAuth, h.postLogout)
}

// @Summary Get current session
// @Description Returns information about the authenticated user bound to the request
// @Tags auth
// @Produce json
// @Success 200 {object} models.SessionResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/session [get]
func (h *AuthHandler) getSession(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*db.User)
	if !ok || user == nil {
		return ErrorResponse(c, "session not found", nil, fiber.StatusUnauthorized)
	}

	return c.JSON(models.SessionResponse{
		User: models.User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Provider:  user.Provider,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

// @Summary Logout
// @Description Terminates the active Kratos session and clears the browser cookie
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/logout [post]
func (h *AuthHandler) postLogout(c *fiber.Ctx) error {
	logoutURL := fmt.Sprintf("%s/self-service/logout/api", h.cfg.KratosPublicURL)

	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, logoutURL, nil)
	if err != nil {
		return ErrorResponse(c, "failed to create logout request", err)
	}

	if cookieHeader := c.Get("Cookie"); cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return ErrorResponse(c, "failed to contact identity provider", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ErrorResponse(c, "identity provider rejected logout", nil, resp.StatusCode)
	}

	// Clear the Kratos session cookie locally so the browser drops it immediately.
	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.SessionCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(models.MessageResponse{Message: "Logged out successfully"})
}
