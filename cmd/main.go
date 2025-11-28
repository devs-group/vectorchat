package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/robfig/cron/v3"
	_ "github.com/yourusername/vectorchat/docs" // Import generated docs
	swaggerDocs "github.com/yourusername/vectorchat/docs"
	"github.com/yourusername/vectorchat/internal/api"
	"github.com/yourusername/vectorchat/internal/crawler"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/llm"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/queue"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/vectorize"
	"github.com/yourusername/vectorchat/pkg/config"
	"github.com/yourusername/vectorchat/pkg/constants"
	"github.com/yourusername/vectorchat/pkg/docprocessor"
	"github.com/yourusername/vectorchat/pkg/jobs"
	stripe_sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

// @title VectorChat API
// @version 1.0
// @description A Go application that vectorizes text and files into PostgreSQL with pgvector
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @description Provide your token as `Bearer {token}` after generating it via /public/oauth/token.
// @in header
// @name Authorization
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

	llmAPIKey := appCfg.LLMAPIKey
	if llmAPIKey == "" {
		llmAPIKey = openaiKey
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

	llmBaseURL := appCfg.LLMBaseURL
	if llmBaseURL == "" {
		llmBaseURL = "https://api.openai.com/v1"
	}
	defaultChatModel := appCfg.LLMModelChat
	llmClient := llm.NewOpenAIClient(llmAPIKey, llmBaseURL, defaultChatModel, nil)
	fallbackModelIDs := []string{defaultChatModel}
	if appCfg.LLMModelPromptGen != "" {
		fallbackModelIDs = append(fallbackModelIDs, appCfg.LLMModelPromptGen)
	}
	llmService := services.NewLLMService(llmClient, fallbackModelIDs)
	if err := waitForLLM(context.Background(), llmClient, 10, 3*time.Second); err != nil {
		return fmt.Errorf("failed to reach LLM backend: %w", err)
	}

	// Initialize crawl4ai client (optional)
	webCrawler, err := crawler.NewAPIClient(appCfg.CrawlerAPIURL, nil)
	if err != nil {
		logger.Warn("failed to initialize crawl4ai client; falling back to built-in crawler", "error", err)
	}

	markitdownClient, err := docprocessor.NewMarkitdownClient(appCfg.MarkitdownURL)
	if err != nil {
		return fmt.Errorf("failed to configure markitdown client: %w", err)
	}
	processor := docprocessor.NewProcessor(markitdownClient)

	// Initialize repositories
	repos := db.NewRepositories(pool)

	// JetStream for crawl queue (optional but recommended)
	var js nats.JetStreamContext
	nc, err := queue.Connect(appCfg.NATSURL, appCfg.NATSUsername, appCfg.NATSPassword)
	if err != nil {
		logger.Warn("failed to connect to nats; crawl scheduling will degrade", "error", err)
	} else {
		defer nc.Drain()
		if js, err = nc.JetStream(); err != nil {
			logger.Warn("failed to init jetstream", "error", err)
		} else if err := queue.EnsureStreams(js); err != nil {
			logger.Warn("failed to ensure jetstream streams", "error", err)
		}
	}

	// Initialize services
	kbService := services.NewKnowledgeBaseService(repos.File, repos.Document, vectorizer, processor, webCrawler, pool)
	sharedKBService := services.NewSharedKnowledgeBaseService(repos.SharedKB, repos.File, repos.Document, kbService)
	hydraService := services.NewHydraService(appCfg.HydraAdminURL, appCfg.HydraPublicURL)
	logger.Info("hydra configuration", "admin_url", appCfg.HydraAdminURL, "public_url", appCfg.HydraPublicURL)
	authService := services.NewAuthService(repos.User)
	chatService := services.NewChatService(repos.Chat, repos.SharedKB, repos.Document, repos.File, repos.Message, repos.Revision, repos.LLMUsage, vectorizer, kbService, llmClient, pool, defaultChatModel)
	apiKeyService := services.NewAPIKeyService(hydraService)
	commonService := services.NewCommonService()
	scheduleService := services.NewCrawlScheduleService(repos.Schedule, kbService, js)
	promptService := services.NewPromptService(llmClient, appCfg.LLMModelPromptGen)

	// Validate that cron library supports our expressions (minute granularity) once at startup
	if _, err := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse("*/5 * * * *"); err != nil {
		logger.Warn("cron parser init failed", "error", err)
	}

	// Start embedded crawl worker if enabled and JetStream available
	if appCfg.CrawlWorkerEnabled && js != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go startCrawlWorker(ctx, js, repos.Schedule, kbService, logger)
	}

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, hydraService)

	// Initialize ownership middleware
	ownershipMiddleware := middleware.NewOwnershipMiddleware(chatService)

	// Initialize subscription limits middleware
	subscriptionLimits := middleware.NewSubscriptionLimitsMiddleware(svc, chatService)

	// Set up Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for file uploads
	})

	app.Use(fiberLogger.New())

	// Initialize API handlers
	chatbotHandler := api.NewChatHandler(authMiddleware, chatService, ownershipMiddleware, commonService, subscriptionLimits, scheduleService, promptService)
	sharedKnowledgeBaseHandler := api.NewSharedKnowledgeBaseHandler(authMiddleware, sharedKBService, scheduleService)
	authHandler := api.NewAuthHandler(authService, authMiddleware, api.AuthConfig{
		KratosPublicURL: appCfg.KratosPublicURL,
		KratosAdminURL:  appCfg.KratosAdminURL,
		SessionCookie:   appCfg.SessionCookieName,
	})
	apiKeyHandler := api.NewAPIKeyHandler(authService, authMiddleware, apiKeyService, commonService)
	subsHandler := api.NewStripeSubHandler(authMiddleware, svc)
	conversationHandler := api.NewConversationHandler(authMiddleware, chatService)
	widgetHandler := api.NewWidgetHandler(authMiddleware)
	queueHandler := api.NewQueueHandler(authMiddleware, services.NewQueueMetricsService(js))
	llmHandler := api.NewLLMHandler(authMiddleware, llmService, svc)

	// Register routes
	chatbotHandler.RegisterRoutes(app)
	sharedKnowledgeBaseHandler.RegisterRoutes(app)
	authHandler.RegisterRoutes(app)
	apiKeyHandler.RegisterRoutes(app)
	subsHandler.RegisterRoutes(app)
	conversationHandler.RegisterRoutes(app)
	widgetHandler.RegisterRoutes(app)
	queueHandler.RegisterRoutes(app)
	llmHandler.RegisterRoutes(app)

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

