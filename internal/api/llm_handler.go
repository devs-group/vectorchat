package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/constants"
	"github.com/yourusername/vectorchat/pkg/models"
	stripe_sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

// LLMHandler serves model discovery endpoints.
type LLMHandler struct {
	AuthMiddleware *middleware.AuthMiddleware
	LLMService     *services.LLMService
	Subscriptions  *stripe_sub.Service
}

// NewLLMHandler builds a handler for LLM-related routes.
func NewLLMHandler(auth *middleware.AuthMiddleware, llmService *services.LLMService, subs *stripe_sub.Service) *LLMHandler {
	return &LLMHandler{AuthMiddleware: auth, LLMService: llmService, Subscriptions: subs}
}

// RegisterRoutes wires LLM endpoints.
func (h *LLMHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/llm", h.AuthMiddleware.RequireAuth)
	group.Get("/models", h.GET_ListModels)
}

// @Summary List available LLM models
// @Description Returns models exposed by the LLM proxy, filtered by subscription plan
// @Tags llm
// @Produce json
// @Success 200 {object} models.LLMModelsResponse
// @Failure 401 {object} models.APIResponse
// @Failure 502 {object} models.APIResponse
// @Security BearerAuth
// @Router /llm/models [get]
// GET_ListModels returns available LLM models filtered by the user's plan.
func (h *LLMHandler) GET_ListModels(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*db.User)
	if !ok || user == nil {
		return ErrorResponse(c, "User not authenticated", nil, http.StatusUnauthorized)
	}

	allowAdvanced := false
	if h.Subscriptions != nil {
		plan, _, err := h.Subscriptions.GetUserPlan(c.Context(), &user.ID, user.Email)
		if err != nil {
			return ErrorResponse(c, "Failed to retrieve subscription", err, http.StatusInternalServerError)
		}
		allowAdvanced = planAllowsAdvanced(plan)
	}

	modelsList, err := h.LLMService.ListModels(c.Context())
	if err != nil && len(modelsList) == 0 {
		return ErrorResponse(c, "Failed to fetch models", err, http.StatusBadGateway)
	}

	filtered := h.LLMService.FilterByPlan(modelsList, allowAdvanced)
	// If filtering removed everything, fall back to the unfiltered list to avoid empty dropdowns.
	if len(filtered) == 0 {
		filtered = modelsList
	}

	return c.JSON(models.LLMModelsResponse{Models: filtered})
}

func planAllowsAdvanced(plan *stripe_sub.Plan) bool {
	if plan == nil || plan.PlanDefinition == nil {
		return false
	}

	raw, ok := plan.PlanDefinition["features"].(map[string]any)
	if !ok {
		return false
	}

	feature, ok := raw[constants.LimitAdvancedModels]
	if !ok {
		return false
	}

	switch v := feature.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	default:
		return false
	}
}
