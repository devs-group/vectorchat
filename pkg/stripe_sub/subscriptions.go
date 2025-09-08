package stripe_sub

import (
	"context"
	"encoding/json"
)

// UpdateSubscription changes the plan for an existing subscription.
func (c *service) UpdateSubscription(ctx context.Context, subID, newPlanKey string, opts UpdateSubOpts) (Subscription, error) {
	var s Subscription
	if err := c.db.GetContext(ctx, &s, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE id=$1`, subID); err != nil {
		return Subscription{}, ErrNotFound("subscription not found")
	}
	plan, err := c.GetPlan(ctx, newPlanKey)
	if err != nil {
		return Subscription{}, err
	}
	quantity := 1
	if opts.Quantity != nil && *opts.Quantity > 0 {
		quantity = *opts.Quantity
	}
	proration := prorationBehavior(opts.Prorate)
	idemKey := ""
	if opts.IdempotencyKey != nil {
		idemKey = *opts.IdempotencyKey
	}
	status, curStart, curEnd, err := c.stripe.UpdateSubscriptionPrice(ctx, s.StripeSubscriptionID, plan.Definition.StripePriceID, quantity, proration, opts.Metadata, &idemKey)
	if err != nil {
		return Subscription{}, err
	}
	var metaJSON []byte
	if len(opts.Metadata) > 0 {
		metaJSON, _ = json.Marshal(opts.Metadata)
	}
	row := c.db.QueryRowxContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET status=$2, current_period_start=$3, current_period_end=$4, metadata=COALESCE($5::jsonb, metadata), updated_at=now() WHERE id=$1
        RETURNING id, customer_id, stripe_subscription_id, status, current_period_start, current_period_end, cancel_at_period_end, metadata, created_at, updated_at`, s.ID, status, curStart, curEnd, nullBytes(metaJSON))
	if err := row.StructScan(&s); err != nil {
		return Subscription{}, err
	}
	return s, nil
}

// CancelSubscription cancels a subscription immediately or at period end.
func (c *service) CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error {
	var s Subscription
	if err := c.db.GetContext(ctx, &s, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE id=$1`, subID); err != nil {
		return ErrNotFound("subscription not found")
	}
	status, curStart, curEnd, err := c.stripe.CancelSubscription(ctx, s.StripeSubscriptionID, atPeriodEnd, nil)
	if err != nil {
		return err
	}
	if atPeriodEnd {
		_, err = c.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET cancel_at_period_end=true, status=$2, current_period_start=$3, current_period_end=$4, updated_at=now() WHERE id=$1`, s.ID, status, curStart, curEnd)
		return err
	}
	_, err = c.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET cancel_at_period_end=false, status=$2, current_period_start=$3, current_period_end=$4, updated_at=now() WHERE id=$1`, s.ID, status, curStart, curEnd)
	return err
}
