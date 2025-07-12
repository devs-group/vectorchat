package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
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

// ChatbotCreateRequest represents the request to create a new chatbot
type ChatbotCreateRequest struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	SystemInstructions string  `json:"system_instructions"`
	ModelName          string  `json:"model_name"`
	TemperatureParam   float64 `json:"temperature_param"`
	MaxTokens          int     `json:"max_tokens"`
}

// ChatbotUpdateRequest represents the request to update a chatbot
type ChatbotUpdateRequest struct {
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	SystemInstructions *string  `json:"system_instructions,omitempty"`
	ModelName          *string  `json:"model_name,omitempty"`
	TemperatureParam   *float64 `json:"temperature_param,omitempty"`
	MaxTokens          *int     `json:"max_tokens,omitempty"`
}

// ChatbotResponse represents a chatbot in responses
type ChatbotResponse struct {
	ID                 uuid.UUID `json:"id"`
	UserID             string    `json:"user_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	SystemInstructions string    `json:"system_instructions"`
	ModelName          string    `json:"model_name"`
	TemperatureParam   float64   `json:"temperature_param"`
	MaxTokens          int       `json:"max_tokens"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ChatMessageRequest represents a chat message request
type ChatMessageRequest struct {
	Query string `json:"query"`
}

// FileUploadResponse represents a file upload response
type FileUploadResponse struct {
	Message   string    `json:"message"`
	ChatID    uuid.UUID `json:"chat_id"`
	ChatbotID uuid.UUID `json:"chatbot_id"`
	File      string    `json:"file"`
}

// ChatFilesResponse represents the response for listing chat files
type ChatFilesResponse struct {
	ChatID uuid.UUID  `json:"chat_id"`
	Files  []FileInfo `json:"files"`
}

// FileInfo represents file information
type FileInfo struct {
	Filename   string    `json:"filename"`
	ID         uuid.UUID `json:"id"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ChatbotsListResponse represents the response for listing chatbots
type ChatbotsListResponse struct {
	Chatbots []ChatbotResponse `json:"chatbots"`
}

// ValidateAndCreateChatbot validates the request and creates a new chatbot
func (s *ChatService) ValidateAndCreateChatbot(ctx context.Context, userID string, req *ChatbotCreateRequest) (*ChatbotResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	// Set default values if not provided
	modelName := req.ModelName
	if modelName == "" {
		modelName = "gpt-4" // or your default model
	}

	temperatureParam := req.TemperatureParam
	if temperatureParam == 0 {
		temperatureParam = 0.7
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	chatbot, err := s.CreateChatbot(ctx, userID, req.Name, req.Description, req.SystemInstructions)
	if err != nil {
		return nil, err
	}

	return &ChatbotResponse{
		ID:                 chatbot.ID,
		UserID:             chatbot.UserID,
		Name:               chatbot.Name,
		Description:        chatbot.Description,
		SystemInstructions: chatbot.SystemInstructions,
		ModelName:          chatbot.ModelName,
		TemperatureParam:   chatbot.TemperatureParam,
		MaxTokens:          chatbot.MaxTokens,
		CreatedAt:          chatbot.CreatedAt,
		UpdatedAt:          chatbot.UpdatedAt,
	}, nil
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

// ListChatbotsFormatted lists all chatbots owned by a user with formatted response
func (s *ChatService) ListChatbotsFormatted(ctx context.Context, userID string) (*ChatbotsListResponse, error) {
	chatbots, err := s.ListChatbots(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Format the response
	response := make([]ChatbotResponse, 0, len(chatbots))
	for _, chatbot := range chatbots {
		response = append(response, ChatbotResponse{
			ID:                 chatbot.ID,
			UserID:             chatbot.UserID,
			Name:               chatbot.Name,
			Description:        chatbot.Description,
			SystemInstructions: chatbot.SystemInstructions,
			ModelName:          chatbot.ModelName,
			TemperatureParam:   chatbot.TemperatureParam,
			MaxTokens:          chatbot.MaxTokens,
			CreatedAt:          chatbot.CreatedAt,
			UpdatedAt:          chatbot.UpdatedAt,
		})
	}

	return &ChatbotsListResponse{
		Chatbots: response,
	}, nil
}

// UpdateChatbotFromRequest updates a chatbot from request data
func (s *ChatService) UpdateChatbotFromRequest(ctx context.Context, chatbotID, userID string, req *ChatbotUpdateRequest) (*ChatbotResponse, error) {
	chatbot, err := s.UpdateChatbotAll(ctx, chatbotID, userID, req.Name, req.Description, req.SystemInstructions, req.ModelName, req.TemperatureParam, req.MaxTokens)
	if err != nil {
		return nil, err
	}

	return &ChatbotResponse{
		ID:                 chatbot.ID,
		UserID:             chatbot.UserID,
		Name:               chatbot.Name,
		Description:        chatbot.Description,
		SystemInstructions: chatbot.SystemInstructions,
		ModelName:          chatbot.ModelName,
		TemperatureParam:   chatbot.TemperatureParam,
		MaxTokens:          chatbot.MaxTokens,
		CreatedAt:          chatbot.CreatedAt,
		UpdatedAt:          chatbot.UpdatedAt,
	}, nil
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

// GetChatbotFormatted retrieves a chatbot by ID with formatted response
func (s *ChatService) GetChatbotFormatted(ctx context.Context, chatbotID, userID string) (*ChatbotResponse, error) {
	chatbotUUID, err := uuid.Parse(chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "invalid chatbot ID format")
	}

	// Verify ownership
	isOwner, err := s.CheckChatbotOwnership(ctx, chatbotUUID, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to verify chatbot ownership")
	}
	if !isOwner {
		return nil, apperrors.ErrUnauthorizedChatbotAccess
	}

	// Get chatbot details
	chatbot, err := s.GetChatbotByID(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	if chatbot == nil {
		return nil, apperrors.ErrChatbotNotFound
	}

	return &ChatbotResponse{
		ID:                 chatbot.ID,
		UserID:             chatbot.UserID,
		Name:               chatbot.Name,
		Description:        chatbot.Description,
		SystemInstructions: chatbot.SystemInstructions,
		ModelName:          chatbot.ModelName,
		TemperatureParam:   chatbot.TemperatureParam,
		MaxTokens:          chatbot.MaxTokens,
		CreatedAt:          chatbot.CreatedAt,
		UpdatedAt:          chatbot.UpdatedAt,
	}, nil
}

// ProcessFileUpload handles file upload processing
func (s *ChatService) ProcessFileUpload(ctx context.Context, chatbotID uuid.UUID, file *multipart.FileHeader, uploadsDir string) (*FileUploadResponse, error) {
	// Create a unique filename
	filename := fmt.Sprintf("%s-%s", chatbotID, filepath.Base(file.Filename))
	uploadPath := filepath.Join(uploadsDir, filename)

	// Save the file
	src, err := file.Open()
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(uploadPath)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create file")
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return nil, apperrors.Wrap(err, "failed to save file")
	}

	// Add file to vector database
	docID := fmt.Sprintf("%s-%s", chatbotID, file.Filename)
	if err := s.AddFile(ctx, docID, uploadPath, chatbotID); err != nil {
		return nil, apperrors.Wrap(err, "failed to vectorize file")
	}

	return &FileUploadResponse{
		Message:   "File uploaded and vectorized successfully",
		ChatID:    chatbotID,
		ChatbotID: chatbotID,
		File:      file.Filename,
	}, nil
}

// ProcessFileUpdate handles file update processing
func (s *ChatService) ProcessFileUpdate(ctx context.Context, chatbotID uuid.UUID, filename string, file *multipart.FileHeader, uploadsDir string) (*MessageResponse, error) {
	// Create file path
	uploadPath := filepath.Join(uploadsDir, fmt.Sprintf("%s-%s", chatbotID, filename))

	// Remove old file if it exists
	if err := os.Remove(uploadPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to delete old file %s: %v", uploadPath, err)
	}

	// Save the new file
	src, err := file.Open()
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(uploadPath)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create file")
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return nil, apperrors.Wrap(err, "failed to save file")
	}

	// Create document ID
	docID := fmt.Sprintf("%s-%s", chatbotID, filename)

	// Update in vector database
	if err := s.AddFile(ctx, docID, uploadPath, chatbotID); err != nil {
		return nil, apperrors.Wrap(err, "failed to vectorize file")
	}

	return &MessageResponse{
		Message: "File updated successfully",
	}, nil
}

// ProcessFileDelete handles file deletion
func (s *ChatService) ProcessFileDelete(ctx context.Context, chatbotID, filename, uploadsDir string) error {
	// Create the document ID that was used when uploading
	docID := fmt.Sprintf("%s-%s", chatbotID, filename)

	// Remove from database
	if err := s.DeleteDocumentByID(ctx, docID); err != nil {
		return apperrors.Wrap(err, "failed to delete document")
	}

	// Remove the file from the uploads directory
	storedFilename := fmt.Sprintf("%s-%s", chatbotID, filename)
	filePath := filepath.Join(uploadsDir, storedFilename)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		// Log the error but don't fail the request if file doesn't exist
		log.Printf("Warning: Failed to delete file %s: %v", filePath, err)
	}

	return nil
}

// GetChatFilesFormatted retrieves all files for a chatbot with formatted response
func (s *ChatService) GetChatFilesFormatted(ctx context.Context, chatbotID uuid.UUID) (*ChatFilesResponse, error) {
	files, err := s.GetFilesByChatbotID(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	respFiles := make([]FileInfo, 0, len(files))
	for _, f := range files {
		respFiles = append(respFiles, FileInfo{
			Filename:   f.Filename,
			ID:         f.ID,
			UploadedAt: f.UploadedAt,
		})
	}

	return &ChatFilesResponse{
		ChatID: chatbotID,
		Files:  respFiles,
	}, nil
}

// ValidateAndParseQuery validates and parses a chat message request
func (s *ChatService) ValidateAndParseQuery(req *ChatMessageRequest, queryForm string) (string, error) {
	query := req.Query
	if query == "" && queryForm != "" {
		query = queryForm
	}

	if query == "" {
		return "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "query parameter is required")
	}

	return query, nil
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

// ParseChatID parses and validates a chat ID
func (s *ChatService) ParseChatID(chatIDStr string) (uuid.UUID, error) {
	if chatIDStr == "" {
		return uuid.Nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chat ID is required")
	}

	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(err, "invalid chat ID format")
	}

	return chatID, nil
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

// ValidateChatIDAndOwnership validates chat ID and user ownership
func (s *ChatService) ValidateChatIDAndOwnership(ctx context.Context, chatIDStr, userID string) (uuid.UUID, error) {
	chatID, err := s.ParseChatID(chatIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	isOwner, err := s.CheckChatbotOwnership(ctx, chatID, userID)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(err, "failed to verify chatbot ownership")
	}
	if !isOwner {
		return uuid.Nil, apperrors.ErrUnauthorizedChatbotAccess
	}

	return chatID, nil
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

// parsePositiveInt parses a string to positive integer with validation
func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, apperrors.New("empty string")
	}
	return strconv.Atoi(s)
}
