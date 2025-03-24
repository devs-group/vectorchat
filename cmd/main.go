package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/yourusername/vectorchat/internal/api"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/vectorize"
)

func main() {
	// Load environment variables
	pgConnStr := os.Getenv("PG_CONNECTION_STRING")
	if pgConnStr == "" {
		pgConnStr = "postgres://postgres:postgres@localhost:5432/vectordb?sslmode=disable"
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Wait for PostgreSQL to be ready
	if err := waitForPostgres(pgConnStr); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Initialize database
	pool, err := db.New(pgConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize user store with the same pool
	userStore := db.NewUserStore(pool)
	
	// Initialize chatbot store with the same pool
	chatbotStore := db.NewChatbotStore(pool)

	// Initialize document store
	documentStore := db.NewDocumentStore(pool)

	// Initialize auth middleware
	// authMiddleware := auth.NewAuthMiddleware(session.New(), userStore)
	
	// Initialize vectorizer
	vectorizer := vectorize.NewOpenAIVectorizer(openaiKey)
	
	// Initialize chatbot service
	chatService := services.NewChatService(documentStore, vectorizer, openaiKey, chatbotStore)

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Get GitHub OAuth credentials from environment variables
	githubID := os.Getenv("GITHUB_ID")
	githubSecret := os.Getenv("GITHUB_SECRET")

	if githubID == "" || githubSecret == "" {
		log.Fatal("GITHUB_ID and GITHUB_SECRET environment variables are required")
	}

	// Initialize stores
	sessionStore := session.New()

	// Initialize OAuth configuration
	oAuthConfig := &api.OAuthConfig{
		GitHubClientID:     githubID,
		GitHubClientSecret: githubSecret,
		RedirectURL:        "http://localhost:8080", // Base URL without the callback path
		Store:             sessionStore,
		UserStore:         userStore,
	}

	// Set up Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for file uploads
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize API handlers
	chatbotHandler := api.NewChatHandler(chatService, documentStore, chatbotStore, uploadsDir)
	oAuthHandler := api.NewOAuthHandler(oAuthConfig)
	homeHandler := api.NewHomeHandler(sessionStore, userStore)

	// Register routes
	homeHandler.RegisterRoutes(app)  // Register home routes first
	chatbotHandler.RegisterRoutes(app)
	oAuthHandler.RegisterRoutes(app)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// waitForPostgres attempts to connect to PostgreSQL with retries
func waitForPostgres(connStr string) error {
	maxRetries := 10
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err := db.New(connStr)
		if err == nil {
			db.Close()
			log.Println("Successfully connected to PostgreSQL")
			return nil
		}

		log.Printf("Failed to connect to PostgreSQL (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("failed to connect to PostgreSQL after %d attempts", maxRetries)
}
