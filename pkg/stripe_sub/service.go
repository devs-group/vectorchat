package stripe_sub

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	stripesdk "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/client"

	"github.com/jmoiron/sqlx"

	pkgdb "github.com/yourusername/vectorchat/pkg/stripe_sub/db"
)

// Config configures the subscription service.
type Config struct {
	DB           *sqlx.DB
	StripeAPIKey string
	// StripeAPIVersion sets the Stripe API version used by stripe-go (e.g. "2023-10-16").
	// If empty, the stripe-go default is used.
	StripeAPIVersion string
	// WebhookSecret is used by the webhook handler to verify signatures.
	WebhookSecret string
}

// Service is the main entrypoint to interact with the package.
type Service struct {
	db               *sqlx.DB
	webhookSecret    string
	stripe           *client.API
	stripeAPIVersion string
}

var ErrActiveSubscription = errors.New("an active subscription already exists")

// New initializes the package: sets Stripe key and runs migrations.
func New(ctx context.Context, cfg Config) (*Service, error) {
	if cfg.DB == nil {
		return nil, errors.New("stripe_sub: cfg.DB is required")
	}
	if strings.TrimSpace(cfg.StripeAPIKey) == "" {
		return nil, errors.New("stripe_sub: cfg.StripeAPIKey is required")
	}
	// Initialize Stripe client.
	stripesdk.Key = cfg.StripeAPIKey
	if strings.TrimSpace(cfg.StripeAPIVersion) == "" {
		{
			fmt.Println("Stripe API version not set; set STRIPE_API_VERSION to match your Stripe webhook endpoint version.")
			fmt.Println("Update in Stripe Dashboard: Developers → Webhooks → Select endpoint → Version")
		}
		return nil, errors.New("stripe_sub: cfg.StripeAPIVersion is required and must match your Stripe webhook endpoint version")
	}
	api := client.New(cfg.StripeAPIKey, nil)

	// Run migrations for this package.
	if err := (pkgdb.Runner{DB: cfg.DB}).Migrate(ctx); err != nil {
		return nil, fmt.Errorf("stripe_sub: migrate: %w", err)
	}

	s := &Service{db: cfg.DB, webhookSecret: cfg.WebhookSecret, stripe: api, stripeAPIVersion: cfg.StripeAPIVersion}
	return s, nil
}

// PlanParams defines the minimal information to create a plan.
type PlanParams struct {
	Key             string
	DisplayName     string
	Active          bool
	BillingInterval string // day, week, month, year
	AmountCents     int64
	Currency        string // e.g. "usd"
	Metadata        map[string]any
	PlanDefinition  map[string]any
}

// CreatePlan inserts a new plan. Key must be unique.
func (s *Service) CreatePlan(ctx context.Context, p PlanParams) (*Plan, error) {
	if p.Key == "" || p.DisplayName == "" {
		return nil, errors.New("key and display_name are required")
	}
	if p.BillingInterval == "" {
		p.BillingInterval = "month"
	}
	if p.Currency == "" {
		p.Currency = "usd"
	}
	var plan Plan
	q := `INSERT INTO stripe_sub_pkg_plans (key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition)
          VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
          RETURNING id, key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition, created_at, updated_at`
	if err := s.db.GetContext(ctx, &plan, q,
		p.Key, p.DisplayName, p.Active, p.BillingInterval, p.AmountCents, strings.ToLower(p.Currency), JSONB(p.Metadata), JSONB(p.PlanDefinition),
	); err != nil {
		return nil, err
	}
	return &plan, nil
}

// UpsertPlan creates a plan if missing, otherwise updates fields.
func (s *Service) UpsertPlan(ctx context.Context, p PlanParams) (*Plan, error) {
	if p.Key == "" || p.DisplayName == "" {
		return nil, errors.New("key and display_name are required")
	}
	if p.BillingInterval == "" {
		p.BillingInterval = "month"
	}
	if p.Currency == "" {
		p.Currency = "usd"
	}
	var plan Plan
	q := `INSERT INTO stripe_sub_pkg_plans (key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition)
          VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
          ON CONFLICT (key)
          DO UPDATE SET display_name=EXCLUDED.display_name, active=EXCLUDED.active, billing_interval=EXCLUDED.billing_interval,
                        amount_cents=EXCLUDED.amount_cents, currency=EXCLUDED.currency, metadata=EXCLUDED.metadata, plan_definition=EXCLUDED.plan_definition,
                        updated_at=now()
          RETURNING id, key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition, created_at, updated_at`
	if err := s.db.GetContext(ctx, &plan, q,
		p.Key, p.DisplayName, p.Active, p.BillingInterval, p.AmountCents, strings.ToLower(p.Currency), JSONB(p.Metadata), JSONB(p.PlanDefinition),
	); err != nil {
		return nil, err
	}
	return &plan, nil
}

