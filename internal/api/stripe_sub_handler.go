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
    app.Post("/stripe/webhook", adaptor.HTTPHandlerFunc(h.Service.WebhookHandler()))

    // Public plans
    app.Get("/billing/plans", adaptor.HTTPHandlerFunc(h.Service.PlansHandler()))

    // Authenticated billing routes
    grp := app.Group("/billing", h.AuthMiddleware.RequireAuth)
    grp.Post("/checkout-session", h.POST_CheckoutSession)
    grp.Get("/subscription", h.GET_Subscription)
    grp.Post("/portal-session", h.POST_PortalSession)
}

// checkoutReq matches the current frontend payload; we map it to the service request.
type checkoutReq struct {
    CustomerID          string             `json:"customer_id"`
    PlanKey             string             `json:"plan_key"`
    SuccessURL          string             `json:"success_url"`
    CancelURL           string             `json:"cancel_url"`
    AllowPromotionCodes bool               `json:"allow_promotion_codes"`
    IdempotencyKey      *string            `json:"idempotency_key"`
    Metadata            map[string]string  `json:"metadata"`
}

func (h *StripeSubHandler) POST_CheckoutSession(c *fiber.Ctx) error {
    var in checkoutReq
    if err := c.BodyParser(&in); err != nil {
        return ErrorResponse(c, "invalid json", err, fiber.StatusBadRequest)
    }
    // Pull authenticated user to get email; CustomerID is used as external_id for reference.
    u := c.Locals("user")
    if u == nil {
        return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized)
    }
    user := u.(*db.User)
    req := sub.CheckoutRequest{
        PlanKey:    in.PlanKey,
        ExternalID: func() *string { if in.CustomerID == "" { return nil }; v := in.CustomerID; return &v }(),
        Email:      user.Email,
        SuccessURL: in.SuccessURL,
        CancelURL:  in.CancelURL,
    }
    id, url, err := h.Service.CreateCheckoutSession(c.Context(), req)
    if err != nil {
        return ErrorResponse(c, "failed to create checkout session", err, fiber.StatusBadRequest)
    }
    return c.JSON(fiber.Map{"session_id": id, "url": url})
}

// GET_Subscription returns the user's latest subscription status.
func (h *StripeSubHandler) GET_Subscription(c *fiber.Ctx) error {
    u := c.Locals("user")
    if u == nil { return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized) }
    user := u.(*db.User)
    // Use user.ID as external reference and their email for fallback.
    ext := user.ID
    sub, err := h.Service.GetUserSubscription(c.Context(), &ext, user.Email)
    if err != nil { return ErrorResponse(c, "failed to fetch subscription", err, fiber.StatusInternalServerError) }
    if sub == nil { return c.JSON(fiber.Map{"subscription": nil}) }
    return c.JSON(fiber.Map{"subscription": sub})
}

// POST_PortalSession creates a Stripe Customer Portal session for the current user and returns the URL.
func (h *StripeSubHandler) POST_PortalSession(c *fiber.Ctx) error {
    u := c.Locals("user")
    if u == nil { return ErrorResponse(c, "missing user session", nil, fiber.StatusUnauthorized) }
    user := u.(*db.User)
    var body struct{ ReturnURL string `json:"return_url"` }
    if err := c.BodyParser(&body); err != nil {
        return ErrorResponse(c, "invalid json", err, fiber.StatusBadRequest)
    }
    if body.ReturnURL == "" {
        return ErrorResponse(c, "missing return_url", nil, fiber.StatusBadRequest)
    }
    ext := user.ID
    url, err := h.Service.CreatePortalSession(c.Context(), sub.PortalRequest{ExternalID: &ext, Email: user.Email, ReturnURL: body.ReturnURL})
    if err != nil {
        return ErrorResponse(c, "failed to create portal session", err, fiber.StatusBadRequest)
    }
    return c.JSON(fiber.Map{"url": url})
}
