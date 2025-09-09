# stripe_sub – Simple Stripe subscriptions for Go

Goals:

- Minimal API: init, create plans, use 3 handlers
- Auto‑migrations of its own tables
- Checkout Sessions only (no portal, no invoices UI)

## Quickstart

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"

    _ "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
    sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

func main() {
    ctx := context.Background()
    sqldb, _ := sql.Open("postgres", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
    db := sqlx.NewDb(sqldb, "postgres")

    svc, err := sub.New(ctx, sub.Config{
        DB:               db,
        StripeAPIKey:     "sk_live_or_test",
        StripeAPIVersion: "2023-10-16", // must match your Stripe webhook endpoint version
        WebhookSecret:    "whsec_...",
    })
    if err != nil { log.Fatal(err) }

    // Define plans at startup (idempotent)
    _, _ = svc.UpsertPlan(ctx, sub.PlanParams{
        Key:             "starter",
        DisplayName:     "Starter",
        Active:          true,
        BillingInterval: "month", // day|week|month|year
        AmountCents:     900,
        Currency:        "usd",
        PlanDefinition:  map[string]any{"seats": 1},
    })

    mux := http.NewServeMux()
    // Public
    mux.HandleFunc("/billing/plans", svc.PlansHandler())
    mux.HandleFunc("/stripe/webhook", svc.WebhookHandler())

    // Authenticated (example: wrap with your auth that sets user)
    mux.HandleFunc("/billing/checkout-session", func(w http.ResponseWriter, r *http.Request) {
        user := mustUser(r) // your auth
        ext := user.ID
        svc.CheckoutAuthedHandlerFor(user.Email, &ext).ServeHTTP(w, r)
    })
    mux.HandleFunc("/billing/subscription", func(w http.ResponseWriter, r *http.Request) {
        user := mustUser(r)
        ext := user.ID
        svc.SubscriptionHandlerFor(user.Email, &ext, true, true).ServeHTTP(w, r)
    })
    mux.HandleFunc("/billing/portal-session", func(w http.ResponseWriter, r *http.Request) {
        user := mustUser(r)
        ext := user.ID
        svc.PortalAuthedHandlerFor(user.Email, &ext).ServeHTTP(w, r)
    })

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### Checkout request body

```json
{
  "plan_key": "starter",
  "external_id": "user-123", // optional
  "email": "user@example.com",
  "success_url": "https://app.example.com/billing/success?session_id={CHECKOUT_SESSION_ID}",
  "cancel_url": "https://app.example.com/billing/cancel"
}
```

Response:

```json
{ "id": "cs_test_...", "url": "https://checkout.stripe.com/c/pay/cs_test_..." }
```

### Webhooks

Configure the endpoint to `/api/stripe/webhook` and send at least:

- `checkout.session.completed`
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`

The handler verifies the signature and keeps your subscriptions table in sync.

Stripe API version must match: set `STRIPE_API_VERSION` (or pass in Config) to the same version configured on your Stripe webhook endpoint. The server logs a short hint on startup.

Environment variables typically used:

- `STRIPE_API_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_API_VERSION` (e.g., `2023-10-16`)

### Using with Fiber (example)

```go
app := fiber.New()
// Public
app.Get("/billing/plans", adaptor.HTTPHandlerFunc(svc.PlansHandler()))
app.Post("/stripe/webhook", adaptor.HTTPHandlerFunc(svc.WebhookHandler()))

// Authenticated routes
app.Post("/billing/checkout-session", func(c *fiber.Ctx) error {
    user := mustFiberUser(c)
    ext := user.ID
    return adaptor.HTTPHandlerFunc(svc.CheckoutAuthedHandlerFor(user.Email, &ext))(c)
})
app.Get("/billing/subscription", func(c *fiber.Ctx) error {
    user := mustFiberUser(c)
    ext := user.ID
    return adaptor.HTTPHandlerFunc(svc.SubscriptionHandlerFor(user.Email, &ext, true, true))(c)
})
app.Post("/billing/portal-session", func(c *fiber.Ctx) error {
    user := mustFiberUser(c)
    ext := user.ID
    return adaptor.HTTPHandlerFunc(svc.PortalAuthedHandlerFor(user.Email, &ext))(c)
})
```

Handlers are plain `net/http` and can be adapted to any framework.

## Enforcing Plan Features in Your Backend

Once webhooks are syncing, you can gate features using the plan attached to the user's latest subscription. The helpers below are available:

- `svc.GetUserPlan(ctx, externalID *string, email string) (*Plan, *Subscription, error)`
- `stripe_sub.IsSubscriptionActive(sub *Subscription, now time.Time) bool`
- `plan.Feature…(key)` accessors: `Feature`, `FeatureBool`, `FeatureInt`, `FeatureFloat`, `FeatureString`

### Example: Simple check inside a handler

```go
func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    user := h.mustUserFromContext(ctx) // your auth

    plan, sub, err := h.subSvc.GetUserPlan(ctx, &user.ID, user.Email)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, err)
        return
    }
    if plan == nil || !stripe_sub.IsSubscriptionActive(sub, time.Now()) {
        writeErrMsg(w, http.StatusPaymentRequired, "an active subscription is required")
        return
    }

    // Feature flags / quotas from plan definition
    // e.g., seats limit, boolean toggle, numeric caps
    seats, _ := plan.FeatureInt("seats")           // int64
    canExport, _ := plan.FeatureBool("export_pdf") // bool

    _ = seats
    _ = canExport
    // ... proceed with business logic
}
```

### Example: HTTP middleware (net/http)

```go
func RequirePlan(subSvc *stripe_sub.Service, feature string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            user := mustUserFromContext(ctx)

            plan, sub, err := subSvc.GetUserPlan(ctx, &user.ID, user.Email)
            if err != nil {
                writeErr(w, http.StatusInternalServerError, err)
                return
            }
            if plan == nil || !stripe_sub.IsSubscriptionActive(sub, time.Now()) {
                writeErrMsg(w, http.StatusPaymentRequired, "active subscription required")
                return
            }
            // If a specific feature key is required, enforce it.
            if feature != "" {
                ok := false
                if v, exists := plan.Feature(feature); exists {
                    // boolean-style features; treat truthy as enabled
                    if b, isBool := v.(bool); isBool && b {
                        ok = true
                    }
                }
                if !ok {
                    writeErrMsg(w, http.StatusForbidden, "feature not included in plan")
                    return
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Example: Fiber middleware (github.com/gofiber/fiber/v2)

```go
// RequirePlan returns a Fiber middleware that ensures a user has an active
// subscription and (optionally) a specific feature enabled by plan.
func RequirePlan(subSvc *stripe_sub.Service, feature string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Replace with your own user retrieval
        user, ok := c.Locals("user").(*User)
        if !ok || user == nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
        }

        plan, sub, err := subSvc.GetUserPlan(c.Context(), &user.ID, user.Email)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }
        if plan == nil || !stripe_sub.IsSubscriptionActive(sub, time.Now()) {
            return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "active subscription required"})
        }
        if feature != "" {
            ok := false
            if v, exists := plan.Feature(feature); exists {
                if b, isBool := v.(bool); isBool && b {
                    ok = true
                }
            }
            if !ok {
                return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "feature not included in plan"})
            }
        }
        return c.Next()
    }
}

// Usage
app := fiber.New()
app.Post("/api/export", RequirePlan(subSvc, "export_pdf"), exportHandler)
```

### Interpreting subscription state

- Active check uses `IsSubscriptionActive(sub, now)`; by default it treats `active`, `trialing`, and `past_due` as active.
- You can tighten or loosen the policy centrally by changing `IsSubscriptionActive` in `service.go`.

### Notes and best practices

- `GetUserPlan` relies on `plan_key` stored in `subscriptions.metadata` (set at checkout). If missing, it returns `plan=nil`—treat as no paid plan.
- For a free tier: if `plan == nil`, fall back to your free limits; otherwise, use `plan.PlanDefinition`.
- Define plan features at startup with `UpsertPlan` and a `PlanDefinition` map, for example:

```go
_, _ = svc.UpsertPlan(ctx, sub.PlanParams{
    Key:            "pro",
    DisplayName:    "Pro",
    Active:         true,
    BillingInterval:"month",
    AmountCents:    1900,
    Currency:       "usd",
    PlanDefinition: map[string]any{
        "seats": 5,
        "export_pdf": true,
        "max_projects": 50,
    },
})
```
