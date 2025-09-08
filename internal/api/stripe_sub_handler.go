package api

import (
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/middleware"
	stripe_sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

type StripeSubHandler struct {
	AuthMiddleware *middleware.AuthMiddleware
	Client         stripe_sub.Client
}

func NewStripeSubHandler(auth *middleware.AuthMiddleware, client stripe_sub.Client) *StripeSubHandler {
	return &StripeSubHandler{AuthMiddleware: auth, Client: client}
}

func (h *StripeSubHandler) RegisterRoutes(app *fiber.App) {
	// Public webhook (Stripe calls it)
	if wh, err := h.Client.WebhookHandler(stripe_sub.Hooks{}); err == nil {
		app.Post("/stripe/webhook", adaptor.HTTPHandler(wh))
	}

	// Authenticated billing routes
	grp := app.Group("/billing", h.AuthMiddleware.RequireAuth)
	grp.Get("/plans", h.GET_Plans)
	grp.Post("/checkout-session", h.POST_CheckoutSession)
	grp.Post("/setup-intent", h.POST_SetupIntent)
}

func (h *StripeSubHandler) GET_Plans(c *fiber.Ctx) error {
	plans, err := h.Client.GetPlans(c.Context())
	if err != nil {
		return ErrorResponse(c, "failed to list plans", err, http.StatusInternalServerError)
	}
	return c.JSON(plans)
}

type checkoutReq struct {
	CustomerID          string            `json:"customer_id"`
	PlanKey             string            `json:"plan_key"`
	SuccessURL          string            `json:"success_url"`
	CancelURL           string            `json:"cancel_url"`
	Quantity            *int              `json:"quantity"`
	AllowPromotionCodes bool              `json:"allow_promotion_codes"`
	IdempotencyKey      *string           `json:"idempotency_key"`
	Metadata            map[string]string `json:"metadata"`
}

func (h *StripeSubHandler) POST_CheckoutSession(c *fiber.Ctx) error {
	var in checkoutReq
	if err := c.BodyParser(&in); err != nil {
		return ErrorResponse(c, "invalid json", err, http.StatusBadRequest)
	}
	if in.CustomerID == "" || in.PlanKey == "" || in.SuccessURL == "" || in.CancelURL == "" {
		return ErrorResponse(c, "missing required fields", nil, http.StatusBadRequest)
	}
	sid, url, err := h.Client.CreateCheckoutSessionForPlan(c.Context(), in.CustomerID, in.PlanKey, in.SuccessURL, in.CancelURL, stripe_sub.CheckoutSessionOpts{
		Quantity: in.Quantity, AllowPromotionCodes: in.AllowPromotionCodes, IdempotencyKey: in.IdempotencyKey, Metadata: in.Metadata,
	})
	if err != nil {
		return ErrorResponse(c, "failed to create checkout session", err, http.StatusBadRequest)
	}
	return c.JSON(fiber.Map{"session_id": sid, "url": url})
}

type setupIntentReq struct {
	CustomerID string `json:"customer_id"`
}

func (h *StripeSubHandler) POST_SetupIntent(c *fiber.Ctx) error {
	var in setupIntentReq
	if err := c.BodyParser(&in); err != nil {
		return ErrorResponse(c, "invalid json", err, http.StatusBadRequest)
	}
	if in.CustomerID == "" {
		return ErrorResponse(c, "missing customer_id", nil, http.StatusBadRequest)
	}
	secret, err := h.Client.CreateSetupIntent(c.Context(), in.CustomerID)
	if err != nil {
		return ErrorResponse(c, "failed to create setup intent", err, http.StatusBadRequest)
	}
	return c.JSON(fiber.Map{"client_secret": secret})
}
