package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/yourusername/vectorchat/pkg/api"
	"github.com/yourusername/vectorchat/pkg/chat"
	"github.com/yourusername/vectorchat/pkg/db"
	"github.com/yourusername/vectorchat/pkg/vectorize"
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
	database, err := db.NewChatStore(pgConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize vectorizer
	vectorizer := vectorize.NewOpenAIVectorizer(openaiKey)

	// Initialize chat service
	chatService := chat.NewChatService(database, vectorizer, openaiKey)

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Set up Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for file uploads
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize API handlers
	chatHandler := api.NewChatHandler(chatService, database, uploadsDir)

	// Register routes
	chatHandler.RegisterRoutes(app)

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
		db, err := db.NewChatStore(connStr)
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
