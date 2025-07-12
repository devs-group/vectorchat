package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/vectorize"
)

// ChatService handles chat interactions with context from vector database
type ChatService struct {
	chatbotRepo  ChatbotRepositoryTx
	documentRepo DocumentRepositoryTx
	fileRepo     FileRepositoryTx
	vectorizer   vectorize.Vectorizer
	openaiKey    string
	db           *Database
	uploadsDir   string
}

// NewChatService creates a new chat service
func NewChatService(
	chatbotRepo ChatbotRepositoryTx,
	documentRepo DocumentRepositoryTx,
	fileRepo FileRepositoryTx,
	vectorizer vectorize.Vectorizer,
	openaiKey string,
	db *Database,
	uploadsDir string,
) *ChatService {
	return &ChatService{
		chatbotRepo:  chatbotRepo,
		documentRepo: documentRepo,
		fileRepo:     fileRepo,
		vectorizer:   vectorizer,
		openaiKey:    openaiKey,
		db:           db,
		uploadsDir:   uploadsDir,
	}
}

// CreateChatbot creates a new chatbot with default settings
func (s *ChatService) CreateChatbot(ctx context.Context, userID, name, description, systemInstructions string) (*Chatbot, error) {
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
	chatbot := &Chatbot{
		ID:                 uuid.New(),
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

	err := s.chatbotRepo.Create(ctx, chatbot)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create chatbot")
	}

	return chatbot, nil
}

// GetChatbotByID retrieves a chatbot by ID without ownership validation
func (s *ChatService) GetChatbotByID(ctx context.Context, chatbotID string) (*Chatbot, error) {
	chatbot, err := s.chatbotRepo.FindByID(ctx, uuid.MustParse(chatbotID))
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// ListChatbots lists all chatbots owned by a user
func (s *ChatService) ListChatbots(ctx context.Context, userID string) ([]*Chatbot, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}

	return s.chatbotRepo.FindByUserID(ctx, userID)
}

