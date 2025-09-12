package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/yourusername/vectorchat/internal/crawler"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/vectorize"
	"github.com/yourusername/vectorchat/pkg/models"
)

// ChatService handles chat interactions with context from vector database
type ChatService struct {
	*CommonService
	chatbotRepo  *db.ChatbotRepository
	documentRepo *db.DocumentRepository
	fileRepo     *db.FileRepository
	messageRepo  *db.ChatMessageRepository
	vectorizer   vectorize.Vectorizer
	openaiKey    string
	db           *db.Database
	uploadsDir   string
}

// NewChatService creates a new chat service
func NewChatService(
	chatbotRepo *db.ChatbotRepository,
	documentRepo *db.DocumentRepository,
	fileRepo *db.FileRepository,
	messageRepo *db.ChatMessageRepository,
	vectorizer vectorize.Vectorizer,
	openaiKey string,
	db *db.Database,
	uploadsDir string,
) *ChatService {
	return &ChatService{
		CommonService: NewCommonService(),
		chatbotRepo:   chatbotRepo,
		documentRepo:  documentRepo,
		fileRepo:      fileRepo,
		messageRepo:   messageRepo,
		vectorizer:    vectorizer,
		openaiKey:     openaiKey,
		db:            db,
		uploadsDir:    uploadsDir,
	}
}

// ChatbotCreateRequest represents the request to create a new chatbot
// Helper function to convert database Chatbot to models.ChatbotResponse
func (s *ChatService) toChatbotResponse(chatbot *db.Chatbot) *models.ChatbotResponse {
	return &models.ChatbotResponse{
		ID:                 chatbot.ID,
		UserID:             chatbot.UserID,
		Name:               chatbot.Name,
		Description:        chatbot.Description,
		SystemInstructions: chatbot.SystemInstructions,
		ModelName:          chatbot.ModelName,
		TemperatureParam:   chatbot.TemperatureParam,
		MaxTokens:          chatbot.MaxTokens,
		IsEnabled:          chatbot.IsEnabled,
		CreatedAt:          chatbot.CreatedAt,
		UpdatedAt:          chatbot.UpdatedAt,
	}
}

// ValidateAndCreateChatbot validates the request and creates a new chatbot
func (s *ChatService) ValidateAndCreateChatbot(ctx context.Context, userID string, req *models.ChatbotCreateRequest) (*models.ChatbotResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	// Set default values if not provided
	modelName := req.ModelName
	if modelName == "" {
		modelName = "gpt-4" // or your default model
	}

	temperature := req.TemperatureParam
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	chatbot, err := s.CreateChatbot(ctx, userID, req.Name, req.Description, req.SystemInstructions, modelName, temperature, maxTokens)
	if err != nil {
		return nil, err
	}

	return s.toChatbotResponse(chatbot), nil
}

