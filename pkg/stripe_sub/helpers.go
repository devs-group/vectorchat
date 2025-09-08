package stripe_sub

import "time"

// nullBytes converts a non-empty JSON byte slice to *string for jsonb columns.
func nullBytes(b []byte) *string { if len(b) == 0 { return nil }; s := string(b); return &s }

// IsActiveSubscription returns true if the subscription status is active or trialing and not past current period end.
func IsActiveSubscription(s Subscription, now time.Time) bool {
    if s.Status != "active" && s.Status != "trialing" { return false }
    if s.CurrentPeriodEnd != nil && now.After(*s.CurrentPeriodEnd) { return false }
    return true
}

