package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/vectorize"
)

// ServiceContainer holds all services and their dependencies
type ServiceContainer struct {
	DB           *Database
	Repositories *Repositories
	Services     *Services
}

// Services holds all business logic services
type Services struct {
	Auth   *AuthService
	APIKey *APIKeyService
	Chat   *ChatService
	Home   *HomeService
}

// NewServiceContainer creates a new service container with all dependencies
func NewServiceContainer(connStr string, vectorizer vectorize.Vectorizer, openaiKey string) (*ServiceContainer, error) {
	// Initialize database
	db, err := NewDatabase(connStr)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	repos := NewRepositories(db)
	reposTx := NewRepositoriesTx(db)

	// Initialize services with dependency injection
	services := &Services{
		Auth:   NewAuthService(repos.User, repos.APIKey),
		APIKey: NewAPIKeyService(repos.APIKey),
		Chat:   NewChatService(reposTx.Chatbot, reposTx.Document, reposTx.File, vectorizer, openaiKey, db),
		Home:   NewHomeService(repos.User),
	}

	return &ServiceContainer{
		DB:           db,
		Repositories: repos,
		Services:     services,
	}, nil
}

// Close closes the database connection
func (sc *ServiceContainer) Close() error {
	return sc.DB.Close()
}

// Example usage functions below demonstrate how to use the new service layer

// ExampleCreateUser demonstrates creating a new user
func ExampleCreateUser(ctx context.Context, container *ServiceContainer) {
	user, err := container.Services.Auth.CreateUser(
		ctx,
		"", // Empty ID will generate a new UUID
		"John Doe",
		"john@example.com",
		"oauth2",
	)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return
	}

	log.Printf("Created user: %s (%s)", user.Name, user.ID)
}

// ExampleCreateAPIKey demonstrates creating an API key for a user
func ExampleCreateAPIKey(ctx context.Context, container *ServiceContainer, userID string) {
	apiKeyResponse, plainTextKey, err := container.Services.APIKey.CreateAPIKey(
		ctx,
		userID,
		"My API Key",
		nil, // No expiration
	)
	if err != nil {
		log.Printf("Error creating API key: %v", err)
		return
	}

	log.Printf("Created API key: %s", apiKeyResponse.ID)
	log.Printf("Plain text key (save this!): %s", plainTextKey)
}

// ExampleCreateChatbot demonstrates creating a new chatbot
func ExampleCreateChatbot(ctx context.Context, container *ServiceContainer, userID string) {
	chatbot, err := container.Services.Chat.CreateChatbot(
		ctx,
		userID,
		"My Assistant",
		"A helpful AI assistant",
		"You are a knowledgeable assistant that helps with questions.",
	)
	if err != nil {
		log.Printf("Error creating chatbot: %v", err)
		return
	}

	log.Printf("Created chatbot: %s (%s)", chatbot.Name, chatbot.ID)
}

// ExampleListChatbots demonstrates listing user's chatbots with pagination
func ExampleListChatbots(ctx context.Context, container *ServiceContainer, userID string) {
	response, err := container.Services.Chat.ListChatbotsWithPagination(
		ctx,
		userID,
		0,  // offset
		10, // limit
	)
	if err != nil {
		log.Printf("Error listing chatbots: %v", err)
		return
	}

	log.Printf("Found %d chatbots (total: %d)", len(response.Data.([]*Chatbot)), response.Total)
}

// ExampleValidateAPIKey demonstrates validating an API key
func ExampleValidateAPIKey(ctx context.Context, container *ServiceContainer, plainTextKey string) {
	userID, err := container.Services.APIKey.ValidateAPIKey(ctx, plainTextKey)
	if err != nil {
		log.Printf("Error validating API key: %v", err)
		return
	}

	log.Printf("API key is valid for user: %s", userID)
}

// ExampleAddFileToChat demonstrates adding a file to a chatbot
func ExampleAddFileToChat(ctx context.Context, container *ServiceContainer, chatbotID, filePath string) {
	err := container.Services.Chat.AddFile(
		ctx,
		uuid.New().String(), // document ID
		filePath,
		uuid.MustParse(chatbotID),
	)
	if err != nil {
		log.Printf("Error adding file to chatbot: %v", err)
		return
	}

	log.Printf("Successfully added file to chatbot: %s", filePath)
}