// CreateChatbot creates a new chatbot with default settings
func (s *ChatService) CreateChatbot(ctx context.Context, userID, name, description, systemInstructions, modelName string, temperature float64, maxTokens int) (*db.Chatbot, error) {
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
	chatbot := &db.Chatbot{
		ID:                 uuid.New(),
		UserID:             userID,
		Name:               name,
		Description:        description,
		SystemInstructions: systemInstructions,
		ModelName:          modelName,
		TemperatureParam:   temperature,
		MaxTokens:          maxTokens,
		IsEnabled:          true,
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
func (s *ChatService) GetChatbotByID(ctx context.Context, chatbotID string) (*db.Chatbot, error) {
	chatbot, err := s.chatbotRepo.FindByID(ctx, uuid.MustParse(chatbotID))
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// ListChatbots lists all chatbots owned by a user
func (s *ChatService) ListChatbots(ctx context.Context, userID string) ([]*db.Chatbot, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}

	return s.chatbotRepo.FindByUserID(ctx, userID)
}

// ListChatbotsFormatted lists all chatbots owned by a user with formatted response
func (s *ChatService) ListChatbotsFormatted(ctx context.Context, userID string) (*models.ChatbotsListResponse, error) {
	chatbots, err := s.ListChatbots(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Format the response
	var formattedChatbots []models.ChatbotResponse
	for _, chatbot := range chatbots {
		formattedChatbots = append(formattedChatbots, *s.toChatbotResponse(chatbot))
	}

	return &models.ChatbotsListResponse{
		Chatbots: formattedChatbots,
	}, nil
}

// UpdateChatbotFromRequest updates a chatbot from request data
func (s *ChatService) UpdateChatbotFromRequest(ctx context.Context, chatID, userID string, req *models.ChatbotUpdateRequest) (*models.ChatbotResponse, error) {
	chatbot, err := s.UpdateChatbotAll(ctx, chatID, userID, req.Name, req.Description, req.SystemInstructions, req.ModelName, req.TemperatureParam, req.MaxTokens)
	if err != nil {
		return nil, err
	}

	return s.toChatbotResponse(chatbot), nil
}

// UpdateChatbotAll updates all chatbot fields in a single operation
func (s *ChatService) UpdateChatbotAll(ctx context.Context, chatbotID, userID string, name, description, systemInstructions, modelName *string, temperature *float64, maxTokens *int) (*db.Chatbot, error) {
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

// ToggleChatbotEnabled toggles the enabled state of a chatbot
func (s *ChatService) ToggleChatbotEnabled(ctx context.Context, chatbotID, userID string, isEnabled bool) (*db.Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	// Parse chatbot ID
	chatbotUUID, err := uuid.Parse(chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "invalid chatbot ID format")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.chatbotRepo.FindByIDAndUserID(ctx, chatbotUUID, userID)
	if err != nil {
		return nil, err
	}

	// Update the enabled state
	chatbot.IsEnabled = isEnabled
	chatbot.UpdatedAt = time.Now()

	// Update in database
	if err := s.chatbotRepo.Update(ctx, chatbot); err != nil {
		return nil, apperrors.Wrap(err, "failed to update chatbot enabled state")
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
		storedFilename := file.Filename
		// Ensure we use the stored form "<chatbotID>-<original>" exactly once
		if !strings.HasPrefix(storedFilename, chatbotUUID.String()+"-") {
			storedFilename = fmt.Sprintf("%s-%s", chatbotUUID, storedFilename)
		}
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
func (s *ChatService) GetChatbotFormatted(ctx context.Context, chatID, userID string) (*models.ChatbotResponse, error) {
	chatbotUUID, err := uuid.Parse(chatID)
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
	chatbot, err := s.GetChatbotByID(ctx, chatbotUUID.String())
	if err != nil {
		return nil, err
	}

	if chatbot == nil {
		return nil, apperrors.ErrChatbotNotFound
	}

	return s.toChatbotResponse(chatbot), nil
}

// ProcessFileUpload validates and stores a new file, similar to ProcessTextUpload
func (s *ChatService) ProcessFileUpload(ctx context.Context, chatbotID uuid.UUID, fileHeader *multipart.FileHeader) (*models.FileUploadResponse, error) {
	f, size, err := s.SaveFile(ctx, chatbotID, fileHeader)
	if err != nil {
		return nil, err
	}

	return &models.FileUploadResponse{
		Message:   "File processed successfully",
		ChatID:    chatbotID,
		ChatbotID: chatbotID,
		File:      f.Filename,
		Filename:  f.Filename,
		Size:      size,
	}, nil
}

// ProcessFileUpdate handles file update processing
func (s *ChatService) ProcessFileUpdate(ctx context.Context, chatbotID uuid.UUID, filename string, fileHeader *multipart.FileHeader) (*models.FileUploadResponse, error) {
	// Find existing file by original filename
	existing, err := s.fileRepo.FindByChatbotIDAndFilename(ctx, chatbotID, filename)
	if err == nil && existing != nil {
		if err := s.DeleteFileSource(ctx, chatbotID, existing.ID.String()); err != nil {
			return nil, err
		}
	}

	f, size, err := s.SaveFile(ctx, chatbotID, fileHeader)
	if err != nil {
		return nil, err
	}

	return &models.FileUploadResponse{
		Message:   "File updated successfully",
		ChatID:    chatbotID,
		ChatbotID: chatbotID,
		File:      f.Filename,
		Filename:  f.Filename,
		Size:      size,
	}, nil
}

// ProcessFileDelete handles file deletion
func (s *ChatService) ProcessFileDelete(ctx context.Context, chatbotID, filename string) error {
	// Parse chatbot ID
	botID, err := s.ParseUUID(chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "invalid chatbot ID")
	}

	// Look up the file metadata using chatbot + filename, with a simple fallback
	file, err := s.fileRepo.FindByChatbotIDAndFilename(ctx, botID, filename)
	if err != nil || file == nil {
		// Try stripping a possible stored prefix if the client passed it
		prefix := fmt.Sprintf("%s-", botID)
		alt := strings.TrimPrefix(filename, prefix)
		if alt != filename {
			file, err = s.fileRepo.FindByChatbotIDAndFilename(ctx, botID, alt)
		}
		if err != nil || file == nil {
			return apperrors.Wrap(apperrors.ErrFileNotFound, "failed to find file")
		}
	}

	return s.DeleteFileSource(ctx, botID, file.ID.String())
}

// GetChatFilesFormatted retrieves all files for a chatbot with formatted response
func (s *ChatService) GetChatFilesFormatted(ctx context.Context, chatbotID uuid.UUID) (*models.ChatFilesResponse, error) {
	files, err := s.GetFilesByChatbotID(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	respFiles := make([]models.FileInfo, 0, len(files))
	for _, f := range files {
		respFiles = append(respFiles, models.FileInfo{
			Filename:   f.Filename,
			ID:         f.ID,
			Size:       f.SizeBytes,
			UploadedAt: f.UploadedAt,
		})
	}

	return &models.ChatFilesResponse{
		ChatID: chatbotID,
		Files:  respFiles,
	}, nil
}

// ValidateAndParseQuery validates and parses a chat message request
func (s *ChatService) ValidateAndParseQuery(req *models.ChatMessageRequest, formQuery string) (string, error) {
	query := req.Query
	if query == "" && formQuery != "" {
		query = formQuery
	}

	if query == "" {
		return "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "query parameter is required")
	}

	return query, nil
}

// AddFile adds a file to the vector database
// AddFileSource indexes a stored file for a chatbot by chunking/vectorizing content
func (s *ChatService) AddFileSource(ctx context.Context, chatbotID uuid.UUID, originalFilename, storedPath string) (*db.File, error) {
	fileID := uuid.New()

	// Start transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	// Insert file metadata with the original filename (cleaner for clients)
	// Determine actual stored file size on disk for accounting
	var sizeBytes int64
	if fi, err := os.Stat(storedPath); err == nil {
		sizeBytes = fi.Size()
	}

	file := &db.File{
		ID:         fileID,
		ChatbotID:  chatbotID,
		Filename:   filepath.Base(originalFilename),
		SizeBytes:  sizeBytes,
		UploadedAt: time.Now(),
	}

	if err := s.fileRepo.CreateTx(ctx, tx, file); err != nil {
		return nil, apperrors.Wrap(err, "failed to insert file metadata")
	}

	// Process file based on extension
	ext := strings.ToLower(filepath.Ext(storedPath))
	// Use a stable base for document IDs
	docIDBase := fmt.Sprintf("%s-%s", chatbotID, filepath.Base(originalFilename))
	if ext == ".pdf" {
		err = s.processPDFFile(ctx, tx, docIDBase, storedPath, chatbotID, &fileID)
	} else {
		err = s.processRegularFile(ctx, tx, docIDBase, storedPath, chatbotID, &fileID)
	}
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Wrap(err, "failed to commit transaction")
	}
	return file, nil
}

// SaveFile writes the uploaded file to disk and indexes it, similar to AddText
func (s *ChatService) SaveFile(ctx context.Context, chatbotID uuid.UUID, fileHeader *multipart.FileHeader) (*db.File, int64, error) {
	// Server-side file size limit (10 MB)
	const maxFileBytes = 10 * 1024 * 1024
	if fileHeader.Size > maxFileBytes {
		return nil, 0, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file exceeds maximum size (10MB)")
	}

	original := filepath.Base(fileHeader.Filename)
	storedName := fmt.Sprintf("%s-%s", chatbotID, original)
	storedPath := filepath.Join(s.uploadsDir, storedName)

	// Save the file to disk
	src, err := fileHeader.Open()
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(storedPath)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to create file")
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to save file")
	}

	// Index the stored file
	file, err := s.AddFileSource(ctx, chatbotID, original, storedPath)
	if err != nil {
		// Best effort cleanup of the stored file if indexing fails
		_ = os.Remove(storedPath)
		return nil, 0, apperrors.Wrap(err, "failed to index file")
	}
	return file, fileHeader.Size, nil
}

// AddText indexes plain text for a chatbot by chunking and vectorizing it
func (s *ChatService) AddText(ctx context.Context, chatbotID uuid.UUID, text string) error {
	if strings.TrimSpace(text) == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "text is required")
	}
	// Server-side limit to prevent excessive payloads
	const maxTextLength = 200_000 // 200 KB
	if len(text) > maxTextLength {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "text exceeds maximum allowed length")
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	// Create synthetic file metadata to group this text upload
	textFilename := fmt.Sprintf("text-%s.txt", time.Now().Format("20060102-150405"))
	file := &db.File{
		ID:         uuid.New(),
		ChatbotID:  chatbotID,
		Filename:   textFilename,
		SizeBytes:  int64(len(text)),
		UploadedAt: time.Now(),
	}
	if err := s.fileRepo.CreateTx(ctx, tx, file); err != nil {
		return apperrors.Wrap(err, "failed to insert text metadata")
	}

	// Chunk text similar to PDF processing
	const chunkSize = 1000
	chunks := chunkText(text, chunkSize)

	for i, chunk := range chunks {
		embedding, err := s.vectorizer.VectorizeText(ctx, chunk)
		if err != nil {
			return apperrors.Wrapf(err, "failed to vectorize text chunk %d", i)
		}

		doc := &db.DocumentWithEmbedding{
			ID:         fmt.Sprintf("%s-text-%d-%s", chatbotID, i, uuid.New()),
			Content:    []byte(chunk),
			Embedding:  embedding,
			ChatbotID:  chatbotID,
			FileID:     &file.ID, // link chunks to text source
			ChunkIndex: intPtr(i),
		}

		if err := s.documentRepo.StoreWithEmbeddingTx(ctx, tx, doc); err != nil {
			return apperrors.Wrapf(err, "failed to store text chunk %d", i)
		}
	}

	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

// AddWebsite crawls a root URL (minimal BFS) and indexes extracted text as documents.
func (s *ChatService) AddWebsite(ctx context.Context, chatbotID uuid.UUID, rootURL string) error {
	if strings.TrimSpace(rootURL) == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "url is required")
	}

	// Create a synthetic file record to group this website source
	// Filename format: website-<host>-<timestamp>
	host := rootURL
	if u, err := s.ParseURL(rootURL); err == nil {
		host = u.Hostname()
	}
	fname := fmt.Sprintf("website-%s-%s", host, time.Now().Format("20060102-150405"))

	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	file := &db.File{
		ID:         uuid.New(),
		ChatbotID:  chatbotID,
		Filename:   fname,
		SizeBytes:  0,
		UploadedAt: time.Now(),
	}
	if err := s.fileRepo.CreateTx(ctx, tx, file); err != nil {
		return apperrors.Wrap(err, "failed to insert website source")
	}

	// Run a small crawl
	pages, err := crawler.CrawlWebsite(ctx, rootURL, crawler.Options{MaxPages: 25, MaxDepth: 2, Timeout: 40 * time.Second})
	if err != nil {
		return apperrors.Wrap(err, "failed to crawl website")
	}

	// Index each page's text and accumulate total ingested bytes
	const chunkSize = 1000
	var totalBytes int64
	for pi, p := range pages {
		if strings.TrimSpace(p.Text) == "" {
			continue
		}
		chunks := chunkText(p.Text, chunkSize)
		for ci, chunk := range chunks {
			totalBytes += int64(len(chunk))
			emb, err := s.vectorizer.VectorizeText(ctx, chunk)
			if err != nil {
				return apperrors.Wrapf(err, "failed to vectorize page %d chunk %d", pi, ci)
			}
			doc := &db.DocumentWithEmbedding{
				ID:         fmt.Sprintf("%s-web-%d-%d-%s", chatbotID, pi, ci, uuid.New().String()),
				Content:    []byte(chunk),
				Embedding:  emb,
				ChatbotID:  chatbotID,
				FileID:     &file.ID,
				ChunkIndex: intPtr(ci),
			}
			if err := s.documentRepo.StoreWithEmbeddingTx(ctx, tx, doc); err != nil {
				return apperrors.Wrapf(err, "failed to store page %d chunk %d", pi, ci)
			}
		}

		// Update file size_bytes before committing
		file.SizeBytes = totalBytes
		if err := s.fileRepo.UpdateTx(ctx, tx, file); err != nil {
			return apperrors.Wrap(err, "failed to update website source size")
		}
	}

	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit website indexing")
	}
	return nil
}

