package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/internal/store"
)

// ChatHandler contains all the dependencies needed for API handlers
type ChatHandler struct {
	ChatService        *services.ChatService
	DocumentStore      *store.DocumentStore
	ChatbotStore       *store.ChatbotStore
	UploadsDir         string
	AuthMiddleware     *middleware.AuthMiddleware
	OwershipMiddleware *middleware.OwnershipMiddleware
}

// NewChatHandler creates a new API handler
func NewChatHandler(
	authMiddleware *middleware.AuthMiddleware,
	chatService *services.ChatService,
	documentStore *store.DocumentStore,
	chatbotStore *store.ChatbotStore,
	uploadsDir string,
	ownershipMiddlware *middleware.OwnershipMiddleware,
) *ChatHandler {
	return &ChatHandler{
		ChatService:        chatService,
		DocumentStore:      documentStore,
		ChatbotStore:       chatbotStore,
		UploadsDir:         uploadsDir,
		AuthMiddleware:     authMiddleware,
		OwershipMiddleware: ownershipMiddlware,
	}
}

// RegisterRoutes registers all API routes
func (h *ChatHandler) RegisterRoutes(app *fiber.App) {
	chat := app.Group("/chat", h.AuthMiddleware.RequireAuth)

	// File upload and management
	chat.Post("/chatbot", h.POST_CreateChatbot)
	chat.Get("/chatbots", h.GET_ListChatbots) // New route to list all chats
	chat.Post("/:chatID/upload", h.OwershipMiddleware.IsChatbotOwner, h.POST_UploadFile)
	chat.Delete("/:chatID/files/:filename", h.OwershipMiddleware.IsChatbotOwner, h.DELETE_ChatFile)
	chat.Put("/:chatID/files/:filename", h.OwershipMiddleware.IsChatbotOwner, h.PUT_UpdateFile)
	chat.Get("/:chatID/files", h.OwershipMiddleware.IsChatbotOwner, h.GET_ChatFiles)

	// Chat
	chat.Post("/:chatID/message", h.OwershipMiddleware.IsChatbotOwner, h.POST_ChatMessage)
}

// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags health
// @Accept json
// @Produce plain
// @Success 200 {string} string "VectorChat API is running"
// @Security ApiKeyAuth
// @Router /health [get]
func (h *ChatHandler) GET_HealthCheck(c *fiber.Ctx) error {
	return c.SendString("VectorChat API is running")
}

// @Summary Upload file
// @Description Upload a file to be used for chat context
// @Tags chat
// @Accept multipart/form-data
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param file formData file true "File to upload"
// @Success 200 {object} FileUploadResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/upload [post]
func (h *ChatHandler) POST_UploadFile(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID format", err, http.StatusBadRequest)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, "No file uploaded", err, http.StatusBadRequest)
	}

	// Create a unique filename
	filename := fmt.Sprintf("%s-%s", chatID, filepath.Base(file.Filename))
	uploadPath := filepath.Join(h.UploadsDir, filename)

	// Save the file
	if err := c.SaveFile(file, uploadPath); err != nil {
		return ErrorResponse(c, "Failed to save file", err)
	}

	// Add file to vector database
	docID := fmt.Sprintf("%s-%s", chatID, file.Filename)
	if err := h.ChatService.AddFile(c.Context(), docID, uploadPath, chatID); err != nil {
		return ErrorResponse(c, "Failed to vectorize file", err)
	}

	return c.JSON(fiber.Map{
		"message":    "File uploaded and vectorized successfully",
		"chat_id":    chatID,
		"chatbot_id": chatID,
		"file":       file.Filename,
	})
}

// @Summary Delete chat file
// @Description Delete a file from a chat session
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param filename path string true "File name"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/files/{filename} [delete]
func (h *ChatHandler) DELETE_ChatFile(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	filename := c.Params("filename")

	if chatID == "" || filename == "" {
		return ErrorResponse(c, "Chat ID and filename are required", nil, http.StatusBadRequest)
	}

	// Create the document ID that was used when uploading
	docID := fmt.Sprintf("%s-%s", chatID, filename)

	// Remove from database
	if err := h.DocumentStore.DeleteDocument(c.Context(), docID); err != nil {
		return ErrorResponse(c, "Failed to delete document", err)
	}

	// Remove the file from the uploads directory
	storedFilename := fmt.Sprintf("%s-%s", chatID, filename)
	filePath := filepath.Join(h.UploadsDir, storedFilename)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		// Log the error but don't fail the request if file doesn't exist
		log.Printf("Warning: Failed to delete file %s: %v", filePath, err)
	}

	return c.JSON(MessageResponse{
		Message: "File deleted successfully",
	})
}

