package stripe_sub

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONB is a minimal helper type to scan/store JSONB values.
type JSONB map[string]any

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return []byte("null"), nil
	}
	b, err := json.Marshal(map[string]any(j))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*j = nil
			return nil
		}
		var m map[string]any
		if err := json.Unmarshal(v, &m); err != nil {
			return err
		}
		*j = m
		return nil
	case string:
		if v == "" {
			*j = nil
			return nil
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return err
		}
		*j = m
		return nil
	default:
		return fmt.Errorf("unsupported JSONB scan type %T", value)
	}
}

// Plan represents a subscription plan stored in DB and exposed via API.
type Plan struct {
	ID              string    `db:"id" json:"id"`
	Key             string    `db:"key" json:"key"`
	DisplayName     string    `db:"display_name" json:"display_name"`
	Active          bool      `db:"active" json:"active"`
	BillingInterval string    `db:"billing_interval" json:"billing_interval"`
	AmountCents     int64     `db:"amount_cents" json:"amount_cents"`
	Currency        string    `db:"currency" json:"currency"`
	Metadata        JSONB     `db:"metadata" json:"metadata"`
	PlanDefinition  JSONB     `db:"plan_definition" json:"plan_definition"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// Customer maps your app user to a Stripe customer
type Customer struct {
	ID               string    `db:"id" json:"id"`
	ExternalID       *string   `db:"ext_id" json:"ext_id"`
	StripeCustomerID string    `db:"stripe_customer_id" json:"stripe_customer_id"`
	Email            string    `db:"email" json:"email"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// Subscription reflects the current state stored for a Stripe subscription.
type Subscription struct {
	ID                 string     `db:"id" json:"id"`
	CustomerID         string     `db:"customer_id" json:"customer_id"`
	StripeSubscription string     `db:"stripe_subscription_id" json:"stripe_subscription_id"`
	Status             string     `db:"status" json:"status"`
	CurrentPeriodStart *time.Time `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `db:"current_period_end" json:"current_period_end"`
	CancelAtPeriodEnd  bool       `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	Metadata           JSONB      `db:"metadata" json:"metadata"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}
