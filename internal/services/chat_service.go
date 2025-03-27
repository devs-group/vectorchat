package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/store"
	"github.com/yourusername/vectorchat/internal/vectorize"
)

// ChatService handles chat interactions with context from vector database
type ChatService struct {
	documentStore *store.DocumentStore
	vectorizer    vectorize.Vectorizer
	openaiKey     string
	chatbotStore  *store.ChatbotStore
}

// NewChatService creates a new chat service
func NewChatService(documentStore *store.DocumentStore, vectorizer vectorize.Vectorizer, openaiKey string, chatbotStore *store.ChatbotStore) *ChatService {
	return &ChatService{
		documentStore: documentStore,
		vectorizer:    vectorizer,
		openaiKey:     openaiKey,
		chatbotStore:  chatbotStore,
	}
}

// AddDocument adds a document to the vector database
func (c *ChatService) AddDocument(ctx context.Context, id string, content string, chatbotID string) error {
	embedding, err := c.vectorizer.VectorizeText(ctx, content)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize document")
	}

	doc := store.Document{
		ID:        id,
		Content:   content,
		Embedding: embedding,
		ChatbotID: uuid.MustParse(chatbotID),
	}

	return c.documentStore.StoreDocument(ctx, doc)
}

// AddFile adds a file to the vector database
func (c *ChatService) AddFile(ctx context.Context, id string, filePath string, chatbotID uuid.UUID) error {
	embedding, err := c.vectorizer.VectorizeFile(ctx, filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize file")
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to read file")
	}

	doc := store.Document{
		ID:        id,
		Content:   string(content),
		Embedding: embedding,
		ChatbotID: chatbotID,
	}

	return c.documentStore.StoreDocument(ctx, doc)
}

// Chat sends a message to the LLM with relevant context from the vector database
func (c *ChatService) Chat(ctx context.Context, userID string, message string) (string, error) {
	return c.ChatWithID(ctx, "default", userID, message)
}

// ChatWithID sends a message to the LLM with relevant context from the specified chat session
func (c *ChatService) ChatWithID(ctx context.Context, chatID string, userID string, message string) (string, error) {
	// Check if chatID is actually a chatbot ID (uuid format)
	if len(chatID) == 36 { // Simple UUID format check
		// Try to find the chatbot (will fail if not a valid chatbot ID)
		chatbot, err := c.chatbotStore.FindChatbotByIDAndUserID(ctx, uuid.MustParse(chatID), userID)
		if err == nil && chatbot != nil {
			// If successful, use ChatWithChatbot instead
			return c.ChatWithChatbot(ctx, chatID, chatbot.UserID, message)
		}
	}

	// Vectorize the query
	queryEmbedding, err := c.vectorizer.VectorizeText(ctx, message)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Find relevant documents
	docs, err := c.documentStore.FindSimilarDocumentsByChatID(ctx, queryEmbedding, chatID, 3)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find similar documents: %v", err)
	}

	// Check if any documents were found for this chat ID
	if len(docs) == 0 {
		return "", apperrors.Wrapf(apperrors.ErrNoDocumentsFound, "chat ID: %s", chatID)
	}

	// Build context from documents
	var contextBuilder strings.Builder
	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, doc.Content))
	}
	context := contextBuilder.String()

	// Create a prompt template
	promptTemplate := prompts.NewPromptTemplate(
		"Context information is below.\n"+
			"---------------------\n"+
			"{{.context}}\n"+
			"---------------------\n"+
			"Given the context information and not prior knowledge, answer the query.\n"+
			"Query: {{.query}}\n"+
			"Answer: ",
		[]string{"context", "query"},
	)

	// Create OpenAI client
	llm, err := openai.New(
		openai.WithToken(c.openaiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to create OpenAI client")
	}

	// Use the LLM directly
	prompt, err := promptTemplate.Format(map[string]any{
		"context": context,
		"query":   message,
	})
	if err != nil {
		return "", apperrors.Wrap(err, "failed to format prompt")
	}

	// Generate response using the LLM
	completion, err := llm.Call(ctx, prompt, llms.WithMaxTokens(1000))
	if err != nil {
		return "", apperrors.Wrap(err, "failed to generate completion")
	}

	return completion, nil
}