// ListActivePlans returns active plans only.
func (s *Service) ListActivePlans(ctx context.Context) ([]Plan, error) {
	var plans []Plan
	if err := s.db.SelectContext(ctx, &plans, `SELECT * FROM stripe_sub_pkg_plans WHERE active = true ORDER BY amount_cents ASC`); err != nil {
		return nil, err
	}
	return plans, nil
}

// CreateCheckoutSession creates a Stripe Checkout Session for a plan and customer (by email/external ID).
// Returns the session ID and redirect URL.
func (s *Service) CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (string, string, error) {
	if strings.TrimSpace(req.PlanKey) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.SuccessURL) == "" || strings.TrimSpace(req.CancelURL) == "" {
		return "", "", errors.New("plan_key, email, success_url, cancel_url are required")
	}
	plan, err := s.getPlanByKey(ctx, req.PlanKey)
	if err != nil {
		return "", "", err
	}
	cust, err := s.EnsureCustomer(ctx, req.ExternalID, req.Email)
	if err != nil {
		return "", "", err
	}
	// Refresh from Stripe to avoid stale DB state, then block if active
    if err := s.syncStripeCustomerSubscriptions(ctx, cust.StripeCustomerID); err != nil {
        return "", "", err
    }
    if active, err := s.hasBlockingActiveSubscription(ctx, cust.ID); err != nil {
        return "", "", err
    } else if active {
        return "", "", ErrActiveSubscription
    }

	params := &stripesdk.CheckoutSessionParams{
		Mode:       stripesdk.String(string(stripesdk.CheckoutSessionModeSubscription)),
		Customer:   stripesdk.String(cust.StripeCustomerID),
		SuccessURL: stripesdk.String(req.SuccessURL),
		CancelURL:  stripesdk.String(req.CancelURL),
		LineItems: []*stripesdk.CheckoutSessionLineItemParams{
			{
				Quantity: stripesdk.Int64(1),
				PriceData: &stripesdk.CheckoutSessionLineItemPriceDataParams{
					Currency: stripesdk.String(strings.ToLower(plan.Currency)),
					ProductData: &stripesdk.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripesdk.String(plan.DisplayName),
					},
					Recurring: &stripesdk.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripesdk.String(string(stripesdk.PriceRecurringInterval(plan.BillingInterval))),
					},
					UnitAmount: stripesdk.Int64(plan.AmountCents),
				},
			},
		},
		SubscriptionData: &stripesdk.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{"plan_key": plan.Key},
		},
	}
	sess, err := s.stripe.CheckoutSessions.New(params)
	if err != nil {
		return "", "", err
	}
	return sess.ID, sess.URL, nil
}

// PortalRequest defines inputs for creating a Stripe Customer Portal session.
type PortalRequest struct {
	ExternalID *string
	Email      string
	ReturnURL  string
}

// CreatePortalSession creates a Stripe Customer Portal session URL for the customer.
func (s *Service) CreatePortalSession(ctx context.Context, req PortalRequest) (string, error) {
	if strings.TrimSpace(req.ReturnURL) == "" || strings.TrimSpace(req.Email) == "" {
		return "", errors.New("return_url and email are required")
	}
	cust, err := s.EnsureCustomer(ctx, req.ExternalID, req.Email)
	if err != nil {
		return "", err
	}
	ps, err := s.stripe.BillingPortalSessions.New(&stripesdk.BillingPortalSessionParams{
		Customer:  stripesdk.String(cust.StripeCustomerID),
		ReturnURL: stripesdk.String(req.ReturnURL),
	})
	if err != nil {
		return "", err
	}
	return ps.URL, nil
}

