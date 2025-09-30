package api

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

// StripeSubHandler wires the subscription service into Fiber routes.
type StripeSubHandler struct {
	AuthMiddleware *middleware.AuthMiddleware
	Service        *sub.Service
}

func NewStripeSubHandler(auth *middleware.AuthMiddleware, svc *sub.Service) *StripeSubHandler {
	return &StripeSubHandler{AuthMiddleware: auth, Service: svc}
}

func (h *StripeSubHandler) RegisterRoutes(app *fiber.App) {
	// Public webhook (Stripe calls it)
	app.Post("/public/stripe/webhook", adaptor.HTTPHandlerFunc(h.Service.WebhookHandler()))

	// Public plans
	app.Get("/public/billing/plans", adaptor.HTTPHandlerFunc(h.Service.PlansHandler()))

	// Authenticated billing routes
	grp := app.Group("/billing", h.AuthMiddleware.RequireAuth)

	grp.Post("/checkout-session", func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil {
			return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized)
		}
		user := u.(*db.User)
		ext := user.ID
		return adaptor.HTTPHandlerFunc(h.Service.CheckoutAuthedHandlerFor(user.Email, &ext))(c)
	})

	grp.Get("/subscription", func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil {
			return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized)
		}
		user := u.(*db.User)
		ext := user.ID
		return adaptor.HTTPHandlerFunc(h.Service.SubscriptionHandlerFor(user.Email, &ext, true, true))(c)
	})

	grp.Post("/portal-session", func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil {
			return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized)
		}
		user := u.(*db.User)
		ext := user.ID
		return adaptor.HTTPHandlerFunc(h.Service.PortalAuthedHandlerFor(user.Email, &ext))(c)
	})

	grp.Get("/limits", func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil {
			return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized)
		}
		user := u.(*db.User)
		ext := user.ID
		return adaptor.HTTPHandlerFunc(h.Service.UserLimitsHandler(user.Email, &ext))(c)
	})
}
