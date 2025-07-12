package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/postgres"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/config"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthConfig holds the OAuth configuration
type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	RedirectURL        string
	SessionStore       *postgres.Storage
	Env                string
}

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	config         *OAuthConfig
	githubOAuth    *oauth2.Config
	store          *postgres.Storage
	authService    *services.AuthService
	authMiddleware *middleware.AuthMiddleware
	apiKeyService  *services.APIKeyService
}

// NewOAuthHandler creates a new OAuth handler with validation
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

// RegisterRoutes registers the OAuth routes
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
// @Failure 500 {object} APIResponse
// @Router /auth/github [get]
func (h *OAuthHandler) GET_GitHubLogin(c *fiber.Ctx) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ErrorResponse(c, "failed to generate state", err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	stateKey := fmt.Sprintf("oauth_state_%s", uuid.New().String())

	err := h.store.Set(stateKey, []byte(state), time.Hour)
	if err != nil {
		return ErrorResponse(c, "failed to save state", err)
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
		return ErrorResponse(c, "auth state is invalid", nil, http.StatusBadRequest)
	}
	expectedState, err := h.store.Get(stateKey)
	if err != nil || expectedState == nil {
		return ErrorResponse(c, "auth state is invalid", err, http.StatusBadRequest)
	}

	if c.Query("state") != string(expectedState) {
		return ErrorResponse(c, "auth state is invalid", nil, http.StatusBadRequest)
	}

	defer h.store.Delete(stateKey)

	code := c.Query("code")
	token, err := h.githubOAuth.Exchange(c.Context(), code)
	if err != nil {
		return ErrorResponse(c, "failed to exchange oauth code", err)
	}

	client := h.githubOAuth.Client(c.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return ErrorResponse(c, "failed to get use info", err)
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return ErrorResponse(c, "failed to parse user info", err)
	}

	if githubUser.Email == "" {
		emails, err := h.getGitHubEmails(client)
		if err != nil {
			return ErrorResponse(c, "failed to get email", err)
		}
		for _, email := range emails {
			if email.Primary && email.Verified {
				githubUser.Email = email.Email
				break
			}
		}
	}

	if len(githubUser.Email) > 255 || len(githubUser.Name) > 100 {
		return ErrorResponse(c, "invalid user data", nil, http.StatusBadRequest)
	}

	user, err := h.authService.FindUserByEmail(c.Context(), githubUser.Email)
	if err != nil && !apperrors.Is(err, apperrors.ErrUserNotFound) {
		return ErrorResponse(c, "failed to find user", err)
	}

	if user == nil {
		user, err = h.authService.CreateUser(c.Context(), uuid.New().String(), githubUser.Name, githubUser.Email, "github")
		if err != nil {
			return ErrorResponse(c, "failed to create user", err)
		}
	}

	sessionID := uuid.New().String()
	sessionKey := fmt.Sprintf("session_%s", sessionID)
	if err := h.store.Set(sessionKey, []byte(user.ID), 8*time.Hour); err != nil {
		return ErrorResponse(c, "failed to save session", err)
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
// @Success 200 {object} SessionResponse
// @Failure 401 {object} APIResponse
// @Security ApiKeyAuth
// @Security CookieAuth
// @Router /auth/session [get]
func (h *OAuthHandler) GET_Session(c *fiber.Ctx) error {
	user := c.Locals("user").(*services.User)
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
		return ErrorResponse(c, "no session id found in cookie", nil, http.StatusUnauthorized)
	}

	sessionKey := fmt.Sprintf("session_%s", sessionID)
	if err := h.store.Delete(sessionKey); err != nil {
		return ErrorResponse(c, "failed to get session", err)
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