// inferPlanKeyFromStripeSub tries to find a matching plan in DB based on the first subscription item's price.
func (s *Service) inferPlanKeyFromStripeSub(sub *stripesdk.Subscription) (string, error) {
	if sub == nil || sub.Items == nil || len(sub.Items.Data) == 0 {
		return "", nil
	}
	item := sub.Items.Data[0]
	if item.Price == nil {
		return "", nil
	}
	price := item.Price
	var interval string
	if price.Recurring != nil {
		interval = string(price.Recurring.Interval)
	}
	currency := string(price.Currency)
	amount := price.UnitAmount
	if interval == "" || currency == "" || amount == 0 {
		return "", nil
	}
	var key string
	// Match active plan by amount + currency + interval
	err := s.db.Get(&key, `SELECT key FROM stripe_sub_pkg_plans WHERE active=true AND amount_cents=$1 AND currency=$2 AND billing_interval=$3 LIMIT 1`, amount, strings.ToLower(currency), interval)
	if err != nil {
		return "", err
	}
	return key, nil
}

// hasActiveSubscription checks if the customer has an active-like subscription.
func (s *Service) hasActiveSubscription(ctx context.Context, customerID string) (bool, *Subscription, error) {
	var sub Subscription
	err := s.db.GetContext(ctx, &sub, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE customer_id=$1 ORDER BY updated_at DESC LIMIT 1`, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, err
	}
	if isActiveLike(sub.Status) {
		return true, &sub, nil
	}
	return false, &sub, nil
}

func isActiveLike(status string) bool {
	switch strings.ToLower(status) {
	case "active", "trialing", "past_due":
		return true
	default:
		return false
	}
}

// hasBlockingActiveSubscription returns true if there exists an active-like subscription that is NOT scheduled to cancel.
func (s *Service) hasBlockingActiveSubscription(ctx context.Context, customerID string) (bool, error) {
    var exists bool
    err := s.db.GetContext(ctx, &exists, `
        SELECT EXISTS (
            SELECT 1 FROM stripe_sub_pkg_subscriptions
            WHERE customer_id=$1
              AND lower(status) IN ('active','trialing','past_due')
              AND cancel_at_period_end = false
        )`, customerID)
    if err != nil { return false, err }
    return exists, nil
}

// EnsureSubscriptionPlanKey ensures the subscription row has a plan_key in metadata, attempting Stripe fetch + inference.
func (s *Service) EnsureSubscriptionPlanKey(ctx context.Context, stripeSubscriptionID string) (string, error) {
	sub, err := s.stripe.Subscriptions.Get(stripeSubscriptionID, nil)
	if err != nil {
		return "", err
	}
	key, err := s.inferPlanKeyFromStripeSub(sub)
	if err != nil || key == "" {
		return key, err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('plan_key', $2), updated_at = now() WHERE stripe_subscription_id = $1`, stripeSubscriptionID, key)
	if err != nil {
		return "", err
	}
	return key, nil
}

// syncStripeCustomerSubscriptions pulls latest subscriptions for a Stripe customer and upserts them locally.
func (s *Service) syncStripeCustomerSubscriptions(ctx context.Context, stripeCustomerID string) error {
	lp := &stripesdk.SubscriptionListParams{Customer: stripesdk.String(stripeCustomerID)}
	lp.Status = stripesdk.String("all")
	it := s.stripe.Subscriptions.List(lp)
	for it.Next() {
		sub := it.Subscription()
		if sub == nil {
			continue
		}
		// Persist subscription snapshot
		var ps, pe *time.Time
		if sub.CurrentPeriodStart != 0 {
			t := time.Unix(sub.CurrentPeriodStart, 0)
			ps = &t
		}
		if sub.CurrentPeriodEnd != 0 {
			t := time.Unix(sub.CurrentPeriodEnd, 0)
			pe = &t
		}
		meta := map[string]any(nil)
		if sub.Metadata != nil {
			meta = make(map[string]any, len(sub.Metadata))
			for k, v := range sub.Metadata {
				meta[k] = v
			}
		}
		if meta == nil || meta["plan_key"] == nil {
			if key, _ := s.inferPlanKeyFromStripeSub(sub); key != "" {
				if meta == nil {
					meta = map[string]any{}
				}
				meta["plan_key"] = key
			}
		}
		// Need local customer id
		cust, err := s.findCustomerByStripeID(ctx, sub.Customer.ID)
		if err != nil {
			continue
		}
		_ = s.upsertSubscription(ctx, cust.ID, sub.ID, string(sub.Status), ps, pe, sub.CancelAtPeriodEnd, meta)
	}
	return it.Err()
}

