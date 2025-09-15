package constants

// Plan keys used to identify subscription plans
const (
	// PlanFree is the key for the free plan
	PlanFree = "free"

	// PlanHobby is the key for the hobby plan
	PlanHobby = "hobby"

	// PlanStandard is the key for the standard plan
	PlanStandard = "standard"
)

// Plan display names
const (
	PlanFreeDisplay     = "Free"
	PlanHobbyDisplay    = "Hobby"
	PlanStandardDisplay = "Standard"
)

// Plan pricing in cents
const (
	PlanFreePrice     = 0
	PlanHobbyPrice    = 1500 // $15.00
	PlanStandardPrice = 4000 // $40.00
)

// Billing intervals
const (
	BillingMonth = "month"
)

// Currency codes
const (
	CurrencyUSD = "usd"
)
