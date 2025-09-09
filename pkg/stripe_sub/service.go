package stripe_sub

import (
    "database/sql"
    "context"
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
	// WebhookSecret is used by the webhook handler to verify signatures.
	WebhookSecret string
}

// Service is the main entrypoint to interact with the package.
type Service struct {
	db            *sqlx.DB
	webhookSecret string
	stripe        *client.API
}

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
	api := client.New(cfg.StripeAPIKey, nil)

	// Run migrations for this package.
	if err := (pkgdb.Runner{DB: cfg.DB}).Migrate(ctx); err != nil {
		return nil, fmt.Errorf("stripe_sub: migrate: %w", err)
	}

	s := &Service{db: cfg.DB, webhookSecret: cfg.WebhookSecret, stripe: api}
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
						Interval: stripesdk.String(string(stripesdk.PriceRecurringIntervalMonth)),
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
    if err != nil { return "", err }
    ps, err := s.stripe.BillingPortalSessions.New(&stripesdk.BillingPortalSessionParams{
        Customer:  stripesdk.String(cust.StripeCustomerID),
        ReturnURL: stripesdk.String(req.ReturnURL),
    })
    if err != nil { return "", err }
    return ps.URL, nil
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
        if errors.Is(err, sql.ErrNoRows) { return nil, nil }
        return nil, err
    }
    return &sub, nil
}