// getPlanByKey returns a single plan by key.
func (s *Service) getPlanByKey(ctx context.Context, key string) (*Plan, error) {
	var p Plan
	err := s.db.GetContext(ctx, &p, `SELECT * FROM stripe_sub_pkg_plans WHERE key=$1 AND active=true`, key)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPlan returns an active plan by its key.
// It is a thin wrapper around the internal lookup.
func (s *Service) GetPlan(ctx context.Context, key string) (*Plan, error) {
	return s.getPlanByKey(ctx, key)
}

// EnsureCustomer ensures we have a local + Stripe customer for the given external ID or email.
func (s *Service) EnsureCustomer(ctx context.Context, externalID *string, email string) (*Customer, error) {
	// Try by external id if provided
	var c Customer
	if externalID != nil && *externalID != "" {
		if err := s.db.GetContext(ctx, &c, `SELECT * FROM stripe_sub_pkg_customers WHERE ext_id=$1`, *externalID); err == nil {
			return &c, nil
		}
	}
	// Try by email
	if err := s.db.GetContext(ctx, &c, `SELECT * FROM stripe_sub_pkg_customers WHERE email=$1 ORDER BY created_at ASC LIMIT 1`, email); err == nil {
		return &c, nil
	}
	// Create Stripe customer and store
	cu, err := s.stripe.Customers.New(&stripesdk.CustomerParams{
		Email: stripesdk.String(email),
		Metadata: func() map[string]string {
			if externalID != nil && *externalID != "" {
				return map[string]string{"external_id": *externalID}
			}
			return nil
		}(),
	})
	if err != nil {
		return nil, err
	}

	q := `INSERT INTO stripe_sub_pkg_customers (ext_id, stripe_customer_id, email) VALUES ($1,$2,$3)
          RETURNING id, ext_id, stripe_customer_id, email, created_at, updated_at`
	var ext any
	if externalID != nil && *externalID != "" {
		ext = *externalID
	}
	if err := s.db.GetContext(ctx, &c, q, ext, cu.ID, email); err != nil {
		return nil, err
	}
	return &c, nil
}

// upsertSubscription writes the subscription state for a customer.
func (s *Service) upsertSubscription(ctx context.Context, customerID string, subID string, status string,
	periodStart, periodEnd *time.Time, cancelAtPeriodEnd bool, metadata map[string]any,
) error {
	q := `INSERT INTO stripe_sub_pkg_subscriptions (customer_id, stripe_subscription_id, status, current_period_start, current_period_end, cancel_at_period_end, metadata)
          VALUES ($1,$2,$3,$4,$5,$6,$7)
          ON CONFLICT (stripe_subscription_id)
          DO UPDATE SET status=EXCLUDED.status, current_period_start=EXCLUDED.current_period_start, current_period_end=EXCLUDED.current_period_end,
                        cancel_at_period_end=EXCLUDED.cancel_at_period_end, metadata=EXCLUDED.metadata, updated_at=now()`
	_, err := s.db.ExecContext(ctx, q, customerID, subID, status, periodStart, periodEnd, cancelAtPeriodEnd, JSONB(metadata))
	return err
}

// findCustomerByStripeID returns a local customer by Stripe customer id.
func (s *Service) findCustomerByStripeID(ctx context.Context, stripeCustomerID string) (*Customer, error) {
	var c Customer
	if err := s.db.GetContext(ctx, &c, `SELECT * FROM stripe_sub_pkg_customers WHERE stripe_customer_id=$1`, stripeCustomerID); err != nil {
		return nil, err
	}
	return &c, nil
}

// lookupCustomer returns an existing customer by ext_id or email. Does not create new records.
func (s *Service) lookupCustomer(ctx context.Context, externalID *string, email string) (*Customer, error) {
	var c Customer
	if externalID != nil && *externalID != "" {
		if err := s.db.GetContext(ctx, &c, `SELECT * FROM stripe_sub_pkg_customers WHERE ext_id=$1`, *externalID); err == nil {
			return &c, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if err := s.db.GetContext(ctx, &c, `SELECT * FROM stripe_sub_pkg_customers WHERE email=$1 ORDER BY created_at ASC LIMIT 1`, email); err == nil {
		return &c, nil
	} else if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else {
		return nil, err
	}
}

// GetUserSubscription returns the most recently updated subscription for a given user identity.
func (s *Service) GetUserSubscription(ctx context.Context, externalID *string, email string) (*Subscription, error) {
	cust, err := s.lookupCustomer(ctx, externalID, email)
	if err != nil || cust == nil {
		return nil, err
	}
	var sub Subscription
	err = s.db.GetContext(ctx, &sub, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE customer_id=$1 ORDER BY updated_at DESC LIMIT 1`, cust.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// GetUserCurrentSubscription prefers an active-like subscription; falls back to the latest record if none active.
func (s *Service) GetUserCurrentSubscription(ctx context.Context, externalID *string, email string) (*Subscription, error) {
    cust, err := s.lookupCustomer(ctx, externalID, email)
    if err != nil || cust == nil {
        return nil, err
    }
    var sub Subscription
    // Prefer active-like subscriptions; among them, prioritize those not scheduled to cancel,
    // then by furthest current_period_end, then by most recently updated.
    err = s.db.GetContext(ctx, &sub, `
        SELECT * FROM stripe_sub_pkg_subscriptions
        WHERE customer_id=$1 AND lower(status) IN ('active','trialing','past_due')
        ORDER BY (CASE WHEN cancel_at_period_end THEN 1 ELSE 0 END) ASC,
                 current_period_end DESC NULLS LAST,
                 updated_at DESC
        LIMIT 1`, cust.ID)
    if err == nil { return &sub, nil }
    if !errors.Is(err, sql.ErrNoRows) { return nil, err }
    // Fallback to most recent by updated_at
    err = s.db.GetContext(ctx, &sub, `
        SELECT * FROM stripe_sub_pkg_subscriptions
        WHERE customer_id=$1
        ORDER BY updated_at DESC
        LIMIT 1`, cust.ID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, nil }
        return nil, err
    }
    return &sub, nil
}

// GetUserPlan returns the user's active plan and the latest subscription record.
// If there is no subscription or it is not active, plan will be nil.
func (s *Service) GetUserPlan(ctx context.Context, externalID *string, email string) (*Plan, *Subscription, error) {
    sub, err := s.GetUserCurrentSubscription(ctx, externalID, email)
	if err != nil || sub == nil {
		return nil, sub, err
	}
	if !IsSubscriptionActive(sub, time.Now()) {
		return nil, sub, nil
	}
	// Expect plan_key in subscription metadata (set during Checkout session creation).
	var planKey string
	if sub.Metadata != nil {
		if v, ok := sub.Metadata["plan_key"]; ok {
			switch t := v.(type) {
			case string:
				planKey = t
			}
		}
	}
	if strings.TrimSpace(planKey) == "" {
		// No recorded plan key; caller may choose to treat as no plan.
		return nil, sub, nil
	}
	plan, err := s.GetPlan(ctx, planKey)
	if err != nil {
		return nil, sub, err
	}
	return plan, sub, nil
}

// IsSubscriptionActive determines if a subscription should be treated as active for access control.
// Policy: active, trialing, and past_due are treated as active; others are not.
func IsSubscriptionActive(sub *Subscription, now time.Time) bool {
	if sub == nil {
		return false
	}
	switch strings.ToLower(sub.Status) {
	case "active", "trialing", "past_due":
		return true
	default:
		return false
	}
}

// RefreshLatestSubscription pulls latest from Stripe for a user and updates DB, returning the newest local record.
func (s *Service) RefreshLatestSubscription(ctx context.Context, externalID *string, email string) (*Subscription, error) {
	cust, err := s.lookupCustomer(ctx, externalID, email)
	if err != nil || cust == nil {
		return nil, err
	}
	if err := s.syncStripeCustomerSubscriptions(ctx, cust.StripeCustomerID); err != nil {
		return nil, err
	}
	var out Subscription
	if err := s.db.GetContext(ctx, &out, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE customer_id=$1 ORDER BY updated_at DESC LIMIT 1`, cust.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}