// UpdateChatbotAll updates all chatbot fields in a single operation
func (s *ChatService) UpdateChatbotAll(ctx context.Context, chatbotID, userID string, name, description, systemInstructions, modelName *string, temperature *float64, maxTokens *int) (*Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.chatbotRepo.FindByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return nil, err
	}

	// Update fields only if provided
	if name != nil {
		if *name == "" {
			return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name cannot be empty")
		}
		chatbot.Name = *name
	}

	if description != nil {
		chatbot.Description = *description
	}

	if systemInstructions != nil {
		chatbot.SystemInstructions = *systemInstructions
	}

	if modelName != nil {
		if *modelName == "" {
			return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "model name cannot be empty")
		}
		chatbot.ModelName = *modelName
	}

	if temperature != nil {
		if *temperature < 0 || *temperature > 2 {
			return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "temperature must be between 0 and 2")
		}
		chatbot.TemperatureParam = *temperature
	}

	if maxTokens != nil {
		if *maxTokens <= 0 {
			return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "max tokens must be positive")
		}
		chatbot.MaxTokens = *maxTokens
	}

	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.chatbotRepo.Update(ctx, chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// DeleteChatbot deletes a chatbot and all associated data
func (s *ChatService) DeleteChatbot(ctx context.Context, chatbotID, userID string) error {
	if chatbotID == "" || userID == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	chatbotUUID := uuid.MustParse(chatbotID)

	// Get all files before deleting them from database (for physical file cleanup)
	files, err := s.fileRepo.FindByChatbotID(ctx, chatbotUUID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get chatbot files for cleanup")
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	// Delete associated documents
	err = s.documentRepo.DeleteByChatbotIDTx(ctx, tx, chatbotUUID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete chatbot documents")
	}

	// Delete associated files from database
	err = s.fileRepo.DeleteByChatbotIDTx(ctx, tx, chatbotUUID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete chatbot files")
	}

	// Delete the chatbot
	err = s.chatbotRepo.DeleteTx(ctx, tx, chatbotUUID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete chatbot")
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}

	// Delete physical files from uploads directory
	for _, file := range files {
		storedFilename := fmt.Sprintf("%s-%s", chatbotID, file.Filename)
		filePath := filepath.Join(s.uploadsDir, storedFilename)

		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			// Log the error but don't fail the entire operation if file doesn't exist
			log.Printf("Warning: Failed to delete physical file %s: %v", filePath, err)
		}
	}

	return nil
}

// CheckChatbotOwnership verifies if a user owns a specific chatbot
func (s *ChatService) CheckChatbotOwnership(ctx context.Context, chatbotID uuid.UUID, userID string) (bool, error) {
	return s.chatbotRepo.CheckOwnership(ctx, chatbotID, userID)
}

// AddFile adds a file to the vector database
func (s *ChatService) AddFile(ctx context.Context, id string, filePath string, chatbotID uuid.UUID) error {
	fileID := uuid.New()
	filename := filepath.Base(filePath)

	// Start transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	// Insert file metadata
	file := &File{
		ID:         fileID,
		ChatbotID:  chatbotID,
		Filename:   filename,
		UploadedAt: time.Now(),
	}

	err = s.fileRepo.CreateTx(ctx, tx, file)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert file metadata")
	}

	// Process file based on extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".pdf" {
		err = s.processPDFFile(ctx, tx, id, filePath, chatbotID, &fileID)
	} else {
		err = s.processRegularFile(ctx, tx, id, filePath, chatbotID, &fileID)
	}

	if err != nil {
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// processPDFFile processes a PDF file by extracting text and chunking it
func (s *ChatService) processPDFFile(ctx context.Context, tx *Transaction, id, filePath string, chatbotID uuid.UUID, fileID *uuid.UUID) error {
	// Extract text from PDF
	pdfText, err := vectorize.ExtractTextFromPDF(filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to extract text from PDF")
	}

	// Chunk the text (e.g., 1000 chars per chunk)
	const chunkSize = 1000
	chunks := chunkText(pdfText, chunkSize)

	for i, chunk := range chunks {
		embedding, err := s.vectorizer.VectorizeText(ctx, chunk)
		if err != nil {
			return apperrors.Wrapf(err, "failed to vectorize PDF chunk %d", i)
		}

		doc := &DocumentWithEmbedding{
			ID:         fmt.Sprintf("%s-%d", id, i),
			Content:    []byte(chunk),
			Embedding:  embedding,
			ChatbotID:  chatbotID,
			FileID:     fileID,
			ChunkIndex: intPtr(i),
		}

		err = s.documentRepo.StoreWithEmbeddingTx(ctx, tx, doc)
		if err != nil {
			return apperrors.Wrapf(err, "failed to store PDF chunk %d", i)
		}
	}

	return nil
}

// processRegularFile processes a regular file by vectorizing its content
func (s *ChatService) processRegularFile(ctx context.Context, tx *Transaction, id, filePath string, chatbotID uuid.UUID, fileID *uuid.UUID) error {
	embedding, err := s.vectorizer.VectorizeFile(ctx, filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize file")
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to read file")
	}

	doc := &DocumentWithEmbedding{
		ID:         id,
		Content:    content,
		Embedding:  embedding,
		ChatbotID:  chatbotID,
		FileID:     fileID,
		ChunkIndex: nil,
	}

	return s.documentRepo.StoreWithEmbeddingTx(ctx, tx, doc)
}

// GetFilesByChatbotID retrieves all files for a given chatbot
func (s *ChatService) GetFilesByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*File, error) {
	return s.fileRepo.FindByChatbotID(ctx, chatbotID)
}

// DeleteDocumentByID deletes a document by its ID
func (s *ChatService) DeleteDocumentByID(ctx context.Context, documentID string) error {
	return s.documentRepo.Delete(ctx, documentID)
}

// ChatWithChatbot handles chat interactions with chatbot context
func (s *ChatService) ChatWithChatbot(ctx context.Context, chatbotID, userID, query string) (string, error) {
	// Retrieve the chatbot with authorization check
	chatbot, err := s.chatbotRepo.FindByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return "", err
	}

	// Vectorize the query
	queryEmbedding, err := s.vectorizer.VectorizeText(ctx, query)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Find relevant documents for this chatbot
	docs, err := s.documentRepo.FindSimilarByChatbot(ctx, queryEmbedding, chatbotID, 5)
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
		openai.WithToken(s.openaiKey),
		openai.WithModel(chatbot.ModelName),
	)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to create OpenAI client")
	}

	// Format the prompt
	prompt, err := promptTemplate.Format(map[string]any{
		"context": context,
		"query":   query,
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

// Helper functions

// chunkText splits text into chunks of the given size
func chunkText(text string, size int) []string {
	var chunks []string
	for start := 0; start < len(text); start += size {
		end := start + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
	}
	return chunks
}

// intPtr returns a pointer to the given int
func intPtr(i int) *int {
	return &i
}
