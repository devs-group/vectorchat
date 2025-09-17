package config

// AppConfig holds the application configuration.
type AppConfig struct {
	GithubID       string `env:"GITHUB_ID" envRequired:"true"`
	GithubSecret   string `env:"GITHUB_SECRET" envRequired:"true"`
	PGConnection   string `env:"PG_CONNECTION_STRING" envRequired:"true"`
	OpenAIKey      string `env:"OPENAI_API_KEY" envRequired:"true"`
	BaseURL        string `env:"BASE_URL" envRequired:"true"`
	IsSSL          bool   `env:"IS_SSL" envDefault:"false"`
	MigrationsPath string `env:"MIGRATIONS_PATH" envRequired:"true"`
	FrontendURL    string `env:"FRONTEND_URL" envRequired:"true"`
	CrawlerAPIURL  string `env:"CRAWLER_API_URL" envDefault:"http://localhost:11235"`
	MarkitdownURL  string `env:"MARKITDOWN_API_URL" envDefault:"http://localhost:8000"`
}
