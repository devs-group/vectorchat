package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/vectorize"
	"github.com/yourusername/vectorchat/pkg/models"
)

// ChatService handles chat interactions with context from vector database
type ChatService struct {
	*CommonService
	chatbotRepo  *db.ChatbotRepository
	sharedKBRepo *db.SharedKnowledgeBaseRepository
	documentRepo *db.DocumentRepository
	fileRepo     *db.FileRepository
	messageRepo  *db.ChatMessageRepository
	revisionRepo *db.RevisionRepository
	vectorizer   vectorize.Vectorizer
	openaiKey    string
	db           *db.Database
	kbService    *KnowledgeBaseService
}

// NewChatService creates a new chat service
func NewChatService(
	chatbotRepo *db.ChatbotRepository,
	sharedKBRepo *db.SharedKnowledgeBaseRepository,
	documentRepo *db.DocumentRepository,
	fileRepo *db.FileRepository,
	messageRepo *db.ChatMessageRepository,
	revisionRepo *db.RevisionRepository,
	vectorizer vectorize.Vectorizer,
	knowledgeService *KnowledgeBaseService,
	openaiKey string,
	database *db.Database,
) *ChatService {
	return &ChatService{
		CommonService: NewCommonService(),
		chatbotRepo:   chatbotRepo,
		sharedKBRepo:  sharedKBRepo,
		documentRepo:  documentRepo,
		fileRepo:      fileRepo,
		messageRepo:   messageRepo,
		revisionRepo:  revisionRepo,
		vectorizer:    vectorizer,
		openaiKey:     openaiKey,
		db:            database,
		kbService:     knowledgeService,
	}
}

