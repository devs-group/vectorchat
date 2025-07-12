package services

import (
	"context"
	"fmt"
	"io/ioutil"
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
}

// NewChatService creates a new chat service
func NewChatService(
	chatbotRepo ChatbotRepositoryTx,
	documentRepo DocumentRepositoryTx,
	fileRepo FileRepositoryTx,
	vectorizer vectorize.Vectorizer,
	openaiKey string,
	db *Database,
) *ChatService {
	return &ChatService{
		chatbotRepo:  chatbotRepo,
		documentRepo: documentRepo,
		fileRepo:     fileRepo,
		vectorizer:   vectorizer,
		openaiKey:    openaiKey,
		db:           db,
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

// GetChatbot retrieves a chatbot by ID and validates ownership
func (s *ChatService) GetChatbot(ctx context.Context, chatbotID, userID string) (*Chatbot, error) {
	chatbot, err := s.chatbotRepo.FindByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return nil, err
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

// ListChatbotsWithPagination lists chatbots with pagination
func (s *ChatService) ListChatbotsWithPagination(ctx context.Context, userID string, offset, limit int) (*PaginatedResponse, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}

	chatbots, total, err := s.chatbotRepo.FindByUserIDWithPagination(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return NewPaginatedResponse(chatbots, total, offset, limit), nil
}

// UpdateChatbot updates a chatbot's basic information
func (s *ChatService) UpdateChatbot(ctx context.Context, chatbotID, userID, name, description string) (*Chatbot, error) {
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
	err = s.chatbotRepo.Update(ctx, chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateSystemInstructions updates a chatbot's system instructions
func (s *ChatService) UpdateSystemInstructions(ctx context.Context, chatbotID, userID, instructions string) (*Chatbot, error) {
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
	err = s.chatbotRepo.Update(ctx, chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateModelSettings updates a chatbot's LLM model settings
func (s *ChatService) UpdateModelSettings(ctx context.Context, chatbotID, userID, modelName string, temperature float64, maxTokens int) (*Chatbot, error) {
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
	err = s.chatbotRepo.Update(ctx, chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateChatbotAll updates all chatbot fields in a single operation
func (s *ChatService) UpdateChatbotAll(ctx context.Context, chatbotID, userID string, name, description, systemInstructions, modelName *string, temperature *float64, maxTokens *int) (*Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
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

	// Delete associated files
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

// GetFilesByChatbotIDWithPagination retrieves files with pagination
func (s *ChatService) GetFilesByChatbotIDWithPagination(ctx context.Context, chatbotID uuid.UUID, offset, limit int) (*PaginatedResponse, error) {
	files, total, err := s.fileRepo.FindByChatbotIDWithPagination(ctx, chatbotID, offset, limit)
	if err != nil {
		return nil, err
	}

	return NewPaginatedResponse(files, total, offset, limit), nil
}

// DeleteFile deletes a file and all associated documents
func (s *ChatService) DeleteFile(ctx context.Context, fileID uuid.UUID, userID string) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	// Get file to check ownership through chatbot
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return err
	}

	// Check chatbot ownership
	owns, err := s.chatbotRepo.CheckOwnership(ctx, file.ChatbotID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return apperrors.ErrUnauthorizedChatbotAccess
	}

	// Delete associated documents
	err = s.documentRepo.DeleteByFileIDTx(ctx, tx, fileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete file documents")
	}

	// Delete the file
	err = s.fileRepo.DeleteTx(ctx, tx, fileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete file")
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}

	return nil
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

// GetChatbotStats returns statistics about a chatbot
func (s *ChatService) GetChatbotStats(ctx context.Context, chatbotID uuid.UUID, userID string) (map[string]interface{}, error) {
	// Check ownership
	owns, err := s.chatbotRepo.CheckOwnership(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, apperrors.ErrUnauthorizedChatbotAccess
	}

	// Get chatbot
	chatbot, err := s.chatbotRepo.FindByID(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	// Get document count
	docCount, err := s.documentRepo.CountByChatbotID(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	// Get file count
	fileCount, err := s.fileRepo.CountByChatbotID(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"chatbot_id":          chatbot.ID,
		"name":                chatbot.Name,
		"description":         chatbot.Description,
		"model_name":          chatbot.ModelName,
		"temperature":         chatbot.TemperatureParam,
		"max_tokens":          chatbot.MaxTokens,
		"created_at":          chatbot.CreatedAt,
		"updated_at":          chatbot.UpdatedAt,
		"document_count":      docCount,
		"file_count":          fileCount,
		"system_instructions": chatbot.SystemInstructions,
	}

	return stats, nil
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