// @Summary Update chat file
// @Description Update a file in a chat session
// @Tags chat
// @Accept multipart/form-data
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param filename path string true "File name"
// @Param file formData file true "Updated file"
// @Success 200 {object} FileUploadResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/files/{filename} [put]
func (h *ChatHandler) PUT_UpdateFile(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Failed to parse chat ID", err, http.StatusBadRequest)
	}
	filename := c.Params("filename")

	user, err := GetUser(c)
	if err != nil {
		return err
	}
	isOwner, err := h.ChatbotStore.CheckChatbotOwnership(c.Context(), chatID, user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to verify chatbot ownership", err)
	}
	if !isOwner {
		return ErrorResponse(c, "You don't have permission to access this chatbot", nil, http.StatusForbidden)
	}

	if filename == "" {
		return ErrorResponse(c, "Filename is required", nil, http.StatusBadRequest)
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, "No file uploaded", err, http.StatusBadRequest)
	}

	// Create file path
	uploadPath := filepath.Join(h.UploadsDir, fmt.Sprintf("%s-%s", chatID, filename))

	// Remove old file if it exists
	if err := os.Remove(uploadPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to delete old file %s: %v", uploadPath, err)
	}

	// Save the new file
	if err := c.SaveFile(file, uploadPath); err != nil {
		return ErrorResponse(c, "Failed to save file", err)
	}

	// Create document ID
	docID := fmt.Sprintf("%s-%s", chatID, filename)

	// Update in vector database
	if err := h.ChatService.AddFile(c.Context(), docID, uploadPath, chatID); err != nil {
		return ErrorResponse(c, "Failed to vectorize file", err)
	}

	return c.JSON(fiber.Map{
		"message": "File updated successfully",
		"chat_id": chatID,
		"file":    filename,
	})
}

// @Summary List chat files
// @Description List all files in a chat session
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Success 200 {object} ChatFilesResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/files [get]
func (h *ChatHandler) GET_ChatFiles(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	if chatID == "" {
		return ErrorResponse(c, "Chat ID is required", nil, http.StatusBadRequest)
	}

	// Get documents for this chat from the database
	docs, err := h.DocumentStore.GetDocumentsByPrefix(c.Context(), chatID+"-")
	if err != nil {
		return ErrorResponse(c, "Failed to retrieve documents", err)
	}

	// Format the response
	files := make([]map[string]string, 0, len(docs))
	for _, doc := range docs {
		// Extract filename from document ID (remove chatID- prefix)
		filename := strings.TrimPrefix(doc.ID, chatID+"-")
		files = append(files, map[string]string{
			"filename": filename,
			"id":       doc.ID,
		})
	}

	return c.JSON(fiber.Map{
		"chat_id": chatID,
		"files":   files,
	})
}

// @Summary Get list of chatbots
// @Description Get a list of all chatbots owned by the current user
// @Tags chat
// @Accept json
// @Produce json
// @Success 200 {array} ChatbotResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbots [get]
func (h *ChatHandler) GET_ListChatbots(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	chatbots, err := h.ChatService.ListChatbots(c.Context(), user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to retrieve chatbots", err)
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

	return c.JSON(fiber.Map{
		"chatbots": response,
	})
}

// @Summary Send chat message
// @Description Send a message and get a response with context from uploaded files
// @Tags chat
// @Accept json
// @Produce json
// @Param message body ChatMessageRequest true "Chat message"
// @Param chatID path string true "Chat session ID"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/message [post]
func (h *ChatHandler) POST_ChatMessage(c *fiber.Ctx) error {
	var req ChatMessageRequest

	chatID := c.Params("chatID")
	if chatID == "" {
		return ErrorResponse(c, "Chat ID is required", nil, http.StatusBadRequest)
	}

	if err := c.BodyParser(&req); err != nil {
		// Try to get query from form data if JSON parsing fails
		req.Query = c.FormValue("query")

		if req.Query == "" {
			return ErrorResponse(c, "Query parameter is required", nil, http.StatusBadRequest)
		}
	}

	// Get user ID from context if authenticated
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	userID := user.ID

	response, err := h.ChatService.ChatWithChatbot(c.Context(), chatID, userID, req.Query)

	if err != nil {
		if apperrors.Is(err, apperrors.ErrNoDocumentsFound) {
			return ErrorResponse(c, "No documents found for this chat. Please upload some files first.", err, http.StatusNotFound)
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedChatbotAccess) {
			return ErrorResponse(c, "You don't have permission to access this chatbot", err, http.StatusForbidden)
		} else if apperrors.Is(err, apperrors.ErrChatbotNotFound) {
			return ErrorResponse(c, "Chatbot not found", err, http.StatusNotFound)
		}
		return ErrorResponse(c, "Chat error", err)
	}

	return c.JSON(fiber.Map{
		"response": response,
	})
}

// @Summary Create a new chatbot
// @Description Create a new chatbot with specified configuration
// @Tags chat
// @Accept json
// @Produce json
// @Param chatbot body ChatbotCreateRequest true "Chatbot configuration"
// @Success 201 {object} ChatbotResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot [post]
func (h *ChatHandler) POST_CreateChatbot(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req ChatbotCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	// Validate request
	if req.Name == "" {
		return ErrorResponse(c, "Name is required", nil, http.StatusBadRequest)
	}

	// Set default values if not provided
	if req.ModelName == "" {
		req.ModelName = "gpt-4" // or your default model
	}
	if req.TemperatureParam == 0 {
		req.TemperatureParam = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2000
	}

	chatbot, err := h.ChatService.CreateChatbot(context.Background(), user.ID, req.Name, req.Description, req.SystemInstructions)
	if err != nil {
		return ErrorResponse(c, "Failed to create chatbot", err)
	}

	// Return response
	return c.Status(http.StatusCreated).JSON(ChatbotResponse{
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