// ChatbotCreateRequest represents the request to create a new chatbot
// Helper function to convert database Chatbot to models.ChatbotResponse
func (s *ChatService) toChatbotResponse(ctx context.Context, chatbot *db.Chatbot, aiMessages int64) (*models.ChatbotResponse, error) {
	sharedIDs, err := s.sharedKBRepo.ListIDsByChatbot(ctx, chatbot.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list shared knowledge bases for chatbot")
	}

	return &models.ChatbotResponse{
		ID:                     chatbot.ID,
		UserID:                 chatbot.UserID,
		Name:                   chatbot.Name,
		Description:            chatbot.Description,
		SystemInstructions:     chatbot.SystemInstructions,
		ModelName:              chatbot.ModelName,
		TemperatureParam:       chatbot.TemperatureParam,
		MaxTokens:              chatbot.MaxTokens,
		SaveMessages:           chatbot.SaveMessages,
		UseMaxTokens:           chatbot.UseMaxTokens,
		IsEnabled:              chatbot.IsEnabled,
		CreatedAt:              chatbot.CreatedAt,
		UpdatedAt:              chatbot.UpdatedAt,
		AIMessagesAmount:       aiMessages,
		SharedKnowledgeBaseIDs: sharedIDs,
	}, nil
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

	saveMessages := true
	if req.SaveMessages != nil {
		saveMessages = *req.SaveMessages
	}

	useMaxTokens := true
	if req.UseMaxTokens != nil {
		useMaxTokens = *req.UseMaxTokens
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	chatbot, err := s.CreateChatbot(ctx, userID, req.Name, req.Description, req.SystemInstructions, modelName, temperature, maxTokens, saveMessages, isEnabled, useMaxTokens)
	if err != nil {
		return nil, err
	}

	if err := s.replaceChatbotSharedKnowledgeBases(ctx, chatbot.ID, userID, req.SharedKnowledgeBaseIDs); err != nil {
		return nil, err
	}

	resp, err := s.toChatbotResponse(ctx, chatbot, 0)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateChatbot creates a new chatbot with default settings
func (s *ChatService) CreateChatbot(ctx context.Context, userID, name, description, systemInstructions, modelName string, temperature float64, maxTokens int, saveMessages bool, isEnabled bool, useMaxTokens bool) (*db.Chatbot, error) {
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
		UseMaxTokens:       useMaxTokens,
		SaveMessages:       saveMessages,
		IsEnabled:          isEnabled,
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

// GetChatbotForUser retrieves a chatbot owned by the provided user
func (s *ChatService) GetChatbotForUser(ctx context.Context, chatbotID, userID string) (*db.Chatbot, error) {
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	id, err := uuid.Parse(chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "invalid chatbot ID format")
	}

	return s.chatbotRepo.FindByIDAndUserID(ctx, id, userID)
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
	chatbotIDs := make([]uuid.UUID, 0, len(chatbots))
	for _, chatbot := range chatbots {
		chatbotIDs = append(chatbotIDs, chatbot.ID)
	}

	counts, err := s.messageRepo.CountAssistantMessagesByChatbotIDs(ctx, chatbotIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to count assistant messages")
	}

	for _, chatbot := range chatbots {
		resp, err := s.toChatbotResponse(ctx, chatbot, counts[chatbot.ID])
		if err != nil {
			return nil, err
		}
		formattedChatbots = append(formattedChatbots, *resp)
	}

	return &models.ChatbotsListResponse{
		Chatbots: formattedChatbots,
	}, nil
}

// UpdateChatbotFromRequest updates a chatbot from request data
func (s *ChatService) UpdateChatbotFromRequest(ctx context.Context, chatID, userID string, req *models.ChatbotUpdateRequest) (*models.ChatbotResponse, error) {
	chatbot, err := s.UpdateChatbotAll(ctx, chatID, userID, req.Name, req.Description, req.SystemInstructions, req.ModelName, req.TemperatureParam, req.MaxTokens, req.SaveMessages, req.UseMaxTokens)
	if err != nil {
		return nil, err
	}
	if req.SharedKnowledgeBaseIDs != nil {
		if err := s.replaceChatbotSharedKnowledgeBases(ctx, chatbot.ID, userID, req.SharedKnowledgeBaseIDs); err != nil {
			return nil, err
		}
	}

	count, err := s.messageRepo.CountAssistantMessagesByChatbotID(ctx, chatbot.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to count assistant messages")
	}

	resp, err := s.toChatbotResponse(ctx, chatbot, count)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateChatbotAll updates all chatbot fields in a single operation
func (s *ChatService) UpdateChatbotAll(ctx context.Context, chatbotID, userID string, name, description, systemInstructions, modelName *string, temperature *float64, maxTokens *int, saveMessages *bool, useMaxTokens *bool) (*db.Chatbot, error) {
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

	if saveMessages != nil {
		chatbot.SaveMessages = *saveMessages
	}

	if useMaxTokens != nil {
		chatbot.UseMaxTokens = *useMaxTokens
	}

	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.chatbotRepo.Update(ctx, chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

func (s *ChatService) replaceChatbotSharedKnowledgeBases(ctx context.Context, chatbotID uuid.UUID, ownerID string, kbIDs []uuid.UUID) error {
	if kbIDs == nil {
		return nil
	}

	ids := dedupeUUIDs(kbIDs)
	if len(ids) == 0 {
		return s.sharedKBRepo.ReplaceChatbotLinks(ctx, chatbotID, []uuid.UUID{})
	}

	owners, err := s.sharedKBRepo.ListOwnersByKnowledgeBaseIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, id := range ids {
		owner, ok := owners[id]
		if !ok {
			return apperrors.ErrSharedKnowledgeBaseNotFound
		}
		if owner != ownerID {
			return apperrors.ErrUnauthorizedKnowledgeBaseAccess
		}
	}

	return s.sharedKBRepo.ReplaceChatbotLinks(ctx, chatbotID, ids)
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

	count, err := s.messageRepo.CountAssistantMessagesByChatbotID(ctx, chatbotUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to count assistant messages")
	}

	resp, err := s.toChatbotResponse(ctx, chatbot, count)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ProcessFileUpload validates and stores a new file, similar to ProcessTextUpload
func (s *ChatService) ProcessFileUpload(ctx context.Context, chatbotID uuid.UUID, fileHeader *multipart.FileHeader) (*models.FileUploadResponse, error) {
	target := KnowledgeBaseTarget{ChatbotID: &chatbotID}
	file, err := s.kbService.IngestFile(ctx, target, fileHeader)
	if err != nil {
		return nil, err
	}

	return &models.FileUploadResponse{
		Message:   "File processed successfully",
		ChatID:    chatbotID,
		ChatbotID: chatbotID,
		File:      file.Filename,
		Filename:  file.Filename,
		Size:      file.SizeBytes,
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

// AddWebsite crawls a root URL (minimal BFS) and indexes extracted text as documents.
func (s *ChatService) AddWebsite(ctx context.Context, chatbotID uuid.UUID, rootURL string) error {
	target := KnowledgeBaseTarget{ChatbotID: &chatbotID}
	_, err := s.kbService.IngestWebsite(ctx, target, rootURL)
	return err
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
	if f.ChatbotID == nil || *f.ChatbotID != chatbotID {
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
	if f.ChatbotID == nil || *f.ChatbotID != chatbotID {
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

	return nil
}

// DeleteDocumentByID deletes a document by its ID
func (s *ChatService) DeleteDocumentByID(ctx context.Context, documentID string) error {
	return s.documentRepo.Delete(ctx, documentID)
}

// ProcessTextUpload validates and sends text to be indexed for a chatbot
func (s *ChatService) ProcessTextUpload(ctx context.Context, chatbotID uuid.UUID, text string) error {
	target := KnowledgeBaseTarget{ChatbotID: &chatbotID}
	_, err := s.kbService.IngestText(ctx, target, text)
	return err
}

// ParseChatID parses and validates a chat ID
func (s *ChatService) ParseChatID(chatIDStr string) (uuid.UUID, error) {
	return s.ParseUUID(chatIDStr)
}

// ChatWithChatbot handles chat interactions without streaming.
func (s *ChatService) ChatWithChatbot(ctx context.Context, chatbot *db.Chatbot, query string, sessionID *string) (string, string, error) {
	return s.chatWithChatbot(ctx, chatbot, query, sessionID, nil)
}

// ChatWithChatbotStream handles chat interactions and streams chunks via the provided callback.
func (s *ChatService) ChatWithChatbotStream(
	ctx context.Context,
	chatbot *db.Chatbot,
	query string,
	sessionID *string,
	streamFn func(context.Context, string) error,
) (string, string, error) {
	if streamFn == nil {
		return "", "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "stream function is required")
	}
	return s.chatWithChatbot(ctx, chatbot, query, sessionID, streamFn)
}

// chatWithChatbot contains the shared logic for handling chatbot conversations with optional streaming.
func (s *ChatService) chatWithChatbot(
	ctx context.Context,
	chatbot *db.Chatbot,
	query string,
	sessionID *string,
	streamFn func(context.Context, string) error,
) (string, string, error) {
	if chatbot == nil {
		return "", "", apperrors.Wrap(apperrors.ErrChatbotNotFound, "chatbot is required")
	}

	// Check if chatbot is enabled
	if !chatbot.IsEnabled {
		return "", "", apperrors.Wrap(apperrors.ErrUnauthorizedChatbotAccess, "chatbot is currently disabled")
	}

	chatbotUUID := chatbot.ID

	// Handle session ID
	var currentSessionID uuid.UUID
	var err error
	if sessionID != nil && *sessionID != "" {
		currentSessionID, err = uuid.Parse(*sessionID)
		if err != nil {
			return "", "", apperrors.Wrap(err, "invalid session ID format")
		}
	} else {
		currentSessionID = uuid.New()
	}

	// Save user's message when persistence is enabled
	if chatbot.SaveMessages {
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
	}

	// Vectorize the query for RAG
	queryEmbedding, err := s.vectorizer.VectorizeText(ctx, query)
	if err != nil {
		return "", "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Check for revised answers first (high priority)
	revisedAnswer, err := s.checkForRevisedAnswer(ctx, queryEmbedding, chatbotUUID, query)
	if err != nil {
		// Log error but continue - revisions are optional enhancement
		log.Printf("Warning: Failed to check for revised answers: %v", err)
	}

	if revisedAnswer != nil {
		slog.Info("revised answer similarity", "similarity", revisedAnswer.Similarity)
	} else {
		slog.Info("revised answer similarity", "similarity", 0.0)
	}
	// If we have a high-confidence revised answer, use it directly
	// If we have a high-confidence revised answer, use it directly
	if revisedAnswer != nil && revisedAnswer.Similarity > 0.95 {
		if streamFn != nil {
			if err := streamFn(ctx, revisedAnswer.RevisedAnswer); err != nil {
				return "", "", err
			}
		}
		// Save the revised answer as assistant's response
		assistantMessage := &db.ChatMessage{
			ID:        uuid.New(),
			ChatbotID: chatbotUUID,
			SessionID: currentSessionID,
			Role:      "assistant",
			Content:   revisedAnswer.RevisedAnswer,
			CreatedAt: time.Now(),
		}
		if err := s.messageRepo.Create(ctx, assistantMessage); err != nil {
			return "", "", apperrors.Wrap(err, "failed to save assistant message")
		}
		return revisedAnswer.RevisedAnswer, currentSessionID.String(), nil
	}

	// Find relevant documents inside of chatbot (RAG context)
	docs, err := s.documentRepo.FindSimilarByChatbot(ctx, queryEmbedding, chatbotUUID, 5)
	if err != nil {
		return "", "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find similar documents: %v", err)
	}

	sharedIDs, err := s.sharedKBRepo.ListIDsByChatbot(ctx, chatbotUUID)
	if err != nil {
		return "", "", apperrors.Wrap(err, "failed to list shared knowledge base ids")
	}

	// Find relevant documents across shared knowledge bases (RAG context)
	var sharedDocs []*db.DocumentWithEmbedding
	if len(sharedIDs) > 0 {
		sharedDocs, err = s.documentRepo.FindSimilarBySharedKnowledgeBases(ctx, queryEmbedding, sharedIDs, 5)
		if err != nil {
			return "", "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find shared knowledge documents: %v", err)
		}
	}

	combinedDocs := append(docs, sharedDocs...)

	// Build RAG context string
	var ragContextBuilder strings.Builder

	// Add revised answer to context if available (but not high confidence)
	if revisedAnswer != nil && revisedAnswer.Similarity > 0.80 {
		ragContextBuilder.WriteString("Note that the previous similar question and answer has been revised! Use this new answer!:\n")
		ragContextBuilder.WriteString("---------------------\n")
		ragContextBuilder.WriteString(fmt.Sprintf("Q: %s\n", revisedAnswer.Question))
		ragContextBuilder.WriteString(fmt.Sprintf("A: %s\n", revisedAnswer.RevisedAnswer))
		ragContextBuilder.WriteString("---------------------\n\n")
	}

	if len(combinedDocs) > 0 {
		ragContextBuilder.WriteString("Context information is below.\n")
		ragContextBuilder.WriteString("---------------------\n")
		for _, doc := range combinedDocs {
			ragContextBuilder.WriteString(string(doc.Content) + "\n\n")
		}
		ragContextBuilder.WriteString("---------------------\n")
	}

	// Fetch conversation history
	const historyLimit = 20
	var history []*db.ChatMessage
	if chatbot.SaveMessages {
		history, err = s.messageRepo.FindRecentBySessionID(ctx, currentSessionID, historyLimit)
		if err != nil {
			return "", "", apperrors.Wrap(err, "failed to fetch conversation history")
		}
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
	finalPrompt := fmt.Sprintf("%s\n\n%s\n%s\nGiven the context information and conversation history, answer the query.\nFormat rules:\n- Use a single H2 heading (##, ≤8 words) only when the reply is an explanation/guide or longer than 2 paragraphs. For short/direct answers or chatty replies, skip the heading.\n- Use Markdown lists when helpful.\n- Wrap any code in fenced blocks with the language tag (```js, ```python, etc.).\n- Do not return HTML; use only Markdown.\nQuery: %s\nAnswer:",
		chatbot.SystemInstructions,
		historyBuilder.String(),
		ragContextBuilder.String(),
		query,
	)

	// Generate response
	var streamedResponse strings.Builder
	callOptions := []llms.CallOption{
		llms.WithTemperature(chatbot.TemperatureParam),
	}
	if chatbot.UseMaxTokens {
		callOptions = append(callOptions, llms.WithMaxTokens(chatbot.MaxTokens))
	}
	if streamFn != nil {
		callOptions = append(callOptions, llms.WithStreamingFunc(func(callCtx context.Context, chunk []byte) error {
			if len(chunk) == 0 {
				return nil
			}
			textChunk := string(chunk)
			streamedResponse.WriteString(textChunk)
			return streamFn(callCtx, textChunk)
		}))
	}

	response, err := llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(finalPrompt),
			},
		},
	}, callOptions...)
	if err != nil {
		slog.Error("chatWithChatbot: LLM call failed",
			"chatbot_id", chatbotUUID.String(),
			"session_id", currentSessionID.String(),
			"err", err,
		)
		return "", "", apperrors.Wrap(err, "failed to generate completion")
	}

	reasoningContent := ""
	completion := streamedResponse.String()
	if completion == "" {
		for _, choice := range response.Choices {
			completion += choice.Content
			reasoningContent += choice.ReasoningContent
		}
	}

	if completion == "" {
		slog.Warn("chatWithChatbot: empty completion from LLM",
			"chatbot_id", chatbotUUID.String(),
			"session_id", currentSessionID.String(),
			"choices", len(response.Choices),
		)
	} else {
		slog.Info("chatWithChatbot: completion summary",
			"chatbot_id", chatbotUUID.String(),
			"session_id", currentSessionID.String(),
			"completion_len", len(completion),
			"max_tokens", chatbot.MaxTokens,
		)
	}

	// Save assistant's message
	if chatbot.SaveMessages {
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
	}

	return completion, currentSessionID.String(), nil
}

// checkForRevisedAnswer looks for similar questions that have been revised by admins
func (s *ChatService) checkForRevisedAnswer(ctx context.Context, queryEmbedding []float32, chatbotID uuid.UUID, query string) (*db.AnswerRevisionWithEmbedding, error) {
	// Look for highly similar revised answers via vector search
	revisions, err := s.revisionRepo.FindSimilarRevisions(ctx, queryEmbedding, chatbotID, 0.85, 1)
	if err != nil {
		return nil, err
	}
	if len(revisions) > 0 {
		return revisions[0], nil
	}
	return nil, nil
}

// CreateAnswerRevision creates a new answer revision for admin corrections
func (s *ChatService) CreateAnswerRevision(ctx context.Context, req *models.CreateRevisionRequest) (*db.AnswerRevision, error) {
	// Validate inputs
	if req.ChatbotID == uuid.Nil {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID is required")
	}
	if req.Question == "" || req.RevisedAnswer == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "question and revised answer are required")
	}

	// Generate embedding for the question
	questionEmbedding, err := s.vectorizer.VectorizeText(ctx, req.Question)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to vectorize question")
	}

	// Create revision
	revision := &db.AnswerRevision{
		ID:                uuid.New(),
		ChatbotID:         req.ChatbotID,
		OriginalMessageID: req.OriginalMessageID,
		Question:          req.Question,
		OriginalAnswer:    req.OriginalAnswer,
		RevisedAnswer:     req.RevisedAnswer,
		QuestionEmbedding: pgvector.NewVector(questionEmbedding),
		RevisionReason:    req.RevisionReason,
		RevisedBy:         req.RevisedBy,
		IsActive:          true,
	}

	// Save to database
	if err := s.revisionRepo.CreateRevision(ctx, revision); err != nil {
		return nil, err
	}

	return revision, nil
}

// GetConversations retrieves all conversations for a chatbot with pagination
func (s *ChatService) GetConversations(ctx context.Context, chatbotID uuid.UUID, limit, offset int) (*models.ConversationsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	total, err := s.revisionRepo.GetTotalConversationsCount(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	requestedOffset := offset
	effectiveOffset := offset
	if total == 0 {
		effectiveOffset = 0
	} else if limit > 0 {
		maxPageIndex := int((total - 1) / int64(limit))
		maxOffset := maxPageIndex * limit
		if effectiveOffset > maxOffset {
			effectiveOffset = maxOffset
		}
	}

	conversations, err := s.revisionRepo.GetConversations(ctx, chatbotID, limit, effectiveOffset)
	if err != nil {
		return nil, err
	}

	response := models.ConversationsResponse{
		Conversations: make([]models.ConversationResponse, 0, len(conversations)),
	}

	for _, conv := range conversations {
		response.Conversations = append(response.Conversations, models.ConversationResponse{
			SessionID:           conv.SessionID,
			FirstMessageContent: conv.FirstMessageContent,
			FirstMessageAt:      conv.FirstMessageAt,
			LastMessageAt:       conv.LastMessageAt,
		})
	}

	page := 1
	if limit > 0 {
		page = (effectiveOffset / limit) + 1
	}

	var totalPages int
	if total > 0 && limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	hasNext := totalPages > 0 && page < totalPages
	hasPrev := totalPages > 0 && page > 1

	var nextPage *int
	var prevPage *int
	var nextOffset *int
	var prevOffset *int

	if hasNext {
		np := page + 1
		nextPage = &np
		no := (np - 1) * limit
		nextOffset = &no
	}
	if hasPrev {
		pp := page - 1
		prevPage = &pp
		po := (pp - 1) * limit
		if po < 0 {
			po = 0
		}
		prevOffset = &po
	}

	var requestedOffsetPtr *int
	if requestedOffset != effectiveOffset {
		requestedOffsetPtr = &requestedOffset
	}

	response.Pagination = models.ConversationPagination{
		Page:            page,
		PerPage:         limit,
		TotalItems:      total,
		TotalPages:      totalPages,
		HasNextPage:     hasNext,
		HasPrevPage:     hasPrev,
		Offset:          effectiveOffset,
		RequestedOffset: requestedOffsetPtr,
		NextPage:        nextPage,
		PrevPage:        prevPage,
		NextOffset:      nextOffset,
		PrevOffset:      prevOffset,
	}

	return &response, nil
}

// DeleteConversation removes all messages associated with a session if it belongs to the chatbot.
func (s *ChatService) DeleteConversation(ctx context.Context, chatbotID, sessionID uuid.UUID) error {
	msgs, err := s.messageRepo.FindRecentBySessionID(ctx, sessionID, 1)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return apperrors.ErrNotFound
	}

	if msgs[0].ChatbotID != chatbotID {
		return apperrors.ErrUnauthorizedChatbotAccess
	}

	rows, err := s.messageRepo.DeleteByChatbotAndSessionID(ctx, chatbotID, sessionID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

// GetRevisions retrieves all revisions for a chatbot
func (s *ChatService) GetRevisions(ctx context.Context, chatbotID uuid.UUID, includeInactive bool) ([]*db.AnswerRevision, error) {
	return s.revisionRepo.GetRevisionsByChat(ctx, chatbotID, includeInactive)
}

// UpdateRevision updates an existing answer revision
func (s *ChatService) UpdateRevision(ctx context.Context, revisionID uuid.UUID, updates map[string]interface{}) error {
	// If question is being updated, regenerate embedding
	if question, ok := updates["question"].(string); ok && question != "" {
		questionEmbedding, err := s.vectorizer.VectorizeText(ctx, question)
		if err != nil {
			return apperrors.Wrap(err, "failed to vectorize updated question")
		}
		updates["question_embedding"] = pgvector.NewVector(questionEmbedding)
	}

	return s.revisionRepo.UpdateRevision(ctx, revisionID, updates)
}

// DeactivateRevision deactivates a revision (soft delete)
func (s *ChatService) DeactivateRevision(ctx context.Context, revisionID uuid.UUID) error {
	return s.revisionRepo.DeactivateRevision(ctx, revisionID)
}

// intPtr returns a pointer to the given int
func intPtr(i int) *int {
	return &i
}

func dedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) <= 1 {
		return ids
	}

	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// GetConversationMessages returns all messages in a session for a chatbot, verifying ownership and session mapping
func (s *ChatService) GetConversationMessages(ctx context.Context, chatbotID uuid.UUID, sessionID uuid.UUID) ([]models.MessageDetails, error) {
	// Fetch one message to validate session belongs to chatbot
	msgs, err := s.messageRepo.FindRecentBySessionID(ctx, sessionID, 1)
	if err != nil {
		return nil, err
	}
	if len(msgs) > 0 {
		if msgs[0].ChatbotID != chatbotID {
			return nil, apperrors.ErrUnauthorizedChatbotAccess
		}
	}

	// Fetch all messages chronologically
	all, err := s.messageRepo.FindAllBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	out := make([]models.MessageDetails, 0, len(all))
	for _, m := range all {
		out = append(out, models.MessageDetails{
			ID:        m.ID,
			ChatbotID: m.ChatbotID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}
