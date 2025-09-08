package stripe_sub

import "context"

func (c *service) CreateCheckoutSessionForPlan(ctx context.Context, customerID, planKey, successURL, cancelURL string, opts CheckoutSessionOpts) (string, string, error) {
    var cust Customer
    if err := c.db.GetContext(ctx, &cust, `SELECT * FROM stripe_sub_pkg_customers WHERE id=$1`, customerID); err != nil { return "", "", ErrNotFound("customer not found") }
    plan, err := c.GetPlan(ctx, planKey)
    if err != nil { return "", "", err }
    qty := 1
    if opts.Quantity != nil && *opts.Quantity > 0 { qty = *opts.Quantity }
    idemKey := ""
    if opts.IdempotencyKey != nil { idemKey = *opts.IdempotencyKey }
    sid, url, err := c.stripe.CreateCheckoutSession(ctx, cust.StripeCustomer, plan.Definition.StripePriceID, successURL, cancelURL, opts.AllowPromotionCodes, qty, opts.Metadata, &idemKey)
    if err != nil { return "", "", err }
    return sid, url, nil
}

func (c *service) CreateSetupIntent(ctx context.Context, customerID string) (string, error) {
    var cust Customer
    if err := c.db.GetContext(ctx, &cust, `SELECT * FROM stripe_sub_pkg_customers WHERE id=$1`, customerID); err != nil { return "", ErrNotFound("customer not found") }
    return c.stripe.CreateSetupIntent(ctx, cust.StripeCustomer)
}

