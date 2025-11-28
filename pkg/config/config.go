package config

// AppConfig holds the application configuration.
type AppConfig struct {
	PGConnection       string `env:"PG_CONNECTION_STRING" envRequired:"true"`
	OpenAIKey          string `env:"OPENAI_API_KEY" envRequired:"true"`
	LLMAPIKey          string `env:"LLM_API_KEY" envDefault:""`
	LLMBaseURL         string `env:"LLM_BASE_URL" envDefault:"https://api.openai.com/v1"`
	LLMModelChat       string `env:"LLM_MODEL_CHAT" envDefault:"gpt-4o-mini"`
	LLMModelPromptGen  string `env:"LLM_MODEL_PROMPT_GEN" envDefault:"gpt-4o-mini"`
	BaseURL            string `env:"BASE_URL" envRequired:"true"`
	IsSSL              bool   `env:"IS_SSL" envDefault:"false"`
	MigrationsPath     string `env:"MIGRATIONS_PATH" envRequired:"true"`
	FrontendURL        string `env:"FRONTEND_URL" envRequired:"true"`
	LightFrontendURL   string `env:"LIGHT_FRONTEND_URL" envDefault:"localhost:3100"`
	KratosPublicURL    string `env:"KRATOS_PUBLIC_URL" envDefault:"http://kratos:4433"`
	KratosAdminURL     string `env:"KRATOS_ADMIN_URL" envDefault:"http://kratos:4434"`
	SessionCookieName  string `env:"SESSION_COOKIE_NAME" envDefault:"vectorauth_session"`
	CrawlerAPIURL      string `env:"CRAWLER_API_URL" envDefault:"http://localhost:11235"`
	MarkitdownURL      string `env:"MARKITDOWN_API_URL" envDefault:"http://localhost:8000"`
	HydraAdminURL      string `env:"HYDRA_ADMIN_URL" envDefault:"http://hydra:4445"`
	HydraPublicURL     string `env:"HYDRA_PUBLIC_URL" envDefault:"http://hydra:4444"`
	NATSURL            string `env:"NATS_URL" envDefault:"nats://nats:4222"`
	NATSUsername       string `env:"NATS_USERNAME" envDefault:""`
	NATSPassword       string `env:"NATS_PASSWORD" envDefault:""`
	CrawlWorkerEnabled bool   `env:"CRAWL_WORKER_ENABLED" envDefault:"true"`
}
