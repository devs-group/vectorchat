package api

import (
	"net/http"

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
// @Param apiKey body services.APIKeyCreateRequest true "API Key Details"
// @Success 200 {object} APIKeyResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [post]
func (h *APIKeyHandler) POST_GenerateAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*services.User)

	// Parse request body
	var req services.APIKeyCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	// Parse and validate request using service
	name, expiresAt, err := h.apiKeyService.ParseAPIKeyRequest(&req)
	if err != nil {
		return ErrorResponse(c, "Invalid request parameters", err, http.StatusBadRequest)
	}

	apiKeyResponse, plainTextKey, err := h.apiKeyService.CreateAPIKey(c.Context(), user.ID, name, expiresAt)
	if err != nil {
		return ErrorResponse(c, "failed to create API key", err)
	}

	nameStr := ""
	if apiKeyResponse.Name != nil {
		nameStr = *apiKeyResponse.Name
	}

	return c.JSON(APIKeyResponse{
		APIKey: APIKey{
			ID:        apiKeyResponse.ID,
			UserID:    apiKeyResponse.UserID,
			Key:       plainTextKey,
			Name:      nameStr,
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

	// Parse pagination parameters using service
	page, limit, offset := h.apiKeyService.ParsePaginationParams(c.Query("page"), c.Query("limit"))

	// Get paginated API keys with formatted response
	response, err := h.apiKeyService.GetAPIKeysWithPagination(c.Context(), user.ID, page, limit, offset)
	if err != nil {
		return ErrorResponse(c, "failed to get API keys", err)
	}

	// Convert to API response format
	var keys []APIKey
	for _, k := range response.APIKeys {
		nameStr := ""
		if k.Name != nil {
			nameStr = *k.Name
		}
		keys = append(keys, APIKey{
			ID:        k.ID,
			UserID:    k.UserID,
			Name:      nameStr,
			CreatedAt: k.CreatedAt,
			ExpiresAt: k.ExpiresAt,
			RevokedAt: k.RevokedAt,
		})
	}

	return c.JSON(APIKeysResponse{
		APIKeys: keys,
		Pagination: PaginationMetadata{
			Page:       response.Pagination.Page,
			Limit:      response.Pagination.Limit,
			Total:      response.Pagination.Total,
			TotalPages: response.Pagination.TotalPages,
			HasNext:    response.Pagination.HasNext,
			HasPrev:    response.Pagination.HasPrev,
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
	if err := h.apiKeyService.RevokeAPIKey(c.Context(), id, user.ID); err != nil {
		return ErrorResponse(c, "failed to revoke API key", err, http.StatusBadRequest)
	}

	return c.JSON(MessageResponse{
		Message: "API key revoked successfully",
	})
}
