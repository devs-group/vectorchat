package api

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
)

// APIKeyHandler handles api keys
type APIKeyHandler struct {
	authService    *services.AuthService
	authMiddleware *middleware.AuthMiddleware
	apiKeyService  *services.APIKeyService
}

// NewAPIKeyHandler creates a new OAuth handler with validation
func NewAPIKeyHandler(
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
	apiKeyService *services.APIKeyService,
) *APIKeyHandler {
	return &APIKeyHandler{
		authService:    authService,
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
	user := c.Locals("user").(*services.User)

	// Parse request body
	var req APIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	// Parse expiration date if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return ErrorResponse(c, "Invalid expiration date format", err, http.StatusBadRequest)
		}
		expiresAt = &parsedTime
	}

	// Create the API key using the service
	apiKeyResponse, plainTextKey, err := h.apiKeyService.CreateAPIKey(c.Context(), user.ID, req.Name, expiresAt)
	if err != nil {
		return ErrorResponse(c, "failed to create API key", err)
	}

	name := ""
	if apiKeyResponse.Name != nil {
		name = *apiKeyResponse.Name
	}

	return c.JSON(APIKeyResponse{
		APIKey: APIKey{
			ID:        apiKeyResponse.ID,
			UserID:    apiKeyResponse.UserID,
			Key:       plainTextKey,
			Name:      name,
			CreatedAt: apiKeyResponse.CreatedAt,
			ExpiresAt: apiKeyResponse.ExpiresAt,
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
	user := c.Locals("user").(*services.User)

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
	paginatedResponse, err := h.apiKeyService.GetAPIKeysWithPagination(c.Context(), user.ID, offset, limit)
	if err != nil {
		return ErrorResponse(c, "failed to get API keys", err)
	}

	// Convert to response format
	var keys []APIKey
	apiKeyResponses := paginatedResponse.Data.([]*services.APIKeyResponse)
	for _, k := range apiKeyResponses {
		name := ""
		if k.Name != nil {
			name = *k.Name
		}
		keys = append(keys, APIKey{
			ID:        k.ID,
			UserID:    k.UserID,
			Name:      name,
			CreatedAt: k.CreatedAt,
			ExpiresAt: k.ExpiresAt,
			RevokedAt: k.RevokedAt,
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(paginatedResponse.Total) / float64(limit)))
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(APIKeysResponse{
		APIKeys: keys,
		Pagination: PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      paginatedResponse.Total,
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
	user := c.Locals("user").(*services.User)

	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, "API key is required", nil, http.StatusBadRequest)
	}

	if err := h.apiKeyService.RevokeAPIKey(c.Context(), id, user.ID); err != nil {
		return ErrorResponse(c, "failed to revoke API key", err, http.StatusBadRequest)
	}

	return c.JSON(MessageResponse{
		Message: "API key revoked successfully",
	})
}
