package api

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/store"
)

// APIKeyHandler handles api keys
type APIKeyHandler struct {
	userStore      *store.UserStore
	authMiddleware *middleware.AuthMiddleware
	apiKeyService  *services.APIKeyService
}

// NewAPIKeyHandler creates a new OAuth handler with validation
func NewAPIKeyHandler(
	userStore *store.UserStore,
	authMiddleware *middleware.AuthMiddleware,
	apiKeyService *services.APIKeyService,
) *APIKeyHandler {
	return &APIKeyHandler{
		userStore:      userStore,
		authMiddleware: authMiddleware,
		apiKeyService:  apiKeyService,
	}
}

// RegisterRoutes registers the OAuth routes
func (h *APIKeyHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Post("/apikey", h.authMiddleware.RequireAuth, h.POST_GenerateAPIKey)
	auth.Get("/apikey", h.authMiddleware.RequireAuth, h.GET_ListAPIKeys)
	auth.Delete("/apikey/:id", h.authMiddleware.RequireAuth, h.DELETE_RevokeAPIKey)
}

// @Summary Generate API key
// @Description Generates a new API key for the authenticated user
// @Tags apiKey
// @Accept json
// @Produce json
// @Param apiKey body APIKeyRequest true "API Key Details"
// @Success 200 {object} APIKeyResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [post]
func (h *APIKeyHandler) POST_GenerateAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*store.User)

	plainTextKey, hashedKey, err := h.apiKeyService.CreateNewAPIKey()
	if err != nil {
		return ErrorResponse(c, "failed to generate new api key", err)
	}

	// Parse request body
	var req APIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	apiKey := &store.APIKey{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Name:      &req.Name,
		Key:       string(hashedKey),
		CreatedAt: time.Now(),
	}

	// Set expiration if provided
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return ErrorResponse(c, "Invalid expiration date format", err, http.StatusBadRequest)
		}
		apiKey.ExpiresAt = &expiresAt
	}
	// If no expiration provided, ExpiresAt remains nil (no expiration)

	if err := h.userStore.CreateAPIKey(c.Context(), apiKey); err != nil {
		return ErrorResponse(c, "failed to save API key", err)
	}

	return c.JSON(APIKeyResponse{
		APIKey: APIKey{
			ID:        apiKey.ID,
			UserID:    apiKey.UserID,
			Key:       plainTextKey,
			Name:      *apiKey.Name,
			CreatedAt: apiKey.CreatedAt,
			ExpiresAt: apiKey.ExpiresAt,
		},
	})
}

// @Summary List API keys
// @Description Lists API keys for the authenticated user with pagination support
// @Tags apiKey
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} APIKeysResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [get]
func (h *APIKeyHandler) GET_ListAPIKeys(c *fiber.Ctx) error {
	user := c.Locals("user").(*store.User)

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get paginated API keys
	apiKeys, total, err := h.userStore.GetAPIKeysWithPagination(c.Context(), user.ID, offset, limit)
	if err != nil {
		return ErrorResponse(c, "failed to get API keys", err)
	}

	// Convert to response format
	var keys []APIKey
	for _, k := range apiKeys {
		keys = append(keys, APIKey{
			ID:        k.ID,
			UserID:    k.UserID,
			CreatedAt: k.CreatedAt,
			ExpiresAt: k.ExpiresAt,
			RevokedAt: k.RevokedAt,
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(APIKeysResponse{
		APIKeys: keys,
		Pagination: PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	})
}

// @Summary Revoke API key
// @Description Revokes an API key for the authenticated user
// @Tags apiKey
// @Accept json
// @Produce json
// @Param id path string true "API key ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey/{id} [delete]
func (h *APIKeyHandler) DELETE_RevokeAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*store.User)

	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, "API key is required", nil, http.StatusBadRequest)
	}

	if err := h.userStore.RevokeAPIKey(c.Context(), id, user.ID); err != nil {
		return ErrorResponse(c, "failed to revoke API key", err, http.StatusBadRequest)
	}

	return c.JSON(MessageResponse{
		Message: "API key revoked successfully",
	})
}
