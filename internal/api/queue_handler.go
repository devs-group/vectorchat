package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

type QueueHandler struct {
	AuthMiddleware *middleware.AuthMiddleware
	Service        *services.QueueMetricsService
}

func NewQueueHandler(auth *middleware.AuthMiddleware, svc *services.QueueMetricsService) *QueueHandler {
	return &QueueHandler{AuthMiddleware: auth, Service: svc}
}

func (h *QueueHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/queue", h.AuthMiddleware.RequireAuth)
	group.Get("/crawl/metrics", h.GET_CrawlMetrics)
}

// GET_CrawlMetrics provides basic JetStream crawl queue stats.
func (h *QueueHandler) GET_CrawlMetrics(c *fiber.Ctx) error {
	if h.Service == nil {
		return ErrorResponse(c, "Queue metrics unavailable", nil, http.StatusServiceUnavailable)
	}
	metrics, err := h.Service.GetCrawlMetrics(c.Context())
	if err != nil {
		return ErrorResponse(c, "Failed to fetch queue metrics", err)
	}
	return c.JSON(models.APIResponse{Message: "ok", Data: metrics})
}
