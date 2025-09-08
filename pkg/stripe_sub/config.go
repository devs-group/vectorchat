package stripe_sub

import (
    "context"
    "database/sql"
    "net/http"
    "os"
    "time"

    "github.com/jmoiron/sqlx"
    sdb "github.com/yourusername/vectorchat/pkg/stripe_sub/db"
)

type Logger interface {
    Debugf(format string, args ...any)
    Infof(format string, args ...any)
    Warnf(format string, args ...any)
    Errorf(format string, args ...any)
}

type nopLogger struct{}

func (n *nopLogger) Debugf(string, ...any) {}
func (n *nopLogger) Infof(string, ...any)  {}
func (n *nopLogger) Warnf(string, ...any)  {}
func (n *nopLogger) Errorf(string, ...any) {}

type Options struct {
    DB            *sqlx.DB
    Stripe        StripeAPI
    Logger        Logger
    WebhookSecret string
}

type Option func(*Options)

func WithDB(db *sqlx.DB) Option { return func(o *Options) { o.DB = db } }
func WithStripe(s StripeAPI) Option { return func(o *Options) { o.Stripe = s } }
func WithLogger(l Logger) Option { return func(o *Options) { o.Logger = l } }
func WithWebhookSecret(secret string) Option { return func(o *Options) { o.WebhookSecret = secret } }

type Client interface {
    UpsertCustomer(ctx context.Context, email, externalID string) (Customer, error)

    UpdateSubscription(ctx context.Context, subID, newPlanKey string, opts UpdateSubOpts) (Subscription, error)
    CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error
    GetSubscription(ctx context.Context, subID string) (Subscription, error)

    GetPlan(ctx context.Context, planKey string) (Plan, error)
    ListActivePlans(ctx context.Context) ([]Plan, error)
    GetPlans(ctx context.Context) ([]Plan, error)

    EnsurePlans(ctx context.Context, specs []PlanSpec) error

    CreateCheckoutSessionForPlan(ctx context.Context, customerID, planKey, successURL, cancelURL string, opts CheckoutSessionOpts) (sessionID, url string, err error)
    CreateSetupIntent(ctx context.Context, customerID string) (clientSecret string, err error)

    Migrate(ctx context.Context) error
    WebhookHandler(h Hooks) (httpHandler, error)
}

type service struct {
    db            *sqlx.DB
    stripe        StripeAPI
    log           Logger
    webhookSecret string
}

func NewClient(opts ...Option) (Client, error) {
    o := &Options{}
    for _, fn := range opts { fn(o) }
    if o.Logger == nil { o.Logger = &nopLogger{} }
    if o.DB == nil {
        dsn := os.Getenv("DATABASE_URL")
        if dsn == "" { return nil, ErrConfig("DATABASE_URL not set and DB not provided") }
        db, err := sqlx.Open("postgres", dsn)
        if err != nil { return nil, err }
        if err := pingSoft(db.DB); err != nil { return nil, err }
        o.DB = db
    }
    if o.Stripe == nil {
        apiKey := os.Getenv("STRIPE_API_KEY")
        if apiKey == "" { return nil, ErrConfig("STRIPE_API_KEY not set and Stripe client not provided") }
        o.Stripe = NewStripeClient(apiKey)
    }
    c := &service{db: o.DB, stripe: o.Stripe, log: o.Logger, webhookSecret: o.WebhookSecret}
    return c, nil
}

func (c *service) Migrate(ctx context.Context) error { return sdb.Runner{DB: c.db}.Migrate(ctx) }

func pingSoft(db *sql.DB) error {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    return db.PingContext(ctx)
}

type httpHandler = http.Handler

