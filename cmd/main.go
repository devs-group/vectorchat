package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/storage/postgres"
	"github.com/gofiber/swagger"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	_ "github.com/lib/pq"                       // PostgreSQL driver
	_ "github.com/yourusername/vectorchat/docs" // Import generated docs
	swaggerDocs "github.com/yourusername/vectorchat/docs"
	"github.com/yourusername/vectorchat/internal/api"
	"github.com/yourusername/vectorchat/internal/crawler"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/vectorize"
	"github.com/yourusername/vectorchat/pkg/config"
	"github.com/yourusername/vectorchat/pkg/constants"
	"github.com/yourusername/vectorchat/pkg/docprocessor"
	stripe_sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

// @title VectorChat API
// @version 1.0
// @description A Go application that vectorizes text and files into PostgreSQL with pgvector
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @securityDefinitions.oauth2.accessCode OAuth2Application
// @tokenUrl https://github.com/login/oauth/access_token
// @authorizationUrl https://github.com/login/oauth/authorize
// @scope.user:email Grants access to email
// @securityDefinitions.apiCookie CookieAuth
// @in cookie
// @name session_id
func main() {
	app := &cli.Command{
		Name:  "vectorchat",
		Usage: "A Go application that vectorizes text and files into PostgreSQL with pgvector",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run the vectorchat application",
				Action: func(context.Context, *cli.Command) error {
					var appCfg config.AppConfig
					err := config.Load(&appCfg)
					if err != nil {
						return err
					}
					if appCfg.IsSSL {
						swaggerDocs.SwaggerInfo.Schemes = []string{"https"}
					} else {
						swaggerDocs.SwaggerInfo.Schemes = []string{"http"}
					}
					swaggerDocs.SwaggerInfo.Host = appCfg.BaseURL
					return runApplication(&appCfg)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// runApplication starts the vectorchat application
func runApplication(appCfg *config.AppConfig) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Load environment variables
	pgConnStr := appCfg.PGConnection
	if pgConnStr == "" {
		pgConnStr = "postgres://postgres:postgres@localhost:5432/vectordb?sslmode=disable"
	}

	openaiKey := appCfg.OpenAIKey
	if openaiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Wait for PostgreSQL to be ready
	if err := waitForPostgres(pgConnStr); err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	// Run database migrations
	if err := runMigrations(pgConnStr, appCfg.MigrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Initialize database
	pool, err := db.NewDatabase(pgConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	logger.Info("Connected to PostgreSQL database")

	// Initialize Stripe Subscriptions service and migrate its tables
	svc, err := stripe_sub.New(context.Background(), stripe_sub.Config{
		DB:             pool.DB,
		StripeAPIKey:   os.Getenv("STRIPE_API_KEY"),
		WebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		DefaultPlanKey: constants.PlanFree,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Seed default plans idempotently
	for _, p := range defaultPlans() {
		if _, err := svc.UpsertPlan(context.Background(), p); err != nil {
			log.Fatal(err)
		}
	}

	// Initialize vectorizer
	vectorizer := vectorize.NewOpenAIVectorizer(openaiKey)

	// Initialize crawl4ai client (optional)
	webCrawler, err := crawler.NewAPIClient(appCfg.CrawlerAPIURL, nil)
	if err != nil {
		logger.Warn("failed to initialize crawl4ai client; falling back to built-in crawler", "error", err)
	}

	markitdownClient, err := docprocessor.NewMarkitdownClient(appCfg.MarkitdownURL)
	if err != nil {
		return fmt.Errorf("failed to configure markitdown client: %w", err)
	}

	// Initialize repositories
	repos := db.NewRepositories(pool)

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(repos.User, repos.APIKey)
	chatService := services.NewChatService(repos.Chat, repos.Document, repos.File, repos.Message, repos.Revision, vectorizer, webCrawler, markitdownClient, openaiKey, pool, uploadsDir)
	apiKeyService := services.NewAPIKeyService(repos.APIKey)
	commonService := services.NewCommonService()

	// Initialize postgres storage with new config
	sessionStore := postgres.New(postgres.Config{
		ConnectionURI: pgConnStr,
		Table:         "fiber_storage",
		Reset:         false,
		GCInterval:    10 * time.Second,
	})

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionStore, authService, apiKeyService)

	// Initialize ownership middleware
	ownershipMiddleware := middleware.NewOwnershipMiddleware(chatService)

	// Initialize subscription limits middleware
	subscriptionLimits := middleware.NewSubscriptionLimitsMiddleware(svc, chatService)

	// Build redirect URL based on SSL configuration
	redirectURL := fmt.Sprintf("http://%s", appCfg.BaseURL)
	if appCfg.IsSSL {
		redirectURL = fmt.Sprintf("https://%s", appCfg.BaseURL)
	}
	// Initialize OAuth configuration
	oAuthConfig := &api.OAuthConfig{
		GitHubClientID:     appCfg.GithubID,
		GitHubClientSecret: appCfg.GithubSecret,
		RedirectURL:        redirectURL,
		SessionStore:       sessionStore,
	}

	// Set up Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for file uploads
	})

	// Configure CORS with more permissive settings
	frontendURL := fmt.Sprintf("http://%s", appCfg.FrontendURL)
	if appCfg.IsSSL {
		frontendURL = fmt.Sprintf("https://%s", appCfg.FrontendURL)
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     frontendURL,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
		MaxAge:           86400, // 24 hours
	}))

	// Initialize API handlers
	chatbotHandler := api.NewChatHandler(authMiddleware, chatService, uploadsDir, ownershipMiddleware, commonService, subscriptionLimits)
	oAuthHandler := api.NewOAuthHandler(oAuthConfig, authService, authMiddleware)
	apiKeyHandler := api.NewAPIKeyHandler(authService, authMiddleware, apiKeyService, commonService)
	subsHandler := api.NewStripeSubHandler(authMiddleware, svc)
	conversationHandler := api.NewConversationHandler(authMiddleware, chatService)

	// Register routes
	chatbotHandler.RegisterRoutes(app)
	oAuthHandler.RegisterRoutes(app)
	apiKeyHandler.RegisterRoutes(app)
	subsHandler.RegisterRoutes(app)
	conversationHandler.RegisterRoutes(app)

	// Add swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	return app.Listen(":" + port)
}

// defaultPlans returns the requested initial plans seeded on startup.
func defaultPlans() []stripe_sub.PlanParams {
	freeFeatures := map[string]any{
		constants.LimitMessageCredits: 500,
		constants.LimitTrainingData:   "100 KB",
		constants.LimitChatbots:       1,
		constants.LimitDataSources:    "5 data sources (websites, files, texts)",
		constants.LimitEmbedWebsites:  true,
		constants.LimitAPIAccess:      true,
	}
	hobbyFeatures := map[string]any{
		"includes":                    "Everything in Free",
		constants.LimitAdvancedModels: true,
		constants.LimitMessageCredits: 2000,
		constants.LimitTrainingData:   "4 MB",
		constants.LimitChatbots:       3,
		constants.LimitDataSources:    "20 data sources (websites, files, texts)",
		constants.LimitAPIAccess:      true,
		constants.LimitAnalytics:      true,
	}
	standardFeatures := map[string]any{
		"includes":                    "Everything in Hobby",
		constants.LimitMessageCredits: 10000,
		constants.LimitTrainingData:   "33 MB",
		constants.LimitChatbots:       5,
		constants.LimitDataSources:    "50 data sources",
		constants.LimitSeats:          3,
		constants.LimitCustomBranding: true,
		"team_collaboration_tools":    true,
		"priority_email_support":      true,
	}

	return []stripe_sub.PlanParams{
		{
			Key: constants.PlanFree, DisplayName: constants.PlanFreeDisplay, Active: true, BillingInterval: constants.BillingMonth, AmountCents: constants.PlanFreePrice, Currency: constants.CurrencyUSD,
			PlanDefinition: map[string]any{"features": freeFeatures},
		},
		{
			Key: constants.PlanHobby, DisplayName: constants.PlanHobbyDisplay, Active: true, BillingInterval: constants.BillingMonth, AmountCents: constants.PlanHobbyPrice, Currency: constants.CurrencyUSD,
			PlanDefinition: map[string]any{"features": hobbyFeatures},
		},
		{
			Key: constants.PlanStandard, DisplayName: constants.PlanStandardDisplay, Active: true, BillingInterval: constants.BillingMonth, AmountCents: constants.PlanStandardPrice, Currency: constants.CurrencyUSD,
			PlanDefinition: map[string]any{"features": standardFeatures, "tags": []string{"Popular"}},
		},
	}
}

// waitForPostgres attempts to connect to PostgreSQL with retries
func waitForPostgres(connStr string) error {
	maxRetries := 10
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		pool, err := db.NewDatabase(connStr)
		if err == nil {
			pool.Close()
			log.Println("Successfully connected to PostgreSQL")
			return nil
		}

		log.Printf("Failed to connect to PostgreSQL (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("failed to connect to PostgreSQL after %d attempts", maxRetries)
}

// runMigrations executes database migrations using goose
func runMigrations(connStr, migrationsPath string) error {
	if migrationsPath == "" {
		return fmt.Errorf("migrations path is not specified in config")
	}

	// Ensure the path is absolute
	migrationDir := migrationsPath
	if !filepath.IsAbs(migrationsPath) {
		absPath, err := filepath.Abs(migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}
		migrationDir = absPath
	}

	log.Printf("Running migrations from %s", migrationDir)

	// Setup goose and get the database connection
	db, err := setupGoose(connStr, migrationDir)
	if err != nil {
		return err
	}

	// Run up migrations
	if err := goose.Up(db, migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Printf("Migrations completed successfully")
	return nil
}

// setupGoose configures goose to use the specified database and migration directory
func setupGoose(connStr, migrationDir string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database for migrations: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	goose.SetBaseFS(nil)

	// Check if migration directory exists
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migration directory does not exist: %s", migrationDir)
	}

	return db, nil
}