func startCrawlWorker(ctx context.Context, js nats.JetStreamContext, scheduleRepo *db.CrawlScheduleRepository, kbService *services.KnowledgeBaseService, logger *slog.Logger) {
	sub, err := js.PullSubscribe(
		jobs.CrawlSubject,
		"crawler-workers",
		nats.BindStream(jobs.CrawlStream),
		nats.ManualAck(),
	)
	if err != nil {
		logger.Warn("crawl worker: failed to subscribe", "error", err)
		return
	}
	logger.Info("crawl worker started (embedded)")

	for {
		select {
		case <-ctx.Done():
			logger.Info("crawl worker stopping")
			return
		default:
		}

		msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			logger.Warn("crawl worker: fetch error", "error", err)
			continue
		}

		for _, msg := range msgs {
			handleCrawlMessage(ctx, msg, kbService, scheduleRepo, logger)
		}
	}
}

func handleCrawlMessage(ctx context.Context, msg *nats.Msg, kb *services.KnowledgeBaseService, repo *db.CrawlScheduleRepository, logger *slog.Logger) {
	var payload jobs.CrawlJobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		logger.Warn("crawl worker: invalid payload", "error", err)
		_ = msg.Term()
		return
	}

	target := services.KnowledgeBaseTarget{
		ChatbotID:             payload.ChatbotID,
		SharedKnowledgeBaseID: payload.SharedKnowledgeBaseID,
	}

	start := time.Now().UTC()
	if _, err := kb.IngestWebsite(ctx, target, payload.RootURL); err != nil {
		if payload.ScheduleID != uuid.Nil {
			errMsg := err.Error()
			status := "failed"
			_ = repo.UpdateRunInfo(context.Background(), payload.ScheduleID, &start, nil, &status, &errMsg)
		}
		logger.Error("crawl worker: crawl failed", "schedule_id", payload.ScheduleID, "error", err)
		_ = msg.Nak()
		return
	}

	if payload.ScheduleID != uuid.Nil {
		status := "completed"
		_ = repo.UpdateRunInfo(context.Background(), payload.ScheduleID, &start, nil, &status, nil)
	}
	_ = msg.Ack()
	logger.Info("crawl worker: crawl completed", "schedule_id", payload.ScheduleID, "url", payload.RootURL)
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

func waitForLLM(ctx context.Context, client llm.Client, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		if _, err := client.ListModels(ctx); err == nil {
			return nil
		} else {
			log.Printf("LLM not ready (attempt %d/%d): %v", i+1, retries, err)
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("LLM not ready after %d attempts", retries)
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
