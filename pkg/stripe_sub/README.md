stripeSub — Minimal Stripe Subscriptions for Go

What it is
- Small Go package to manage Stripe subscriptions with Postgres + sqlx.
- Zero global state. Simple, production‑oriented defaults. No caching, no DB idempotency.
- Use Stripe Checkout + webhooks to create/update subscriptions automatically.

Key features
- Plans in DB with JSONB definition (Stripe price ID, features, limits).
- EnsurePlans on startup to declare/seed plans programmatically.
- Hosted Checkout: create a session, let webhook write the DB.
- SetupIntent: get a client secret for Payment Element to save/update PMs.
- Query/update: GetPlans, GetPlan, GetSubscription, UpdateSubscription, CancelSubscription.

Install
- go get github.com/yourusername/vectorchat/pkg/stripe_sub

Requirements
- Postgres. The migration enables pgcrypto for UUIDs.
- Env vars:
  - DATABASE_URL (e.g., postgres://user:pass@host:5432/db?sslmode=disable)
  - STRIPE_API_KEY
  - STRIPE_WEBHOOK_SECRET (for the webhook handler)

Tables (created by Migrate)
- stripe_sub_pkg_customers
- stripe_sub_pkg_plans
- stripe_sub_pkg_subscriptions
- stripe_sub_pkg_schema_migrations

Quick start (server)
1) Create client and migrate

  c, _ := stripe_sub.NewClient(
    stripe_sub.WithDB(db),
    stripe_sub.WithWebhookSecret(os.Getenv("STRIPE_WEBHOOK_SECRET")),
  )
  _ = c.Migrate(ctx)

2) Declare plans on startup (idempotent)

  _ = c.EnsurePlans(ctx, []stripe_sub.PlanSpec{{
    Key: "pro", DisplayName: "Pro", Active: true, BillingInterval: "month", AmountCents: 2000, Currency: "usd",
    Definition: stripe_sub.PlanDefinition{StripePriceID: "price_123", Features: []string{"feature1"}, Limits: map[string]int{"seats": 5}},
  }})

3) Mount endpoints (stdlib http)

  http.Handle("/api/stripe/checkout-session", stripe_sub.CheckoutSessionHTTPHandler(c, logger))
  http.Handle("/api/stripe/setup-intent",     stripe_sub.SetupIntentHTTPHandler(c, logger))
  http.Handle("/api/plans",                    stripe_sub.PlansHTTPHandler(c, logger))

  wh, _ := c.WebhookHandler(stripe_sub.Hooks{})
  http.Handle("/stripe/webhook", wh)

That’s it. Checkout + webhook will create/update subscriptions in your DB.

