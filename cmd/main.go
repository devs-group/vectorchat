package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/storage/postgres"
	"github.com/gofiber/swagger"
	"github.com/urfave/cli/v3"

	_ "github.com/yourusername/vectorchat/docs" // Import generated docs
	"github.com/yourusername/vectorchat/internal/api"
	"github.com/yourusername/vectorchat/internal/config"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/store"
	"github.com/yourusername/vectorchat/internal/vectorize"
)

// @title VectorChat API
// @version 1.0
// @description A Go application that vectorizes text and files into PostgreSQL with pgvector
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @securityDefinitions.oauth2.accessCode OAuth2Application
// @tokenUrl https://github.com/login/oauth/access_token
// @authorizationUrl https://github.com/login/oauth/authorize
// @scope.user:email Grants access to email
func main() {
	app := &cli.Command{
		Name:  "vectorchat",
		Usage: "A Go application that vectorizes text and files into PostgreSQL with pgvector",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run the vectorchat application",
				Action: func(ctx context.Context) error {
					return runApplication()
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// runApplication starts the vectorchat application
func runApplication() error {
	var appCfg config.AppConfig
	err := config.Load(&appCfg)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

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

	// Initialize database
	pool, err := store.New(pgConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize user store with the same pool
	userStore := store.NewUserStore(pool)

	// Initialize chatbot store with the same pool
	chatbotStore := store.NewChatbotStore(pool)

	// Initialize document store
	documentStore := store.NewDocumentStore(pool)

	// Initialize vectorizer
	vectorizer := vectorize.NewOpenAIVectorizer(openaiKey)

	// Initialize chatbot service
	chatService := services.NewChatService(documentStore, vectorizer, openaiKey, chatbotStore)

	// Initialize api key service
	apiKeyService := services.NewAPIKeyService()

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %v", err)
	}

	// Initialize postgres sotrage with new config
	sessionStore := postgres.New(postgres.Config{
		ConnectionURI: pgConnStr,
		Table:         "fiber_storage",
		Reset:         false,
		GCInterval:    10 * time.Second,
	})

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionStore, userStore, apiKeyService)

	// Initialize ownership middleware
	ownershipMiddleware := middleware.NewOwnershipMiddleware(chatbotStore)

	// Initialize OAuth configuration
	oAuthConfig := &api.OAuthConfig{
		GitHubClientID:     appCfg.GithubID,
		GitHubClientSecret: appCfg.GithubSecret,
		RedirectURL:        appCfg.BaseURL,
		SessionStore:       sessionStore,
	}

	// Set up Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for file uploads
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize API handlers
	chatbotHandler := api.NewChatHandler(authMiddleware, chatService, documentStore, chatbotStore, uploadsDir, ownershipMiddleware)
	oAuthHandler := api.NewOAuthHandler(oAuthConfig, userStore, authMiddleware, apiKeyService)
	homeHandler := api.NewHomeHandler(sessionStore, userStore, authMiddleware)

	// Register routes
	homeHandler.RegisterRoutes(app) // Register home routes first
	chatbotHandler.RegisterRoutes(app)
	oAuthHandler.RegisterRoutes(app)

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

// waitForPostgres attempts to connect to PostgreSQL with retries
func waitForPostgres(connStr string) error {
	maxRetries := 10
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		store, err := store.New(connStr)
		if err == nil {
			store.Close()
			log.Println("Successfully connected to PostgreSQL")
			return nil
		}

		log.Printf("Failed to connect to PostgreSQL (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("failed to connect to PostgreSQL after %d attempts", maxRetries)
}
