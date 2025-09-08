package stripe_sub

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

func CheckoutSessionHTTPHandler(c Client, log Logger) http.Handler {
	if log == nil {
		log = &nopLogger{}
	}
	type reqBody struct {
		CustomerID          string            `json:"customer_id"`
		PlanKey             string            `json:"plan_key"`
		SuccessURL          string            `json:"success_url"`
		CancelURL           string            `json:"cancel_url"`
		Quantity            *int              `json:"quantity"`
		AllowPromotionCodes bool              `json:"allow_promotion_codes"`
		IdempotencyKey      *string           `json:"idempotency_key"`
		Metadata            map[string]string `json:"metadata"`
	}
	type respBody struct {
		SessionID string `json:"session_id"`
		URL       string `json:"url"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		body, err := io.ReadAll(io.LimitReader(r.Body, 64<<10))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var in reqBody
		if err := json.Unmarshal(body, &in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if in.CustomerID == "" || in.PlanKey == "" || in.SuccessURL == "" || in.CancelURL == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}
		sid, url, err := c.CreateCheckoutSessionForPlan(r.Context(), in.CustomerID, in.PlanKey, in.SuccessURL, in.CancelURL, CheckoutSessionOpts{Quantity: in.Quantity, AllowPromotionCodes: in.AllowPromotionCodes, IdempotencyKey: in.IdempotencyKey, Metadata: in.Metadata})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respBody{SessionID: sid, URL: url})
	})
}

func SetupIntentHTTPHandler(c Client, log Logger) http.Handler {
	if log == nil {
		log = &nopLogger{}
	}
	type reqBody struct {
		CustomerID string `json:"customer_id"`
	}
	type respBody struct {
		ClientSecret string `json:"client_secret"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		body, err := io.ReadAll(io.LimitReader(r.Body, 16<<10))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var in reqBody
		if err := json.Unmarshal(body, &in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if in.CustomerID == "" {
			http.Error(w, "missing customer_id", http.StatusBadRequest)
			return
		}
		secret, err := c.CreateSetupIntent(context.Background(), in.CustomerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respBody{ClientSecret: secret})
	})
}

type Hooks struct {
	OnSubscriptionActivated func(ctx context.Context, s Subscription)
	OnSubscriptionCanceled  func(ctx context.Context, s Subscription)
	OnPaymentFailed         func(ctx context.Context, s Subscription, reason string)
}

func (c *service) WebhookHandler(h Hooks) (httpHandler, error) {
	if c.webhookSecret == "" {
		return nil, ErrConfig("webhook secret not configured")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sig := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(body, sig, c.webhookSecret)
		if err != nil {
			c.log.Warnf("stripe webhook signature verify failed: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch stripe.EventType(event.Type) {
		case stripe.EventTypeCheckoutSessionCompleted:
			c.handleCheckoutCompleted(r.Context(), event, h)
		case stripe.EventTypeCustomerSubscriptionCreated, stripe.EventTypeCustomerSubscriptionUpdated:
			c.handleSubscriptionUpsert(r.Context(), event, h)
		case stripe.EventTypeCustomerSubscriptionDeleted:
			c.handleSubscriptionDeleted(r.Context(), event, h)
		case stripe.EventTypeInvoicePaymentSucceeded:
			c.handleInvoicePaymentSucceeded(r.Context(), event, h)
		case stripe.EventTypeInvoicePaymentFailed:
			c.handleInvoicePaymentFailed(r.Context(), event, h)
		}
		w.WriteHeader(http.StatusOK)
	}), nil
}

func (c *service) handleCheckoutCompleted(ctx context.Context, event stripe.Event, h Hooks) {
	var cs stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
		return
	}
	if cs.Mode != stripe.CheckoutSessionModeSubscription || cs.Subscription == nil {
		return
	}
	subID := cs.Subscription.ID
	c.upsertSubscriptionByID(ctx, subID, h)
}

func (c *service) handleSubscriptionUpsert(ctx context.Context, event stripe.Event, h Hooks) {
	var ss stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &ss); err != nil {
		return
	}
	c.upsertSubscriptionByID(ctx, ss.ID, h)
}

