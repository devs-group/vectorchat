package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/postgres"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/config"
	"github.com/yourusername/vectorchat/pkg/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	RedirectURL        string
	SessionStore       *postgres.Storage
	Env                string
}

type OAuthHandler struct {
	config         *OAuthConfig
	githubOAuth    *oauth2.Config
	store          *postgres.Storage
	authService    *services.AuthService
	authMiddleware *middleware.AuthMiddleware
	apiKeyService  *services.APIKeyService
}

func NewOAuthHandler(
	config *OAuthConfig,
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
) *OAuthHandler {
	if config.GitHubClientID == "" || config.GitHubClientSecret == "" {
		panic("Missing required OAuth configuration")
	}

	githubOAuth := &oauth2.Config{
		ClientID:     config.GitHubClientID,
		ClientSecret: config.GitHubClientSecret,
		RedirectURL:  config.RedirectURL + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &OAuthHandler{
		config:         config,
		githubOAuth:    githubOAuth,
		store:          config.SessionStore,
		authService:    authService,
		authMiddleware: authMiddleware,
	}
}

func (h *OAuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Get("/github", h.GET_GitHubLogin)
	auth.Get("/github/callback", h.GET_GitHubCallback)
	auth.Get("/session", h.authMiddleware.RequireAuth, h.GET_Session)
	auth.Post("/logout", h.authMiddleware.RequireAuth, h.POST_Logout)
}

// @Summary Initiate GitHub OAuth login
// @Description Redirects to GitHub for OAuth authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitHub OAuth"
// @Failure 500 {object} models.APIResponse
// @Router /auth/github [get]
func (h *OAuthHandler) GET_GitHubLogin(c *fiber.Ctx) error {
	log.Printf("[DEBUG] GitHub login initiated - URL: %s", c.OriginalURL())

	// Generate OAuth state using service
	oauthState, err := h.authService.GenerateOAuthState()
	if err != nil {
		log.Printf("[ERROR] Failed to generate OAuth state: %v", err)
		return ErrorResponse(c, "failed to generate state", err)
	}
	log.Printf("[DEBUG] Generated OAuth state - StateKey: '%s', State: '%s'", oauthState.StateKey, oauthState.State)

	err = h.store.Set(oauthState.StateKey, []byte(oauthState.State), time.Hour)
	if err != nil {
		log.Printf("[ERROR] Failed to save state to store: %v", err)
		return ErrorResponse(c, "failed to save state", err)
	}
	log.Printf("[DEBUG] Saved state to store successfully")

	url := h.githubOAuth.AuthCodeURL(oauthState.State)
	log.Printf("[DEBUG] Generated GitHub OAuth URL: %s", url)

	// Determine cookie settings based on environment
	isSecure := h.config.Env == "production"
	sameSite := "Lax" // Use Lax for OAuth flows

	log.Printf("[DEBUG] Setting cookie - Name: oauth_state_key, Value: %s, Secure: %t, SameSite: %s",
		oauthState.StateKey, isSecure, sameSite)

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state_key",
		Value:    oauthState.StateKey,
		Expires:  time.Now().Add(time.Hour),
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: sameSite,
		Path:     "/",
	})

	log.Printf("[DEBUG] Redirecting to GitHub OAuth URL")
	return c.Redirect(url)
}

// @Summary GitHub OAuth callback
// @Description Handles the GitHub OAuth callback and sets session
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth code"
// @Param state query string true "OAuth state"
// @Success 302 {string} string "Redirect to /"
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/github/callback [get]
func (h *OAuthHandler) GET_GitHubCallback(c *fiber.Ctx) error {
	log.Printf("[DEBUG] OAuth callback started - URL: %s", c.OriginalURL())

	stateKey := c.Cookies("oauth_state_key")
	log.Printf("[DEBUG] Retrieved stateKey from cookie: '%s'", stateKey)

	if stateKey == "" {
		log.Printf("[ERROR] No oauth_state_key cookie found")
		return ErrorResponse(c, "auth state is invalid", nil, http.StatusBadRequest)
	}

	expectedState, err := h.store.Get(stateKey)
	log.Printf("[DEBUG] Retrieved expectedState from store - key: '%s', value: '%s', error: %v",
		stateKey, string(expectedState), err)

	if err != nil || expectedState == nil {
		log.Printf("[ERROR] Failed to get state from store - key: '%s', error: %v, expectedState is nil: %t",
			stateKey, err, expectedState == nil)
		return ErrorResponse(c, "auth state is invalid", err, http.StatusBadRequest)
	}

	receivedState := c.Query("state")
	log.Printf("[DEBUG] State comparison - received: '%s', expected: '%s', match: %t",
		receivedState, string(expectedState), receivedState == string(expectedState))

	if receivedState != string(expectedState) {
		log.Printf("[ERROR] OAuth state mismatch - received: '%s', expected: '%s'",
			receivedState, string(expectedState))
		return ErrorResponse(c, "auth state is invalid", nil, http.StatusBadRequest)
	}

	log.Printf("[DEBUG] OAuth state validation successful, proceeding with token exchange")
	defer h.store.Delete(stateKey)

	code := c.Query("code")
	token, err := h.githubOAuth.Exchange(c.Context(), code)
	if err != nil {
		return ErrorResponse(c, "failed to exchange oauth code", err)
	}

	client := h.githubOAuth.Client(c.Context(), token)

	// Process GitHub callback using service
	user, err := h.authService.ProcessGitHubCallback(c.Context(), client)
	if err != nil {
		return ErrorResponse(c, "failed to process GitHub callback", err)
	}

	sessionID := uuid.New().String()
	sessionKey := fmt.Sprintf("session_%s", sessionID)
	if err := h.store.Set(sessionKey, []byte(user.ID), 8*time.Hour); err != nil {
		return ErrorResponse(c, "failed to save session", err)
	}

	// Use consistent cookie settings
	isSecure := h.config.Env == "production"

	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(8 * time.Hour),
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		Path:     "/",
	})
	var appCfg config.AppConfig
	err = config.Load(&appCfg)
	if err != nil {
		return ErrorResponse(c, "failed to load app config", err)
	}
	frontendURL := fmt.Sprintf("http://%s", appCfg.FrontendURL)
	if appCfg.IsSSL {
		frontendURL = fmt.Sprintf("https://%s", appCfg.FrontendURL)
	}
	return c.Redirect(fmt.Sprintf("%s/chat", frontendURL))
}

// @Summary Get current session
// @Description Returns information about the current authenticated session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.SessionResponse
// @Failure 401 {object} models.APIResponse
// @Security ApiKeyAuth
// @Security CookieAuth
// @Router /auth/session [get]
func (h *OAuthHandler) GET_Session(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)
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
// @Description Logs out the current user and clears the session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func (h *OAuthHandler) POST_Logout(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return ErrorResponse(c, "no session id found in cookie", nil, http.StatusUnauthorized)
	}

	sessionKey := fmt.Sprintf("session_%s", sessionID)
	if err := h.store.Delete(sessionKey); err != nil {
		return ErrorResponse(c, "failed to get session", err)
	}

	// Use consistent cookie settings with login
	isSecure := h.config.Env == "production"

	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(models.MessageResponse{
		Message: "Logged out successfully",
	})
}
