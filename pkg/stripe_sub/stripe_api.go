package stripe_sub

import (
    "context"
    "time"
)

type StripeAPI interface {
    CreateCustomer(ctx context.Context, email string) (stripeCustomerID string, err error)
    AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error
    SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error
    CreateSubscription(ctx context.Context, customerID, priceID string, quantity int, trialDays int, coupon *string, prorationBehavior string, metadata map[string]string, idemKey *string) (subID string, status string, curStart, curEnd *time.Time, err error)
    UpdateSubscriptionPrice(ctx context.Context, subID, newPriceID string, quantity int, prorationBehavior string, metadata map[string]string, idemKey *string) (status string, curStart, curEnd *time.Time, err error)
    CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool, idemKey *string) (status string, curStart, curEnd *time.Time, err error)

    CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string, allowPromotionCodes bool, quantity int, metadata map[string]string, idemKey *string) (sessionID, url string, err error)
    CreateSetupIntent(ctx context.Context, customerID string) (clientSecret string, err error)

    GetCustomer(ctx context.Context, customerID string) (email string, err error)
    GetSubscription(ctx context.Context, subID string) (customerID, status string, curStart, curEnd *time.Time, cancelAtPeriodEnd bool, err error)
}

func NewStripeClient(apiKey string) StripeAPI { return newStripeGoClient(apiKey) }

