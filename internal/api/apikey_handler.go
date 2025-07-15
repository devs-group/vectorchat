package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

// APIKeyHandler handles api keys
type APIKeyHandler struct {
	authService    *services.AuthService
	authMiddleware *middleware.AuthMiddleware
	apiKeyService  *services.APIKeyService
	commonService  *services.CommonService
}

// NewAPIKeyHandler creates a new OAuth handler with validation
func NewAPIKeyHandler(
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
	apiKeyService *services.APIKeyService,
	commonService *services.CommonService,
) *APIKeyHandler {
	return &APIKeyHandler{
		authService:    authService,
		authMiddleware: authMiddleware,
		apiKeyService:  apiKeyService,
		commonService:  commonService,
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
// @Param apiKey body models.APIKeyCreateRequest true "API Key Details"
// @Success 200 {object} models.APIKeyCreateResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [post]
func (h *APIKeyHandler) POST_GenerateAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)

	// Parse request body
	var req models.APIKeyCreateRequest
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

	return c.JSON(models.APIKeyCreateResponse{
		APIKey: models.APIKeyResponse{
			ID:        apiKeyResponse.ID,
			UserID:    apiKeyResponse.UserID,
			Name:      apiKeyResponse.Name,
			CreatedAt: apiKeyResponse.CreatedAt,
			ExpiresAt: apiKeyResponse.ExpiresAt,
		},
		PlainKey: plainTextKey,
		Message:  "API key created successfully. Save this key as it won't be shown again.",
	})
}

// @Summary List API keys
// @Description Lists API keys for the authenticated user with pagination support
// @Tags apiKey
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} models.APIKeysListResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey [get]
func (h *APIKeyHandler) GET_ListAPIKeys(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)

	// Parse pagination parameters using service
	page, limit, offset := h.commonService.ParsePaginationParams(c.Query("page"), c.Query("limit"))

	// Get paginated API keys with formatted response
	response, err := h.apiKeyService.GetAPIKeysWithPagination(c.Context(), user.ID, page, limit, offset)
	if err != nil {
		return ErrorResponse(c, "failed to get API keys", err)
	}

	// Convert to API response format
	var keys []*models.APIKeyResponse
	for _, k := range response.APIKeys {
		keys = append(keys, &models.APIKeyResponse{
			ID:        k.ID,
			UserID:    k.UserID,
			Name:      k.Name,
			CreatedAt: k.CreatedAt,
			ExpiresAt: k.ExpiresAt,
			RevokedAt: k.RevokedAt,
		})
	}

	return c.JSON(models.APIKeysListResponse{
		APIKeys: keys,
		Pagination: &models.PaginationMetadata{
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
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /auth/apikey/{id} [delete]
func (h *APIKeyHandler) DELETE_RevokeAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*db.User)

	id := c.Params("id")
	if err := h.apiKeyService.RevokeAPIKey(c.Context(), id, user.ID); err != nil {
		return ErrorResponse(c, "failed to revoke API key", err, http.StatusBadRequest)
	}

	return c.JSON(models.MessageResponse{
		Message: "API key revoked successfully",
	})
}
