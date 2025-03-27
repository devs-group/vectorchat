package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/postgres"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthConfig holds the OAuth configuration
type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	RedirectURL        string
	Store              *postgres.Storage
	Env                string
}

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	config         *OAuthConfig
	githubOAuth    *oauth2.Config
	store          *postgres.Storage
	userStore      *db.UserStore
	authMiddleware *middleware.AuthMiddleware
}

// NewOAuthHandler creates a new OAuth handler with validation
func NewOAuthHandler(config *OAuthConfig, userStore *db.UserStore, authMiddleware *middleware.AuthMiddleware) *OAuthHandler {
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
		store:          config.Store,
		userStore:      userStore,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the OAuth routes
func (h *OAuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Get("/github", h.GET_GitHubLogin)
	auth.Get("/github/callback", h.GET_GitHubCallback)
	auth.Get("/session", h.authMiddleware.RequireAuth, h.GET_Session)
	auth.Post("/logout", h.authMiddleware.RequireAuth, h.POST_Logout)
	auth.Post("/apikey", h.authMiddleware.RequireAuth, h.POST_GenerateAPIKey)
	auth.Get("/apikey", h.authMiddleware.RequireAuth, h.GET_ListAPIKeys)
	auth.Delete("/apikey/:id", h.authMiddleware.RequireAuth, h.DELETE_RevokeAPIKey)
}

// @Summary Initiate GitHub OAuth login
// @Description Redirects to GitHub for OAuth authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitHub OAuth"
// @Failure 500 {object} APIResponse
// @Router /auth/github [get]
func (h *OAuthHandler) GET_GitHubLogin(c *fiber.Ctx) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		slog.Error("Failed to generate state", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
	state := base64.URLEncoding.EncodeToString(b)
	stateKey := fmt.Sprintf("oauth_state_%s", uuid.New().String())

	err := h.store.Set(stateKey, []byte(state), time.Hour)
	if err != nil {
		slog.Error("Failed to save state", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	url := h.githubOAuth.AuthCodeURL(state)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state_key",
		Value:    stateKey,
		Expires:  time.Now().Add(time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})
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
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /auth/github/callback [get]
func (h *OAuthHandler) GET_GitHubCallback(c *fiber.Ctx) error {
	stateKey := c.Cookies("oauth_state_key")
	if stateKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid authentication state",
		})
	}

	expectedState, err := h.store.Get(stateKey)
	if err != nil || expectedState == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid authentication state",
		})
	}
	defer h.store.Delete(stateKey)

	if c.Query("state") != string(expectedState) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid authentication state",
		})
	}

	code := c.Query("code")
	token, err := h.githubOAuth.Exchange(c.Context(), code)
	if err != nil {
		slog.Error("Failed to exchange OAuth code", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	client := h.githubOAuth.Client(c.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		slog.Error("Failed to get user info", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		slog.Error("Failed to parse user info", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	if githubUser.Email == "" {
		emails, err := h.getGitHubEmails(client)
		if err != nil {
			slog.Error("Failed to get emails", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication failed",
			})
		}
		for _, email := range emails {
			if email.Primary && email.Verified {
				githubUser.Email = email.Email
				break
			}
		}
	}

	if len(githubUser.Email) > 255 || len(githubUser.Name) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user data",
		})
	}

	user, err := h.userStore.FindUserByEmail(c.Context(), githubUser.Email)
	if err != nil && !apperrors.Is(err, apperrors.ErrUserNotFound) {
		slog.Error("Failed to find user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

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
			slog.Error("Failed to create user", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
		slog.Info("User created", "email", user.Email)
	}

	sessionID := uuid.New().String()
	sessionKey := fmt.Sprintf("session_%s", sessionID)
	if err := h.store.Set(sessionKey, []byte(user.ID), 8*time.Hour); err != nil {
		slog.Error("Failed to save session", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(8 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})

	return c.Redirect("/")
}

// @Summary Get current session
// @Description Returns information about the current authenticated session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} SessionResponse
// @Failure 401 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/session [get]
func (h *OAuthHandler) GET_Session(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)
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

// @Summary Logout
// @Description Logs out the current user and clears the session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func (h *OAuthHandler) POST_Logout(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	sessionKey := fmt.Sprintf("session_%s", sessionID)
	if err := h.store.Delete(sessionKey); err != nil {
		slog.Error("Failed to delete session", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(MessageResponse{
		Message: "Logged out successfully",
	})
}

// @Summary Generate API key
// @Description Generates a new API key for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} APIKeyResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [post]
func (h *OAuthHandler) POST_GenerateAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		slog.Error("Failed to generate API key", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
	key := "vc_" + base64.URLEncoding.EncodeToString(b)
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash API key", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	apiKey := &db.APIKey{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Key:       string(hashedKey),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	if err := h.userStore.CreateAPIKey(c.Context(), apiKey); err != nil {
		slog.Error("Failed to create API key", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(APIKeyResponse{
		APIKey: APIKey{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			Key:       key,
			CreatedAt: apiKey.CreatedAt,
			ExpiresAt: apiKey.ExpiresAt,
		},
	})
}

// @Summary List API keys
// @Description Lists all API keys for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} APIKeysResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [get]
func (h *OAuthHandler) GET_ListAPIKeys(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)

	apiKeys, err := h.userStore.GetAPIKeys(c.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get API keys", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	var keys []APIKey
	for _, k := range apiKeys {
		keys = append(keys, APIKey{
			ID:        k.ID,
			UserID:    k.UserID,
			CreatedAt: k.CreatedAt,
			ExpiresAt: k.ExpiresAt,
		})
	}
	return c.JSON(APIKeysResponse{
		APIKeys: keys,
	})
}

// @Summary Revoke API key
// @Description Revokes an API key for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "API key ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey/{id} [delete]
func (h *OAuthHandler) DELETE_RevokeAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "API key ID is required",
		})
	}

	if err := h.userStore.RevokeAPIKey(c.Context(), id, user.ID); err != nil {
		slog.Error("Failed to revoke API key", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(MessageResponse{
		Message: "API key revoked successfully",
	})
}

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