func (c *service) handleSubscriptionDeleted(ctx context.Context, event stripe.Event, h Hooks) {
	var ss stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &ss); err != nil {
		return
	}
	_, _ = c.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET status=$2, cancel_at_period_end=false, updated_at=now() WHERE stripe_subscription_id=$1`, ss.ID, string(ss.Status))
	if h.OnSubscriptionCanceled != nil {
		if s, ok := c.findLocalByStripeID(ctx, ss.ID); ok {
			h.OnSubscriptionCanceled(ctx, s)
		}
	}
}

func (c *service) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event, h Hooks) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return
	}
	if inv.Subscription == nil {
		return
	}
	_, _ = c.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET status='active', updated_at=now() WHERE stripe_subscription_id=$1`, inv.Subscription.ID)
}

func (c *service) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event, h Hooks) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return
	}
	if inv.Subscription == nil {
		return
	}
	reason := ""
	if inv.PaymentIntent != nil && inv.PaymentIntent.LastPaymentError != nil {
		reason = inv.PaymentIntent.LastPaymentError.Error()
	}
	if h.OnPaymentFailed != nil {
		if s, ok := c.findLocalByStripeID(ctx, inv.Subscription.ID); ok {
			h.OnPaymentFailed(ctx, s, reason)
		}
	}
}

func (c *service) upsertSubscriptionByID(ctx context.Context, stripeSubID string, h Hooks) {
	custID, status, curStart, curEnd, cancelAtPeriodEnd, err := c.stripe.GetSubscription(ctx, stripeSubID)
	if err != nil {
		c.log.Warnf("fetch subscription failed: %v", err)
		return
	}
	if custID == "" {
		c.log.Warnf("subscription has no customer id: %s", stripeSubID)
		return
	}
	var customer Customer
	err = c.db.GetContext(ctx, &customer, `SELECT * FROM stripe_sub_pkg_customers WHERE stripe_customer_id=$1`, custID)
	if err != nil {
		email, err2 := c.stripe.GetCustomer(ctx, custID)
		if err2 != nil {
			c.log.Warnf("fetch customer failed: %v", err2)
			return
		}
		row := c.db.QueryRowxContext(ctx, `INSERT INTO stripe_sub_pkg_customers (ext_id, stripe_customer_id, email) VALUES (NULL,$1,$2) RETURNING id, ext_id, stripe_customer_id, email, created_at, updated_at`, custID, email)
		if err2 = row.StructScan(&customer); err2 != nil {
			c.log.Warnf("insert customer failed: %v", err2)
			return
		}
	}
	var existingID string
	_ = c.db.GetContext(ctx, &existingID, `SELECT id FROM stripe_sub_pkg_subscriptions WHERE stripe_subscription_id=$1`, stripeSubID)
	if existingID == "" {
		_, err = c.db.ExecContext(ctx, `INSERT INTO stripe_sub_pkg_subscriptions (customer_id, stripe_subscription_id, status, current_period_start, current_period_end, cancel_at_period_end) VALUES ($1,$2,$3,$4,$5,$6)`, customer.ID, stripeSubID, status, curStart, curEnd, cancelAtPeriodEnd)
		if err != nil {
			c.log.Warnf("insert subscription failed: %v", err)
			return
		}
	} else {
		_, err = c.db.ExecContext(ctx, `UPDATE stripe_sub_pkg_subscriptions SET status=$2, current_period_start=$3, current_period_end=$4, cancel_at_period_end=$5, updated_at=now() WHERE stripe_subscription_id=$1`, stripeSubID, status, curStart, curEnd, cancelAtPeriodEnd)
		if err != nil {
			c.log.Warnf("update subscription failed: %v", err)
			return
		}
	}
	if h.OnSubscriptionActivated != nil && status == string(stripe.SubscriptionStatusActive) {
		if s, ok := c.findLocalByStripeID(ctx, stripeSubID); ok {
			h.OnSubscriptionActivated(ctx, s)
		}
	}
}

func (c *service) findLocalByStripeID(ctx context.Context, stripeSubID string) (Subscription, bool) {
	var s Subscription
	if err := c.db.GetContext(ctx, &s, `SELECT * FROM stripe_sub_pkg_subscriptions WHERE stripe_subscription_id=$1`, stripeSubID); err == nil {
		return s, true
	}
	return Subscription{}, false
}

// PlansHTTPHandler returns active plans as JSON
func PlansHTTPHandler(c Client, log Logger) http.Handler {
	if log == nil {
		log = &nopLogger{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		plans, err := c.GetPlans(r.Context())
		if err != nil {
			http.Error(w, "failed to load plans", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(plans)
	})
}
