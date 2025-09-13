package constants

// Plan keys used to identify subscription plans
const (
	// PlanFree is the key for the free plan
	PlanFree = "free"

	// PlanHobby is the key for the hobby plan
	PlanHobby = "hobby"

	// PlanStandard is the key for the standard plan
	PlanStandard = "standard"

	// PlanEnterprise is the key for enterprise plans
	PlanEnterprise = "enterprise"
)

// Plan display names
const (
	PlanFreeDisplay       = "Free"
	PlanHobbyDisplay      = "Hobby"
	PlanStandardDisplay   = "Standard"
	PlanEnterpriseDisplay = "Enterprise"
)

// Plan pricing in cents
const (
	PlanFreePrice     = 0
	PlanHobbyPrice    = 4000  // $40.00
	PlanStandardPrice = 15000 // $150.00
)

// Billing intervals
const (
	BillingMonth = "month"
	BillingYear  = "year"
)

// Currency codes
const (
	CurrencyUSD = "usd"
	CurrencyEUR = "eur"
	CurrencyGBP = "gbp"
)
