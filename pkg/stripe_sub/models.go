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

// Feature returns a raw feature value from PlanDefinition.
func (p *Plan) Feature(key string) (any, bool) {
    if p == nil || p.PlanDefinition == nil {
        return nil, false
    }
    v, ok := p.PlanDefinition[key]
    return v, ok
}

// FeatureString returns a string feature if present.
func (p *Plan) FeatureString(key string) (string, bool) {
    v, ok := p.Feature(key)
    if !ok || v == nil { return "", false }
    s, ok := v.(string)
    return s, ok
}

// FeatureBool returns a bool feature if present.
func (p *Plan) FeatureBool(key string) (bool, bool) {
    v, ok := p.Feature(key)
    if !ok || v == nil { return false, false }
    b, ok := v.(bool)
    return b, ok
}

// FeatureFloat returns a float64 feature if present.
func (p *Plan) FeatureFloat(key string) (float64, bool) {
    v, ok := p.Feature(key)
    if !ok || v == nil { return 0, false }
    switch n := v.(type) {
    case float64:
        return n, true
    case float32:
        return float64(n), true
    case int:
        return float64(n), true
    case int64:
        return float64(n), true
    case int32:
        return float64(n), true
    case json.Number:
        f, err := n.Float64()
        if err != nil { return 0, false }
        return f, true
    default:
        return 0, false
    }
}

// FeatureInt returns an int64 feature with best-effort conversion from JSON numbers.
func (p *Plan) FeatureInt(key string) (int64, bool) {
    v, ok := p.Feature(key)
    if !ok || v == nil { return 0, false }
    switch n := v.(type) {
    case int64:
        return n, true
    case int:
        return int64(n), true
    case int32:
        return int64(n), true
    case float64:
        return int64(n), true
    case float32:
        return int64(n), true
    case json.Number:
        i, err := n.Int64()
        if err != nil { return 0, false }
        return i, true
    default:
        return 0, false
    }
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
