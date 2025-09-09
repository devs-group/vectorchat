package stripe_sub

import (
    "context"
    "encoding/json"
    "io"
    "log"
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
            log.Printf("stripe webhook: read body error: %v", err)
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        sig := r.Header.Get("Stripe-Signature")
        if s.webhookSecret == "" {
            log.Printf("stripe webhook: missing webhook secret (Stripe-Signature header=%q)", sig)
            writeErrMsg(w, http.StatusPreconditionFailed, "webhook secret not configured")
            return
        }
        // Verify signatures, but ignore API version mismatch to be resilient across endpoint versions.
        event, err := webhook.ConstructEventWithOptions(payload, sig, s.webhookSecret, webhook.ConstructEventOptions{
            IgnoreAPIVersionMismatch: true,
        })
        if err != nil {
            // Log exact signature verification error for debugging
            log.Printf("stripe webhook: signature verification failed: %v (Stripe-Signature=%q, payload_len=%d)", err, sig, len(payload))
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

// --- Generic identity-based handlers (framework-agnostic) ---

// Identity represents the authenticated user identity needed to create sessions.
type Identity struct {
    Email      string
    ExternalID *string
}

// IdentityProvider extracts Identity from the incoming request (caller may ignore r and capture values in a closure).
type IdentityProvider func(r *http.Request) (Identity, error)

// AppCheckoutRequest is the minimal client payload when using an authenticated identity.
type AppCheckoutRequest struct {
    CustomerID string `json:"customer_id"`
    PlanKey    string `json:"plan_key"`
    SuccessURL string `json:"success_url"`
    CancelURL  string `json:"cancel_url"`
}

// CheckoutAuthedHandler creates a Checkout Session using IdentityProvider for email and optional external ID.
// The request body should contain plan_key, success_url, cancel_url; optionally customer_id to override external_id.
func (s *Service) CheckoutAuthedHandler(idp IdentityProvider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in AppCheckoutRequest
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        ident, err := idp(r)
        if err != nil {
            writeErr(w, http.StatusUnauthorized, err)
            return
        }
        ext := ident.ExternalID
        if strings.TrimSpace(in.CustomerID) != "" {
            // Allow client to supply app-specific external reference
            v := in.CustomerID
            ext = &v
        }
        id, url, err := s.CreateCheckoutSession(r.Context(), CheckoutRequest{
            PlanKey:    in.PlanKey,
            ExternalID: ext,
            Email:      ident.Email,
            SuccessURL: in.SuccessURL,
            CancelURL:  in.CancelURL,
        })
        if err != nil {
            if err == ErrActiveSubscription {
                writeErr(w, http.StatusConflict, err)
                return
            }
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        writeJSON(w, http.StatusOK, CheckoutResponse{ID: id, URL: url})
    }
}

// CheckoutAuthedHandlerFor is a convenience wrapper when you already have identity values.
func (s *Service) CheckoutAuthedHandlerFor(email string, externalID *string) http.HandlerFunc {
    return s.CheckoutAuthedHandler(func(_ *http.Request) (Identity, error) {
        return Identity{Email: email, ExternalID: externalID}, nil
    })
}

// SubscriptionHandler returns the user's current subscription (optionally refreshing first).
type SubscriptionHandlerOptions struct {
    IdentityProvider IdentityProvider
    RefreshFirst     bool
    EnsurePlanKey    bool
}

func (s *Service) SubscriptionHandler(opts SubscriptionHandlerOptions) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if opts.IdentityProvider == nil { writeErrMsg(w, http.StatusInternalServerError, "missing identity provider"); return }
        ident, err := opts.IdentityProvider(r)
        if err != nil { writeErr(w, http.StatusUnauthorized, err); return }
        if opts.RefreshFirst {
            _, _ = s.RefreshLatestSubscription(r.Context(), ident.ExternalID, ident.Email)
        }
        subRec, err := s.GetUserCurrentSubscription(r.Context(), ident.ExternalID, ident.Email)
        if err != nil { writeErr(w, http.StatusInternalServerError, err); return }
        if subRec == nil { writeJSON(w, http.StatusOK, map[string]any{"subscription": nil}); return }
        if opts.EnsurePlanKey && (subRec.Metadata == nil || subRec.Metadata["plan_key"] == nil) {
            if key, err := s.EnsureSubscriptionPlanKey(r.Context(), subRec.StripeSubscription); err == nil && key != "" {
                if subRec.Metadata == nil { subRec.Metadata = map[string]any{} }
                subRec.Metadata["plan_key"] = key
            }
        }
        writeJSON(w, http.StatusOK, map[string]any{"subscription": subRec})
    }
}

// SubscriptionHandlerFor is a convenience wrapper when you already have identity values.
func (s *Service) SubscriptionHandlerFor(email string, externalID *string, refreshFirst, ensurePlanKey bool) http.HandlerFunc {
    return s.SubscriptionHandler(SubscriptionHandlerOptions{
        IdentityProvider: func(_ *http.Request) (Identity, error) { return Identity{Email: email, ExternalID: externalID}, nil },
        RefreshFirst:     refreshFirst,
        EnsurePlanKey:    ensurePlanKey,
    })
}

// PortalAuthedHandler creates a billing portal session using IdentityProvider for email and ext id.
func (s *Service) PortalAuthedHandler(idp IdentityProvider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in struct{ ReturnURL string `json:"return_url"` }
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil { writeErr(w, http.StatusBadRequest, err); return }
        if strings.TrimSpace(in.ReturnURL) == "" { writeErrMsg(w, http.StatusBadRequest, "missing return_url"); return }
        ident, err := idp(r)
        if err != nil { writeErr(w, http.StatusUnauthorized, err); return }
        url, err := s.CreatePortalSession(r.Context(), PortalRequest{ ExternalID: ident.ExternalID, Email: ident.Email, ReturnURL: in.ReturnURL })
        if err != nil { writeErr(w, http.StatusBadRequest, err); return }
        writeJSON(w, http.StatusOK, map[string]any{"url": url})
    }
}

// PortalAuthedHandlerFor is a convenience wrapper when you already have identity values.
func (s *Service) PortalAuthedHandlerFor(email string, externalID *string) http.HandlerFunc {
    return s.PortalAuthedHandler(func(_ *http.Request) (Identity, error) {
        return Identity{Email: email, ExternalID: externalID}, nil
    })
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
        for k, v := range sub.Metadata { meta[k] = v }
    }
    if meta == nil || meta["plan_key"] == nil {
        if key, _ := s.inferPlanKeyFromStripeSub(sub); key != "" {
            if meta == nil { meta = map[string]any{} }
            meta["plan_key"] = key
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
