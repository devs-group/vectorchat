stripe_sub — Minimal Stripe subscriptions for Go
================================================

Framework‑agnostic HTTP handlers and a tiny service layer to sell subscriptions with Stripe Checkout and keep your Postgres in sync via webhooks. No invoices UI, no customer portal UIs — just the essentials.

Why use it
- Small surface area: initialize, define plans, mount 3–4 handlers.
- Framework‑agnostic: plain `net/http`, adapt anywhere.
- Self‑contained migrations: creates and updates its own tables and indexes.
- Practical defaults: idempotent Checkout/Portal calls, email normalization, simple “active” policy.

Install
- `go get github.com/yourusername/vectorchat/pkg/stripe_sub`

Quickstart
```go
ctx := context.Background()
sqldb, _ := sql.Open("postgres", "postgres://user:pass@localhost:5432/db?sslmode=disable")
db := sqlx.NewDb(sqldb, "postgres")

svc, _ := stripe_sub.New(ctx, stripe_sub.Config{
    DB:            db,
    StripeAPIKey:  os.Getenv("STRIPE_API_KEY"),
    WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
})

// Plans (idempotent at startup)
_, _ = svc.UpsertPlan(ctx, stripe_sub.PlanParams{
    Key: "starter", DisplayName: "Starter", Active: true,
    BillingInterval: "month", AmountCents: 900, Currency: "usd",
    PlanDefinition: map[string]any{"seats": 1},
})

// Routes (examples)
mux := http.NewServeMux()
mux.HandleFunc("/billing/plans", svc.PlansHandler())              // GET
mux.HandleFunc("/stripe/webhook", svc.WebhookHandler())            // POST
mux.HandleFunc("/billing/checkout-session", func(w http.ResponseWriter, r *http.Request){
    user := mustUser(r); ext := user.ID
    svc.CheckoutAuthedHandlerFor(user.Email, &ext).ServeHTTP(w, r)
})
mux.HandleFunc("/billing/subscription", func(w http.ResponseWriter, r *http.Request){
    user := mustUser(r); ext := user.ID
    svc.SubscriptionHandlerFor(user.Email, &ext, true, true).ServeHTTP(w, r)
})
mux.HandleFunc("/billing/portal-session", func(w http.ResponseWriter, r *http.Request){
    user := mustUser(r); ext := user.ID
    svc.PortalAuthedHandlerFor(user.Email, &ext).ServeHTTP(w, r)
})
```

HTTP contracts
- Checkout (POST `/billing/checkout-session`): body contains `plan_key`, `success_url`, `cancel_url`. Identity comes from your auth (email and optional external ID). Response `{ "id", "url" }` for Stripe redirect.
- Subscription (GET `/billing/subscription`): returns the user’s latest record; when `EnsurePlanKey` is enabled the plan key is inferred if missing.
- Plans (GET `/billing/plans`): active plans for your UI.
- Webhook (POST `/stripe/webhook`): verifies signature and syncs subscriptions.

Stripe webhooks
Enable these events for your endpoint:
- `checkout.session.completed`
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`

Enforcing access in your app
- Fetch plan + subscription: `plan, sub, err := svc.GetUserPlan(ctx, &user.ID, user.Email)`
- Active check: `stripe_sub.IsSubscriptionActive(sub, time.Now())`
- Feature flags/quotas from `plan.PlanDefinition` via `Feature*` helpers.

Data model & migrations
- Postgres tables: `stripe_sub_pkg_customers`, `stripe_sub_pkg_plans`, `stripe_sub_pkg_subscriptions` (+ indexes).
- Migrations run automatically when calling `stripe_sub.New`.

Configuration
- `STRIPE_API_KEY`, `STRIPE_WEBHOOK_SECRET` (or pass in `Config`).
- Prices are defined inline from your plan records; no Price IDs required.

Compatibility
- Plain `net/http`; adapt to Fiber, Chi, Gin, etc.
- Works with Stripe API v76 client (see `go.mod`).
