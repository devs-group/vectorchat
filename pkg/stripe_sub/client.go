package stripe_sub

import (
	"context"
	"strings"
)

// UpsertCustomer finds or creates a customer with the given email and optional external id.
func (c *service) UpsertCustomer(ctx context.Context, email, externalID string) (Customer, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var cust Customer
	if externalID != "" {
		if err := c.db.GetContext(ctx, &cust, `SELECT * FROM stripe_sub_pkg_customers WHERE ext_id=$1`, externalID); err == nil {
			return cust, nil
		}
	}
	if err := c.db.GetContext(ctx, &cust, `SELECT * FROM stripe_sub_pkg_customers WHERE email=$1 ORDER BY created_at ASC LIMIT 1`, email); err == nil {
		if externalID != "" && (cust.ExternalID == nil || *cust.ExternalID == "") {
			_, _ = c.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_customers SET ext_id=$1, updated_at=now() WHERE id=$2`, externalID, cust.ID)
			cust.ExternalID = &externalID
		}
		return cust, nil
	}
	scid, err := c.stripe.CreateCustomer(ctx, email)
	if err != nil {
		return Customer{}, err
	}
	row := c.db.QueryRowxContext(ctx, `INSERT INTO stripe_sub_pkg_customers (ext_id, stripe_customer_id, email) VALUES ($1,$2,$3) RETURNING id, ext_id, stripe_customer_id, email, created_at, updated_at`, nullIfEmpty(externalID), scid, email)
	if err := row.StructScan(&cust); err != nil {
		return Customer{}, err
	}
	return cust, nil
}

func (c *service) GetSubscription(ctx context.Context, subID string) (Subscription, error) {
	var s Subscription
	if err := c.db.GetContext(ctx, &s, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE id=$1`, subID); err != nil {
		return Subscription{}, ErrNotFound("subscription not found")
	}
	return s, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func prorationBehavior(prorate *bool) string {
	if prorate != nil && *prorate {
		return "create_prorations"
	}
	return "none"
}
