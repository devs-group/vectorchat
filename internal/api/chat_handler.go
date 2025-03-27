package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
)

// ChatHandler contains all the dependencies needed for API handlers
type ChatHandler struct {
	ChatService        *services.ChatService
	DocumentStore      *db.DocumentStore
	ChatbotStore       *db.ChatbotStore
	UploadsDir         string
	AuthMiddleware     *middleware.AuthMiddleware
	OwershipMiddleware *middleware.OwnershipMiddleware
}

// NewChatHandler creates a new API handler
func NewChatHandler(authMiddleware *middleware.AuthMiddleware, chatService *services.ChatService, documentStore *db.DocumentStore, chatbotStore *db.ChatbotStore, uploadsDir string, ownershipMiddlware *middleware.OwnershipMiddleware) *ChatHandler {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID format",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Create a unique filename
	filename := fmt.Sprintf("%s-%s", chatID, filepath.Base(file.Filename))
	uploadPath := filepath.Join(h.UploadsDir, filename)

	// Save the file
	if err := c.SaveFile(file, uploadPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save file: %v", err),
		})
	}

	// Add file to vector database
	docID := fmt.Sprintf("%s-%s", chatID, file.Filename)
	if err := h.ChatService.AddFile(c.Context(), docID, uploadPath, chatID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to vectorize file: %v", err),
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chat ID and filename are required",
		})
	}

	// Create the document ID that was used when uploading
	docID := fmt.Sprintf("%s-%s", chatID, filename)

	// Remove from database
	if err := h.DocumentStore.DeleteDocument(c.Context(), docID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete document: %v", err),
		})
	}

	// Remove the file from the uploads directory
	storedFilename := fmt.Sprintf("%s-%s", chatID, filename)
	filePath := filepath.Join(h.UploadsDir, storedFilename)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		// Log the error but don't fail the request if file doesn't exist
		log.Printf("Warning: Failed to delete file %s: %v", filePath, err)
	}

	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
		"chat_id": chatID,
		"file":    filename,
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
	chatID, err := getUUIDParam(c, "chatID")
	if err != nil {
		return err
	}
	filename := c.Params("filename")

	if user, ok := c.Locals("user").(*db.User); ok {
		isOwner, err := h.ChatbotStore.CheckChatbotOwnership(c.Context(), chatID, user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to verify chatbot ownership: %v", err),
			})
		}
		if !isOwner {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You don't have permission to access this chatbot",
			})
		}
	}

	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Filename is required",
		})
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Create file path
	uploadPath := filepath.Join(h.UploadsDir, fmt.Sprintf("%s-%s", chatID, filename))

	// Remove old file if it exists
	if err := os.Remove(uploadPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to delete old file %s: %v", uploadPath, err)
	}

	// Save the new file
	if err := c.SaveFile(file, uploadPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save file: %v", err),
		})
	}

	// Create document ID
	docID := fmt.Sprintf("%s-%s", chatID, filename)

	// Update in vector database
	if err := h.ChatService.AddFile(c.Context(), docID, uploadPath, chatID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to vectorize file: %v", err),
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chat ID is required",
		})
	}

	// Get documents for this chat from the database
	docs, err := h.DocumentStore.GetDocumentsByPrefix(c.Context(), chatID+"-")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to retrieve documents: %v", err),
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chat ID is required",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		// Try to get query from form data if JSON parsing fails
		req.Query = c.FormValue("query")

		if req.Query == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Query parameter is required",
			})
		}
	}

	// Get user ID from context if authenticated
	var userID string
	if user, ok := c.Locals("user").(*db.User); ok {
		userID = user.ID
	}

	response, err := h.ChatService.ChatWithChatbot(c.Context(), chatID, userID, req.Query)

	if err != nil {
		if apperrors.Is(err, apperrors.ErrNoDocumentsFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "No documents found for this chat. Please upload some files first.",
				"chat_id": chatID,
			})
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedChatbotAccess) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You don't have permission to access this chatbot",
			})
		} else if apperrors.Is(err, apperrors.ErrChatbotNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Chatbot not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Chat error: %v", err),
		})
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
	// Get user from context
	user, ok := c.Locals("user").(*db.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Error: "Authentication required",
		})
	}

	// Parse request body
	var req ChatbotCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Error: "Name is required",
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Error: fmt.Sprintf("Failed to create chatbot: %v", err),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(ChatbotResponse{
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
