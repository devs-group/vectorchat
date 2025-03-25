package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/auth"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthConfig holds the OAuth configuration
type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	RedirectURL        string
	Store              *session.Store
}

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	config      *OAuthConfig
	githubOAuth *oauth2.Config
	store       *session.Store
	userStore   *db.UserStore
	authMiddleware *auth.AuthMiddleware
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(config *OAuthConfig, userStore *db.UserStore, authMiddleware *auth.AuthMiddleware) *OAuthHandler {
	githubOAuth := &oauth2.Config{
		ClientID:     config.GitHubClientID,
		ClientSecret: config.GitHubClientSecret,
		RedirectURL:  config.RedirectURL + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &OAuthHandler{
		config:      config,
		githubOAuth: githubOAuth,
		store:       config.Store,
		userStore:   userStore,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the OAuth routes
func (h *OAuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// GitHub OAuth routes
	auth.Get("/github", h.GET_GitHubLogin)
	auth.Get("/github/callback", h.GET_GitHubCallback)

	// Session management
	auth.Get("/session", h.authMiddleware.RequireAuth, h.GET_Session)
	auth.Post("/logout", h.authMiddleware.RequireAuth, h.POST_Logout)

	// API key management
	auth.Post("/apikey", h.authMiddleware.RequireAuth, h.POST_GenerateAPIKey)
	auth.Get("/apikey", h.authMiddleware.RequireAuth, h.GET_ListAPIKeys)
	auth.Delete("/apikey/:id", h.authMiddleware.RequireAuth, h.DELETE_RevokeAPIKey)
}

// @Summary Initiate GitHub OAuth login
// @Description Redirects to GitHub for OAuth authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {object} LoginResponse
// @Failure 500 {object} APIResponse
// @Router /auth/github [get]
func (h *OAuthHandler) GET_GitHubLogin(c *fiber.Ctx) error {
	// Generate a random state
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate state",
		})
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Store state in session
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}
	sess.Set("oauth_state", state)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save session",
		})
	}

	// Redirect to GitHub
	url := h.githubOAuth.AuthCodeURL(state)
	return c.Redirect(url)
}

// @Summary GitHub OAuth callback
// @Description Handles the GitHub OAuth callback
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth code"
// @Param state query string true "OAuth state"
// @Success 302 {object} SessionResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /auth/github/callback [get]
func (h *OAuthHandler) GET_GitHubCallback(c *fiber.Ctx) error {
	// Get state from session
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}

	// Verify state
	expectedState := sess.Get("oauth_state")
	if expectedState == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid OAuth state",
		})
	}

	state := c.Query("state")
	if state != expectedState.(string) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid OAuth state",
		})
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := h.githubOAuth.Exchange(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to exchange code for token: %v", err),
		})
	}

	// Get user info from GitHub
	client := h.githubOAuth.Client(c.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get user info: %v", err),
		})
	}
	defer resp.Body.Close()

	// Parse user info
	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse user info: %v", err),
		})
	}

	// If email is not provided, get it from the emails API
	if githubUser.Email == "" {
		emails, err := h.getGitHubEmails(client)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to get user emails: %v", err),
			})
		}

		for _, email := range emails {
			if email.Primary {
				githubUser.Email = email.Email
				break
			}
		}
	}

	// Check if user exists
	user, err := h.userStore.FindUserByEmail(c.Context(), githubUser.Email)
	if err != nil && !apperrors.Is(err, apperrors.ErrUserNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to find user: %v", err),
		})
	}

	// Create user if not exists
	if user == nil {
		user = &db.User{
			ID:        uuid.New().String(),
			Name:      githubUser.Name,
			Email:     githubUser.Email,
			Provider:  "github",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := h.userStore.CreateUser(c.Context(), user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create user: %v", err),
			})
		}
	}

	// Set user in session
	sess.Set("user_id", user.ID)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save session",
		})
	}

	// Redirect to home page
	return c.Redirect("/")
}

// getGitHubEmails gets the user's emails from GitHub
func (h *OAuthHandler) getGitHubEmails(client *http.Client) ([]struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, err
	}

	return emails, nil
}

// @Summary Get current session
// @Description Returns current session information
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} SessionResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /auth/session [get]
func (h *OAuthHandler) GET_Session(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
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

	user, err := h.userStore.FindUserByID(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to find user: %v", err),
		})
	}

	return c.JSON(SessionResponse{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Provider:  user.Provider,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

// @Summary Logout user
// @Description Logs out the current user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @Router /auth/logout [post]
func (h *OAuthHandler) POST_Logout(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}

	if err := sess.Destroy(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to destroy session",
		})
	}

	return c.JSON(MessageResponse{
		Message: "Logged out successfully",
	})
}

// @Summary Generate API key
// @Description Generates a new API key for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} APIKeyResponse
// @Failure 401 {object} APIResponse
// @Router /auth/apikey [post]
func (h *OAuthHandler) POST_GenerateAPIKey(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
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

	// Generate API key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate API key",
		})
	}
	key := base64.URLEncoding.EncodeToString(b)

	// Store API key
	apiKey := &db.APIKey{
		ID:        uuid.New().String(),
		UserID:    userID.(string),
		Key:       key,
		CreatedAt: time.Now(),
	}

	if err := h.userStore.CreateAPIKey(c.Context(), apiKey); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create API key: %v", err),
		})
	}

	return c.JSON(APIKeyResponse{
		APIKey: APIKey{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			Key:       apiKey.Key,
			CreatedAt: apiKey.CreatedAt,
		},
	})
}

// @Summary List API keys
// @Description Lists all API keys for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} APIKeysResponse
// @Failure 401 {object} APIResponse
// @Router /auth/apikey [get]
func (h *OAuthHandler) GET_ListAPIKeys(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
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

	apiKeys, err := h.userStore.GetAPIKeys(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get API keys: %v", err),
		})
	}

	var keys []APIKey
	for _, k := range apiKeys {
		keys = append(keys, APIKey{
			ID:        k.ID,
			UserID:    k.UserID,
			Key:       k.Key,
			CreatedAt: k.CreatedAt,
		})
	}
	return c.JSON(APIKeysResponse{
		APIKeys: keys,
	})
}

// @Summary Revoke API key
// @Description Revokes an API key
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "API key ID"
// @Success 200 {object} MessageResponse
// @Failure 401 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Router /auth/apikey/{id} [delete]
func (h *OAuthHandler) DELETE_RevokeAPIKey(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
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

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "API key ID is required",
		})
	}

	if err := h.userStore.RevokeAPIKey(c.Context(), id, userID.(string)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to revoke API key: %v", err),
		})
	}

	return c.JSON(MessageResponse{
		Message: "API key revoked successfully",
	})
}
