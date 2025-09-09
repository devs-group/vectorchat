BEGIN;

-- Customers: lookup by email
CREATE INDEX IF NOT EXISTS idx_stripe_sub_customers_email
  ON stripe_sub_pkg_customers (email);

-- Subscriptions: common queries
CREATE INDEX IF NOT EXISTS idx_stripe_sub_subscriptions_customer_updated_at
  ON stripe_sub_pkg_subscriptions (customer_id, updated_at DESC);

-- Subscriptions: active-like preference and ordering
CREATE INDEX IF NOT EXISTS idx_stripe_sub_subscriptions_cust_status_cancel_period
  ON stripe_sub_pkg_subscriptions (
    customer_id,
    (lower(status)),
    cancel_at_period_end,
    current_period_end DESC,
    updated_at DESC
  );

-- Plans: inference by amount/currency/interval among active
CREATE INDEX IF NOT EXISTS idx_stripe_sub_plans_active_amount_currency_interval
  ON stripe_sub_pkg_plans (active, amount_cents, currency, billing_interval);

COMMIT;
