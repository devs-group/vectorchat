package stripe_sub

import "time"

type Customer struct {
    ID             string  `db:"id" json:"id"`
    ExternalID     *string `db:"ext_id" json:"external_id,omitempty"`
    StripeCustomer string  `db:"stripe_customer_id" json:"stripe_customer_id"`
    Email          string  `db:"email" json:"email"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
    UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type Subscription struct {
    ID                   string     `db:"id" json:"id"`
    CustomerID           string     `db:"customer_id" json:"customer_id"`
    StripeSubscriptionID string     `db:"stripe_subscription_id" json:"stripe_subscription_id"`
    Status               string     `db:"status" json:"status"`
    CurrentPeriodStart   *time.Time `db:"current_period_start" json:"current_period_start,omitempty"`
    CurrentPeriodEnd     *time.Time `db:"current_period_end" json:"current_period_end,omitempty"`
    CancelAtPeriodEnd    bool       `db:"cancel_at_period_end" json:"cancel_at_period_end"`
    Metadata             []byte     `db:"metadata" json:"metadata,omitempty"`
    CreatedAt            time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

type PlanDefinition struct {
    StripePriceID string                 `json:"stripe_price_id"`
    Features      map[string]any         `json:"features,omitempty"`
    TrialDays     int                    `json:"trial_days,omitempty"`
    Tags          []string               `json:"tags,omitempty"`
}

type Plan struct {
    ID              string         `db:"id" json:"id"`
    Key             string         `db:"key" json:"key"`
    DisplayName     string         `db:"display_name" json:"display_name"`
    Active          bool           `db:"active" json:"active"`
    BillingInterval string         `db:"billing_interval" json:"billing_interval"`
    AmountCents     int64          `db:"amount_cents" json:"amount_cents"`
    Currency        string         `db:"currency" json:"currency"`
    Metadata        []byte         `db:"metadata" json:"metadata,omitempty"`
    Definition      PlanDefinition `db:"plan_definition" json:"plan_definition"`
    CreatedAt       time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

type CreateSubOpts struct {
    TrialDays       *int
    Quantity        *int
    Prorate         *bool
    CouponCode      *string
    PaymentMethodID *string
    IdempotencyKey  *string
    Metadata        map[string]string
}

type UpdateSubOpts struct {
    Prorate        *bool
    Quantity       *int
    IdempotencyKey *string
    Metadata       map[string]string
}

type CheckoutSessionOpts struct {
    Quantity            *int
    AllowPromotionCodes bool
    IdempotencyKey      *string
    Metadata            map[string]string
}
