package stripe_sub

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

func (c *service) GetPlan(ctx context.Context, planKey string) (Plan, error) {
	var row struct {
		ID, Key, DisplayName, BillingInterval, Currency string
		Active                                          bool
		AmountCents                                     int64
		Metadata                                        []byte
		DefinitionRaw                                   []byte `db:"plan_definition"`
		CreatedAt, UpdatedAt                            time.Time
	}
	err := c.db.GetContext(ctx, &row, `SELECT id, key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition, created_at, updated_at FROM stripe_sub_pkg_plans WHERE key=$1 AND active=true`, planKey)
	if err != nil {
		return Plan{}, ErrNotFound("plan not found")
	}
	var def PlanDefinition
	if err := json.Unmarshal(row.DefinitionRaw, &def); err != nil {
		return Plan{}, ErrInternal("invalid plan_definition JSON")
	}
	if err := validatePlan(row.Currency, row.AmountCents, def); err != nil {
		return Plan{}, ErrValidation(err.Error())
	}
	p := Plan{ID: row.ID, Key: row.Key, DisplayName: row.DisplayName, Active: row.Active, BillingInterval: row.BillingInterval, AmountCents: row.AmountCents, Currency: row.Currency, Metadata: row.Metadata, Definition: def, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	return p, nil
}

func (c *service) ListActivePlans(ctx context.Context) ([]Plan, error) {
	rows := []struct {
		ID, Key, DisplayName, BillingInterval, Currency string
		Active                                          bool
		AmountCents                                     int64
		Metadata                                        []byte
		DefinitionRaw                                   []byte `db:"plan_definition"`
		CreatedAt, UpdatedAt                            time.Time
	}{}
	if err := c.db.SelectContext(ctx, &rows, `SELECT id, key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition, created_at, updated_at FROM stripe_sub_pkg_plans WHERE active=true`); err != nil {
		return nil, err
	}
	res := make([]Plan, 0, len(rows))
	for _, r := range rows {
		var def PlanDefinition
		if err := json.Unmarshal(r.DefinitionRaw, &def); err != nil {
			return nil, ErrInternal("invalid plan_definition JSON")
		}
		if err := validatePlan(r.Currency, r.AmountCents, def); err != nil {
			return nil, ErrValidation(err.Error())
		}
		res = append(res, Plan{ID: r.ID, Key: r.Key, DisplayName: r.DisplayName, Active: r.Active, BillingInterval: r.BillingInterval, AmountCents: r.AmountCents, Currency: r.Currency, Metadata: r.Metadata, Definition: def, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt})
	}
	return res, nil
}

// GetPlans returns the active plans. Alias of ListActivePlans for convenience.
func (c *service) GetPlans(ctx context.Context) ([]Plan, error) { return c.ListActivePlans(ctx) }

func validatePlan(currency string, amount int64, def PlanDefinition) error {
	if def.StripePriceID == "" {
		return errors.New("plan_definition.stripe_price_id is required")
	}
	if currency == "" {
		return errors.New("currency is required")
	}
	if amount <= 0 && def.StripePriceID != "price_free" {
		return errors.New("amount_cents must be > 0")
	}
	return nil
}

type PlanSpec struct {
	Key             string
	DisplayName     string
	Active          bool
	BillingInterval string
	AmountCents     int64
	Currency        string
	Metadata        map[string]string
	Definition      PlanDefinition
}

func (c *service) EnsurePlans(ctx context.Context, specs []PlanSpec) error {
	for _, sp := range specs {
		if sp.Key == "" {
			return ErrValidation("plan key required")
		}
		if err := validatePlan(sp.Currency, sp.AmountCents, sp.Definition); err != nil {
			return ErrValidation(err.Error())
		}
		defJSON, _ := json.Marshal(sp.Definition)
		var metaJSON []byte
		if len(sp.Metadata) > 0 {
			metaJSON, _ = json.Marshal(sp.Metadata)
		}
		_, err := c.db.ExecContext(ctx, `INSERT INTO stripe_sub_pkg_plans (key, display_name, active, billing_interval, amount_cents, currency, metadata, plan_definition)
            VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8::jsonb)
            ON CONFLICT (key) DO UPDATE SET
              display_name=EXCLUDED.display_name,
              active=EXCLUDED.active,
              billing_interval=EXCLUDED.billing_interval,
              amount_cents=EXCLUDED.amount_cents,
              currency=EXCLUDED.currency,
              metadata=COALESCE(EXCLUDED.metadata, stripe_sub_pkg_plans.metadata),
              plan_definition=EXCLUDED.plan_definition,
              updated_at=now()`, sp.Key, sp.DisplayName, sp.Active, sp.BillingInterval, sp.AmountCents, sp.Currency, nullBytes(metaJSON), string(defJSON))
		if err != nil {
			return err
		}
	}
	return nil
}
