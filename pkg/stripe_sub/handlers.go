package stripe_sub

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
    "time"

    stripesdk "github.com/stripe/stripe-go/v76"
    "github.com/stripe/stripe-go/v76/webhook"
)

// PlansHandler returns an http.HandlerFunc that lists active plans.
func (s *Service) PlansHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		plans, err := s.ListActivePlans(ctx)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, plans)
	}
}

// CheckoutRequest payload expected by CheckoutHandler.
type CheckoutRequest struct {
	PlanKey    string  `json:"plan_key"`
	ExternalID *string `json:"external_id"`
	Email      string  `json:"email"`
	SuccessURL string  `json:"success_url"`
	CancelURL  string  `json:"cancel_url"`
}

type CheckoutResponse struct {
    ID  string `json:"id"`
    URL string `json:"url"`
}

// PortalRequestDTO payload for PortalHandler.
type PortalRequestDTO struct {
    ExternalID *string `json:"external_id"`
    Email      string  `json:"email"`
    ReturnURL  string  `json:"return_url"`
}

type PortalResponse struct { URL string `json:"url"` }

// CheckoutHandler creates a Stripe Checkout Session for a subscription.
// It only supports subscription mode with inline price data.
func (s *Service) CheckoutHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req CheckoutRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        id, url, err := s.CreateCheckoutSession(r.Context(), req)
        if err != nil {
            if strings.Contains(strings.ToLower(err.Error()), "required") {
                writeErr(w, http.StatusBadRequest, err)
                return
            }
            writeErr(w, http.StatusInternalServerError, err)
            return
        }
        writeJSON(w, http.StatusOK, CheckoutResponse{ID: id, URL: url})
    }
}

// WebhookHandler processes Stripe webhooks to sync subscriptions.
func (s *Service) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		sig := r.Header.Get("Stripe-Signature")
		if s.webhookSecret == "" {
			writeErrMsg(w, http.StatusPreconditionFailed, "webhook secret not configured")
			return
		}
		event, err := webhook.ConstructEvent(payload, sig, s.webhookSecret)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}

		ctx := r.Context()
		switch event.Type {
		case "checkout.session.completed":
			var obj stripesdk.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &obj); err != nil {
				writeErr(w, http.StatusBadRequest, err)
				return
			}
			if obj.Customer == nil || obj.Subscription == nil {
				break
			}
			s.handleSubscriptionSync(ctx, obj.Customer.ID, obj.Subscription.ID)

		case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
			var sub stripesdk.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				writeErr(w, http.StatusBadRequest, err)
				return
			}
			s.syncSubscriptionObject(ctx, &sub)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

// PortalHandler creates a Stripe Billing Portal session and returns its URL.
func (s *Service) PortalHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in PortalRequestDTO
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        url, err := s.CreatePortalSession(r.Context(), PortalRequest{
            ExternalID: in.ExternalID,
            Email:      in.Email,
            ReturnURL:  in.ReturnURL,
        })
        if err != nil {
            if strings.Contains(strings.ToLower(err.Error()), "required") {
                writeErr(w, http.StatusBadRequest, err)
                return
            }
            writeErr(w, http.StatusInternalServerError, err)
            return
        }
        writeJSON(w, http.StatusOK, PortalResponse{URL: url})
    }
}

func (s *Service) handleSubscriptionSync(ctx context.Context, stripeCustomerID, stripeSubscriptionID string) {
	// Fetch current subscription from Stripe and persist.
	// Using Expand is optional; here we keep it simple and rely on customer.subscription.* events as well.
	// Best effort; errors are ignored to keep webhook fast.
	_ = s.syncSubscriptionByIDs(ctx, stripeCustomerID, stripeSubscriptionID)
}

func (s *Service) syncSubscriptionByIDs(ctx context.Context, stripeCustomerID, subID string) error {
	// Fetch sub from Stripe
	sub, err := s.stripe.Subscriptions.Get(subID, nil)
	if err != nil {
		return err
	}
	return s.syncSubscriptionObject(ctx, sub)
}

func (s *Service) syncSubscriptionObject(ctx context.Context, sub *stripesdk.Subscription) error {
	cust, err := s.findCustomerByStripeID(ctx, sub.Customer.ID)
	if err != nil {
		return err
	}
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
	return s.upsertSubscription(ctx, cust.ID, sub.ID, string(sub.Status), ps, pe, sub.CancelAtPeriodEnd, meta)
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func writeErrMsg(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}
