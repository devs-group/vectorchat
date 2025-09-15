package constants

// Subscription limit keys used for checking plan restrictions
const (
	// LimitChatbots defines the maximum number of chatbots a user can create
	LimitChatbots = "chatbots"

	// LimitDataSources defines the maximum number of data sources (files/websites) per chatbot
	LimitDataSources = "limit_to_train_on"

	// LimitTrainingData defines the maximum training data size per chatbot
	LimitTrainingData = "training_data_per_chatbot"

	// LimitMessageCredits defines the monthly message credits
	LimitMessageCredits = "message_credits_per_month"

	// LimitAPIAccess defines whether the user has API access
	LimitAPIAccess = "api_access"

	// LimitEmbedWebsites defines whether the user can embed on unlimited websites
	LimitEmbedWebsites = "embed_on_unlimited_websites"

	// LimitSeats defines the number of team seats (for team plans)
	LimitSeats = "seats"

	// LimitCustomBranding defines whether custom branding is available
	LimitCustomBranding = "custom_branding"

	// LimitAnalytics defines whether analytics are available
	LimitAnalytics = "basic_analytics"

	// LimitAdvancedModels defines access to advanced AI models
	LimitAdvancedModels = "access_to_advanced_models"
)

// Default limit values for free tier
const (
	DefaultChatbots       = 1
	DefaultDataSources    = 5
	DefaultTrainingData   = "400 KB"
	DefaultMessageCredits = 100
	DefaultInactivityDays = 14
)
