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
        DB:            db,
        StripeAPIKey:  "sk_live_or_test",
        WebhookSecret: "whsec_...",
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
    mux.HandleFunc("/api/plans", svc.PlansHandler())
    mux.HandleFunc("/api/checkout", svc.CheckoutHandler())
    mux.HandleFunc("/api/stripe/webhook", svc.WebhookHandler())

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

