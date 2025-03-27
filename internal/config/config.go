package config

// AppConfig holds the application configuration.
type AppConfig struct {
	GithubID     string `env:"GITHUB_ID"`
	GithubSecret string `env:"GITHUB_SECRET"`
	PGConnection string `env:"PG_CONNECTION_STRING"`
	OpenAIKey    string `env:"OPENAI_API_KEY"`
}