// ExampleChatWithBot demonstrates chatting with a chatbot
func ExampleChatWithBot(ctx context.Context, container *ServiceContainer, chatbotID, userID, message string) {
	response, err := container.Services.Chat.ChatWithChatbot(
		ctx,
		chatbotID,
		userID,
		message,
	)
	if err != nil {
		log.Printf("Error chatting with bot: %v", err)
		return
	}

	log.Printf("Bot response: %s", response)
}

// ExampleGetAPIKeysWithPagination demonstrates getting API keys with pagination
func ExampleGetAPIKeysWithPagination(ctx context.Context, container *ServiceContainer, userID string) {
	response, err := container.Services.APIKey.GetAPIKeysWithPagination(
		ctx,
		userID,
		0,  // offset
		10, // limit
	)
	if err != nil {
		log.Printf("Error getting API keys: %v", err)
		return
	}

	apiKeys := response.Data.([]*APIKeyResponse)
	log.Printf("Found %d API keys (total: %d)", len(apiKeys), response.Total)
}

// ExampleGetChatbotStats demonstrates getting chatbot statistics
func ExampleGetChatbotStats(ctx context.Context, container *ServiceContainer, chatbotID, userID string) {
	stats, err := container.Services.Chat.GetChatbotStats(
		ctx,
		uuid.MustParse(chatbotID),
		userID,
	)
	if err != nil {
		log.Printf("Error getting chatbot stats: %v", err)
		return
	}

	log.Printf("Chatbot stats: %+v", stats)
}

// ExampleTransactionUsage demonstrates how to use transactions for complex operations
func ExampleTransactionUsage(ctx context.Context, container *ServiceContainer) {
	tx, err := container.DB.BeginTx(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}
	defer tx.Rollback()

	// Example: Create user and API key in a single transaction
	user := &User{
		ID:        uuid.New().String(),
		Name:      "Transaction User",
		Email:     "tx@example.com",
		Provider:  "oauth2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Use transaction-aware repository methods
	userRepo := container.Repositories.User.(UserRepositoryTx)
	err = userRepo.CreateTx(ctx, tx, user)
	if err != nil {
		log.Printf("Error creating user in transaction: %v", err)
		return
	}

	apiKey := &APIKey{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Key:       "hashed_key_value",
		Name:      stringPtr("Transaction API Key"),
		CreatedAt: time.Now(),
	}

	apiKeyRepo := container.Repositories.APIKey.(APIKeyRepositoryTx)
	err = apiKeyRepo.CreateTx(ctx, tx, apiKey)
	if err != nil {
		log.Printf("Error creating API key in transaction: %v", err)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	log.Printf("Successfully created user and API key in transaction")
}

// ExampleErrorHandling demonstrates proper error handling
func ExampleErrorHandling(ctx context.Context, container *ServiceContainer) {
	// Try to find a non-existent user
	_, err := container.Services.Auth.FindUserByID(ctx, "non-existent-id")
	if err != nil {
		// Check for specific error types
		if container.Repositories.User != nil {
			switch {
			case IsNoRowsError(err):
				log.Printf("User not found (no rows error)")
			default:
				log.Printf("Database error: %v", err)
			}
		}
	}
}

// ExampleCleanup demonstrates proper cleanup
func ExampleCleanup(container *ServiceContainer) {
	// Close database connections
	if err := container.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

// Helper functions

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// Main example function that shows complete usage
func ExampleCompleteUsage() {
	ctx := context.Background()

	// Initialize the service container
	container, err := NewServiceContainer(
		"postgres://user:pass@localhost/vectorchat?sslmode=disable",
		nil, // vectorizer would be initialized separately
		"your-openai-key",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer container.Close()

	// Example workflow
	log.Println("=== Creating User ===")
	ExampleCreateUser(ctx, container)

	log.Println("=== Creating API Key ===")
	ExampleCreateAPIKey(ctx, container, "user-id")

	log.Println("=== Creating Chatbot ===")
	ExampleCreateChatbot(ctx, container, "user-id")

	log.Println("=== Listing Chatbots ===")
	ExampleListChatbots(ctx, container, "user-id")

	log.Println("=== Transaction Example ===")
	ExampleTransactionUsage(ctx, container)

	log.Println("=== Error Handling Example ===")
	ExampleErrorHandling(ctx, container)
}
