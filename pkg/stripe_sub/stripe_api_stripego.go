package stripe_sub

import (
    "context"
    "fmt"
    "time"

    stripe "github.com/stripe/stripe-go/v76"
    checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
    "github.com/stripe/stripe-go/v76/client"
    pm "github.com/stripe/stripe-go/v76/paymentmethod"
    setupintent "github.com/stripe/stripe-go/v76/setupintent"
    sub "github.com/stripe/stripe-go/v76/subscription"
)

type stripeGoClient struct{ c *client.API }

func newStripeGoClient(apiKey string) StripeAPI {
    sc := &client.API{}
    sc.Init(apiKey, nil)
    return &stripeGoClient{c: sc}
}

func (s *stripeGoClient) CreateCustomer(ctx context.Context, email string) (string, error) {
    cp := &stripe.CustomerParams{Email: stripe.String(email)}
    cp.Context = ctx
    cus, err := s.c.Customers.New(cp)
    if err != nil { return "", err }
    return cus.ID, nil
}

func (s *stripeGoClient) AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
    ap := &stripe.PaymentMethodAttachParams{Customer: stripe.String(customerID)}
    ap.Context = ctx
    _, err := pm.Attach(paymentMethodID, ap)
    return err
}

func (s *stripeGoClient) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
    up := &stripe.CustomerParams{InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{DefaultPaymentMethod: stripe.String(paymentMethodID)}}
    up.Context = ctx
    _, err := s.c.Customers.Update(customerID, up)
    return err
}

func (s *stripeGoClient) CreateSubscription(ctx context.Context, customerID, priceID string, quantity int, trialDays int, coupon *string, prorationBehavior string, metadata map[string]string, idemKey *string) (string, string, *time.Time, *time.Time, error) {
    items := []*stripe.SubscriptionItemsParams{{Price: stripe.String(priceID), Quantity: stripe.Int64(int64(quantity))}}
    params := &stripe.SubscriptionParams{Customer: stripe.String(customerID), Items: items, ProrationBehavior: stripe.String(prorationBehavior), Metadata: metadata}
    params.Context = ctx
    if trialDays > 0 { params.TrialPeriodDays = stripe.Int64(int64(trialDays)) }
    if coupon != nil && *coupon != "" { params.Coupon = stripe.String(*coupon) }
    if idemKey != nil && *idemKey != "" { params.IdempotencyKey = stripe.String(*idemKey) }
    ss, err := s.c.Subscriptions.New(params)
    if err != nil { return "", "", nil, nil, err }
    return ss.ID, string(ss.Status), toPtrTime(ss.CurrentPeriodStart), toPtrTime(ss.CurrentPeriodEnd), nil
}

func (s *stripeGoClient) UpdateSubscriptionPrice(ctx context.Context, subID, newPriceID string, quantity int, prorationBehavior string, metadata map[string]string, idemKey *string) (string, *time.Time, *time.Time, error) {
    params := &stripe.SubscriptionParams{ProrationBehavior: stripe.String(prorationBehavior), Metadata: metadata, Items: []*stripe.SubscriptionItemsParams{{Price: stripe.String(newPriceID), Quantity: stripe.Int64(int64(quantity))}}}
    params.Context = ctx
    if idemKey != nil && *idemKey != "" { params.IdempotencyKey = stripe.String(*idemKey) }
    ss, err := s.c.Subscriptions.Update(subID, params)
    if err != nil { return "", nil, nil, err }
    return string(ss.Status), toPtrTime(ss.CurrentPeriodStart), toPtrTime(ss.CurrentPeriodEnd), nil
}

func (s *stripeGoClient) CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool, idemKey *string) (string, *time.Time, *time.Time, error) {
    params := &stripe.SubscriptionCancelParams{InvoiceNow: stripe.Bool(false), Prorate: stripe.Bool(false)}
    params.Context = ctx
    if atPeriodEnd {
        up := &stripe.SubscriptionParams{CancelAtPeriodEnd: stripe.Bool(true)}
        up.Context = ctx
        ss, err := s.c.Subscriptions.Update(subID, up)
        if err != nil { return "", nil, nil, err }
        return string(ss.Status), toPtrTime(ss.CurrentPeriodStart), toPtrTime(ss.CurrentPeriodEnd), nil
    }
    if idemKey != nil && *idemKey != "" { params.IdempotencyKey = stripe.String(*idemKey) }
    ss, err := sub.Cancel(subID, params)
    if err != nil { return "", nil, nil, err }
    return string(ss.Status), toPtrTime(ss.CurrentPeriodStart), toPtrTime(ss.CurrentPeriodEnd), nil
}

func (s *stripeGoClient) CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string, allowPromotionCodes bool, quantity int, metadata map[string]string, idemKey *string) (string, string, error) {
    q := int64(1)
    if quantity > 0 { q = int64(quantity) }
    params := &stripe.CheckoutSessionParams{
        Customer: stripe.String(customerID),
        Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        SuccessURL: stripe.String(successURL),
        CancelURL: stripe.String(cancelURL),
        LineItems: []*stripe.CheckoutSessionLineItemParams{{Price: stripe.String(priceID), Quantity: stripe.Int64(q)}},
        Metadata: metadata,
    }
    params.Context = ctx
    if allowPromotionCodes { params.AllowPromotionCodes = stripe.Bool(true) }
    if idemKey != nil && *idemKey != "" { params.IdempotencyKey = stripe.String(*idemKey) }
    sess, err := checkoutsession.New(params)
    if err != nil { return "", "", err }
    return sess.ID, sess.URL, nil
}

func (s *stripeGoClient) CreateSetupIntent(ctx context.Context, customerID string) (string, error) {
    params := &stripe.SetupIntentParams{Customer: stripe.String(customerID), Usage: stripe.String(string(stripe.SetupIntentUsageOffSession))}
    params.Context = ctx
    si, err := setupintent.New(params)
    if err != nil { return "", err }
    if si.ClientSecret == "" { return "", fmt.Errorf("empty client secret from setup intent") }
    return si.ClientSecret, nil
}

func (s *stripeGoClient) GetCustomer(ctx context.Context, customerID string) (string, error) {
    cp := &stripe.CustomerParams{}
    cp.Context = ctx
    cu, err := s.c.Customers.Get(customerID, cp)
    if err != nil { return "", err }
    return cu.Email, nil
}

func (s *stripeGoClient) GetSubscription(ctx context.Context, subID string) (string, string, *time.Time, *time.Time, bool, error) {
    sp := &stripe.SubscriptionParams{}
    sp.Context = ctx
    ss, err := s.c.Subscriptions.Get(subID, sp)
    if err != nil { return "", "", nil, nil, false, err }
    custID := ""
    if ss.Customer != nil { custID = ss.Customer.ID }
    return custID, string(ss.Status), toPtrTime(ss.CurrentPeriodStart), toPtrTime(ss.CurrentPeriodEnd), ss.CancelAtPeriodEnd, nil
}

func toPtrTime(ts int64) *time.Time { if ts == 0 { return nil }; t := time.Unix(ts, 0).UTC(); return &t }

