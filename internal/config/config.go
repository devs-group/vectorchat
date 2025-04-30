package config

// AppConfig holds the application configuration.
type AppConfig struct {
	GithubID     string `env:"GITHUB_ID" envRequired:"true"`
	GithubSecret string `env:"GITHUB_SECRET" envRequired:"true"`
	PGConnection string `env:"PG_CONNECTION_STRING" envRequired:"true"`
	OpenAIKey    string `env:"OPENAI_API_KEY" envRequired:"true"`
	BaseURL      string `env:"BASE_URL" envRequired:"true"`
}