// ParseURL is a helper to parse URL strings (minimal wrapper)
func (s *ChatService) ParseURL(u string) (*url.URL, error) {
	return url.Parse(u)
}

// processPDFFile processes a PDF file by extracting text and chunking it
func (s *ChatService) processPDFFile(ctx context.Context, tx *db.Transaction, id, filePath string, chatbotID uuid.UUID, fileID *uuid.UUID) error {
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

		doc := &db.DocumentWithEmbedding{
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
func (s *ChatService) processRegularFile(ctx context.Context, tx *db.Transaction, id, filePath string, chatbotID uuid.UUID, fileID *uuid.UUID) error {
	embedding, err := s.vectorizer.VectorizeFile(ctx, filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize file")
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to read file")
	}

	doc := &db.DocumentWithEmbedding{
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
func (s *ChatService) GetFilesByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*db.File, error) {
	// Exclude text sources from regular files list
	return s.fileRepo.FindNonTextByChatbotID(ctx, chatbotID)
}

// GetTextSources retrieves text sources for a chatbot
func (s *ChatService) GetTextSources(ctx context.Context, chatbotID uuid.UUID) ([]*db.File, error) {
	return s.fileRepo.FindTextByChatbotID(ctx, chatbotID)
}

// DeleteTextSource deletes a text source and its associated chunks
func (s *ChatService) DeleteTextSource(ctx context.Context, chatbotID uuid.UUID, sourceID string) error {
	fid, err := uuid.Parse(sourceID)
	if err != nil {
		return apperrors.Wrap(err, "invalid text source ID")
	}

	// Ensure the file exists and belongs to this chatbot and is a text source
	f, err := s.fileRepo.FindByID(ctx, fid)
	if err != nil {
		return apperrors.Wrap(err, "failed to find text source")
	}
	if f.ChatbotID != chatbotID {
		return apperrors.Wrap(apperrors.ErrUnauthorizedChatbotAccess, "text source does not belong to chatbot")
	}
	if !strings.HasPrefix(f.Filename, "text-") {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "not a text source")
	}

	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if err := s.documentRepo.DeleteByFileIDTx(ctx, tx, f.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete text documents")
	}
	if err := s.fileRepo.DeleteTx(ctx, tx, f.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete text source")
	}
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

// DeleteFileSource deletes a file source and its associated chunks and disk file
func (s *ChatService) DeleteFileSource(ctx context.Context, chatbotID uuid.UUID, sourceID string) error {
	fid, err := uuid.Parse(sourceID)
	if err != nil {
		return apperrors.Wrap(err, "invalid file source ID")
	}

	// Ensure the file exists and belongs to this chatbot
	f, err := s.fileRepo.FindByID(ctx, fid)
	if err != nil {
		return apperrors.Wrap(err, "failed to find file source")
	}
	if f.ChatbotID != chatbotID {
		return apperrors.Wrap(apperrors.ErrUnauthorizedChatbotAccess, "file does not belong to chatbot")
	}

	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if err := s.documentRepo.DeleteByFileIDTx(ctx, tx, f.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete file documents")
	}
	if err := s.fileRepo.DeleteTx(ctx, tx, f.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete file source")
	}
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}

	// Remove physical file from disk (best effort)
	storedName := fmt.Sprintf("%s-%s", chatbotID, f.Filename)
	filePath := filepath.Join(s.uploadsDir, storedName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to delete file %s: %v", filePath, err)
	}
	return nil
}

// DeleteDocumentByID deletes a document by its ID
func (s *ChatService) DeleteDocumentByID(ctx context.Context, documentID string) error {
	return s.documentRepo.Delete(ctx, documentID)
}

// ProcessTextUpload validates and sends text to be indexed for a chatbot
func (s *ChatService) ProcessTextUpload(ctx context.Context, chatbotID uuid.UUID, text string) error {
	return s.AddText(ctx, chatbotID, text)
}

// ParseChatID parses and validates a chat ID
func (s *ChatService) ParseChatID(chatIDStr string) (uuid.UUID, error) {
	return s.ParseUUID(chatIDStr)
}

// ChatWithChatbot handles chat interactions with chatbot context, including conversation history.
func (s *ChatService) ChatWithChatbot(ctx context.Context, chatID, userID, query string, sessionID *string) (string, string, error) {
	chatbotUUID, err := uuid.Parse(chatID)
	if err != nil {
		return "", "", apperrors.Wrap(err, "invalid chatbot ID format")
	}

	// Retrieve the chatbot with authorization check
	chatbot, err := s.chatbotRepo.FindByIDAndUserID(ctx, chatbotUUID, userID)
	if err != nil {
		return "", "", err
	}

	// Check if chatbot is enabled
	if !chatbot.IsEnabled {
		return "", "", apperrors.Wrap(apperrors.ErrUnauthorizedChatbotAccess, "chatbot is currently disabled")
	}

	// Handle session ID
	var currentSessionID uuid.UUID
	if sessionID != nil && *sessionID != "" {
		currentSessionID, err = uuid.Parse(*sessionID)
		if err != nil {
			return "", "", apperrors.Wrap(err, "invalid session ID format")
		}
	} else {
		currentSessionID = uuid.New()
	}

	// Save user's message
	userMessage := &db.ChatMessage{
		ID:        uuid.New(),
		ChatbotID: chatbotUUID,
		SessionID: currentSessionID,
		Role:      "user",
		Content:   query,
		CreatedAt: time.Now(),
	}
	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		return "", "", apperrors.Wrap(err, "failed to save user message")
	}

	// Vectorize the query for RAG
	queryEmbedding, err := s.vectorizer.VectorizeText(ctx, query)
	if err != nil {
		return "", "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Find relevant documents (RAG context)
	docs, err := s.documentRepo.FindSimilarByChatbot(ctx, queryEmbedding, chatID, 5)
	if err != nil {
		return "", "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find similar documents: %v", err)
	}

	// Build RAG context string
	var ragContextBuilder strings.Builder
	if len(docs) > 0 {
		ragContextBuilder.WriteString("Context information is below.\n")
		ragContextBuilder.WriteString("---------------------\n")
		for _, doc := range docs {
			ragContextBuilder.WriteString(string(doc.Content) + "\n\n")
		}
		ragContextBuilder.WriteString("---------------------\n")
	}

	// Fetch conversation history
	const historyLimit = 10
	history, err := s.messageRepo.FindRecentBySessionID(ctx, currentSessionID, historyLimit)
	if err != nil {
		return "", "", apperrors.Wrap(err, "failed to fetch conversation history")
	}

	// Build history string
	var historyBuilder strings.Builder
	if len(history) > 0 {
		historyBuilder.WriteString("This is the recent conversation history:\n")
		for _, msg := range history {
			historyBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}

	// Create OpenAI client with chatbot's model settings
	llm, err := openai.New(
		openai.WithToken(s.openaiKey),
		openai.WithModel(chatbot.ModelName),
	)
	if err != nil {
		return "", "", apperrors.Wrap(err, "failed to create OpenAI client")
	}

	// Construct the final prompt
	finalPrompt := fmt.Sprintf("%s\n\n%s\n%s\nGiven the context information and conversation history, answer the query.\nQuery: %s\nAnswer:",
		chatbot.SystemInstructions,
		historyBuilder.String(),
		ragContextBuilder.String(),
		query,
	)

	// Generate response
	completion, err := llm.Call(ctx, finalPrompt, llms.WithMaxTokens(chatbot.MaxTokens), llms.WithTemperature(chatbot.TemperatureParam))
	if err != nil {
		return "", "", apperrors.Wrap(err, "failed to generate completion")
	}

	// Save assistant's message
	assistantMessage := &db.ChatMessage{
		ID:        uuid.New(),
		ChatbotID: chatbotUUID,
		SessionID: currentSessionID,
		Role:      "assistant",
		Content:   completion,
		CreatedAt: time.Now(),
	}
	if err := s.messageRepo.Create(ctx, assistantMessage); err != nil {
		// Log this error, but don't fail the request since the user got a response
		log.Printf("ERROR: failed to save assistant message: %v", err)
	}

	return completion, currentSessionID.String(), nil
}

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