// ChatWithChatbot sends a message to the LLM with relevant context from a specific chatbot
func (c *ChatService) ChatWithChatbot(ctx context.Context, chatbotID, userID, message string) (string, error) {
	// Retrieve the chatbot with authorization check
	chatbot, err := c.chatbotStore.FindChatbotByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return "", err // Already wrapped
	}

	// Check if user has access to this chatbot
	if chatbot.UserID != userID && userID != "" {
		return "", apperrors.Wrapf(apperrors.ErrUnauthorizedChatbotAccess,
			"user %s does not own chatbot %s", userID, chatbotID)
	}

	// Vectorize the query
	queryEmbedding, err := c.vectorizer.VectorizeText(ctx, message)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Find relevant documents for this chatbot
	docs, err := c.documentStore.FindSimilarDocumentsByChatbot(ctx, queryEmbedding, chatbotID, 5)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find similar documents: %v", err)
	}

	// Check if any documents were found for this chatbot
	if len(docs) == 0 {
		return "", apperrors.Wrapf(apperrors.ErrNoDocumentsFound, "chatbot ID: %s", chatbotID)
	}

	// Build context from documents
	var contextBuilder strings.Builder
	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, doc.Content))
	}
	context := contextBuilder.String()

	// Create custom prompt with chatbot's system instructions
	promptTemplate := prompts.NewPromptTemplate(
		chatbot.SystemInstructions+"\n\n"+
			"Context information is below.\n"+
			"---------------------\n"+
			"{{.context}}\n"+
			"---------------------\n"+
			"Given the context information and not prior knowledge, answer the query.\n"+
			"Query: {{.query}}\n"+
			"Answer: ",
		[]string{"context", "query"},
	)

	// Create OpenAI client with chatbot's model settings
	llm, err := openai.New(
		openai.WithToken(c.openaiKey),
		openai.WithModel(chatbot.ModelName),
	)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to create OpenAI client")
	}

	// Apply temperature separately after client creation if the package doesn't support it directly
	// Or consider using a method like this if available:
	// openai.WithModelParams(map[string]interface{}{
	//     "temperature": chatbot.TemperatureParam,
	// })

	// Format the prompt
	prompt, err := promptTemplate.Format(map[string]any{
		"context": context,
		"query":   message,
	})
	if err != nil {
		return "", apperrors.Wrap(err, "failed to format prompt")
	}

	// Generate response using the LLM with chatbot's max tokens
	completion, err := llm.Call(ctx, prompt, llms.WithMaxTokens(chatbot.MaxTokens))
	if err != nil {
		return "", apperrors.Wrap(err, "failed to generate completion")
	}

	return completion, nil
}

// CreateChatbot creates a new chatbot with default settings
func (s *ChatService) CreateChatbot(ctx context.Context, userID, name, description, systemInstructions string) (*store.Chatbot, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}
	if name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	// Set default values
	if systemInstructions == "" {
		systemInstructions = "You are a helpful AI assistant."
	}

	now := time.Now()
	chatbot := store.Chatbot{
		UserID:             userID,
		Name:               name,
		Description:        description,
		SystemInstructions: systemInstructions,
		ModelName:          "gpt-3.5-turbo", // Default model
		TemperatureParam:   0.7,             // Default temperature
		MaxTokens:          2000,            // Default max tokens
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err := s.chatbotStore.CreateChatbot(ctx, &chatbot)
	if err != nil {
		return nil, err // Already wrapped in the store
	}

	return &chatbot, nil
}

// GetChatbot retrieves a chatbot by ID and validates ownership
func (s *ChatService) GetChatbot(ctx context.Context, chatbotID, userID string) (*store.Chatbot, error) {
	chatbot, err := s.chatbotStore.FindChatbotByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return nil, err // Already wrapped in the store
	}

	// Check ownership
	if chatbot.UserID != userID {
		return nil, apperrors.Wrapf(apperrors.ErrUnauthorizedChatbotAccess,
			"user %s does not own chatbot %s", userID, chatbotID)
	}

	return chatbot, nil
}

// ListChatbots lists all chatbots owned by a user
func (s *ChatService) ListChatbots(ctx context.Context, userID string) ([]store.Chatbot, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}

	return s.chatbotStore.FindChatbotsByUserID(ctx, userID)
}

// UpdateChatbot updates a chatbot's basic information
func (s *ChatService) UpdateChatbot(ctx context.Context, chatbotID, userID, name, description string) (*store.Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}
	if name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	chatbot.Name = name
	chatbot.Description = description
	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.chatbotStore.UpdateChatbot(ctx, *chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateSystemInstructions updates a chatbot's system instructions
func (s *ChatService) UpdateSystemInstructions(ctx context.Context, chatbotID, userID, instructions string) (*store.Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}
	if instructions == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "instructions are required")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	chatbot.SystemInstructions = instructions
	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.chatbotStore.UpdateChatbot(ctx, *chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateModelSettings updates a chatbot's LLM model settings
func (s *ChatService) UpdateModelSettings(ctx context.Context, chatbotID, userID, modelName string, temperature float64, maxTokens int) (*store.Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}
	if modelName == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "model name is required")
	}
	if temperature < 0 || temperature > 2 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "temperature must be between 0 and 2")
	}
	if maxTokens <= 0 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "max tokens must be positive")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	chatbot.ModelName = modelName
	chatbot.TemperatureParam = temperature
	chatbot.MaxTokens = maxTokens
	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.chatbotStore.UpdateChatbot(ctx, *chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// DeleteChatbot deletes a chatbot
func (s *ChatService) DeleteChatbot(ctx context.Context, chatbotID, userID string) error {
	if chatbotID == "" || userID == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	return s.chatbotStore.DeleteChatbot(ctx, uuid.MustParse(chatbotID), userID)
}
