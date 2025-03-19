package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/pkg/chat"
	"github.com/yourusername/vectorchat/pkg/db"
	apperrors "github.com/yourusername/vectorchat/pkg/errors"
)

// Handler contains all the dependencies needed for API handlers
type Handler struct {
	ChatService *chat.ChatService
	Database    db.VectorDB
	UploadsDir  string
}

// NewHandler creates a new API handler
func NewHandler(chatService *chat.ChatService, database db.VectorDB, uploadsDir string) *Handler {
	return &Handler{
		ChatService: chatService,
		Database:    database,
		UploadsDir:  uploadsDir,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// Add a simple root route for testing
	app.Get("/", h.GET_HealthCheck)

	// File upload and management
	app.Post("/upload", h.POST_UploadFile)
	app.Delete("/files/:chatID/:filename", h.DELETE_ChatFile)
	app.Put("/files/:chatID/:filename", h.PUT_UpdateFile)
	app.Get("/files/:chatID", h.GET_ChatFiles)

	// Chat
	app.Post("/chat", h.POST_ChatMessage)
}

// GET_HealthCheck handles the health check endpoint
func (h *Handler) GET_HealthCheck(c *fiber.Ctx) error {
	return c.SendString("VectorChat API is running")
}

// POST_UploadFile handles file uploads
func (h *Handler) POST_UploadFile(c *fiber.Ctx) error {
	// Get chat ID from form
	chatID := c.FormValue("chat_id", fmt.Sprintf("chat-%d", time.Now().Unix()))

	// Get uploaded file
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
	if err := h.ChatService.AddFile(c.Context(), docID, uploadPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to vectorize file: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "File uploaded and vectorized successfully",
		"chat_id": chatID,
		"file":    file.Filename,
	})
}

// DELETE_ChatFile handles file deletion from a chat session
func (h *Handler) DELETE_ChatFile(c *fiber.Ctx) error {
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
	if err := h.Database.DeleteDocument(c.Context(), docID); err != nil {
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
		"file": filename,
	})
}

// PUT_UpdateFile handles updating files in a chat session
func (h *Handler) PUT_UpdateFile(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	filename := c.Params("filename")
	
	if chatID == "" || filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chat ID and filename are required",
		})
	}
	
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}
	
	// Create a unique filename
	storedFilename := fmt.Sprintf("%s-%s", chatID, filename)
	uploadPath := filepath.Join(h.UploadsDir, storedFilename)
	
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
	if err := h.ChatService.AddFile(c.Context(), docID, uploadPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to vectorize file: %v", err),
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "File updated successfully",
		"chat_id": chatID,
		"file": filename,
	})
}

// GET_ChatFiles handles listing all files in a chat session
func (h *Handler) GET_ChatFiles(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	
	if chatID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chat ID is required",
		})
	}
	
	// Get documents for this chat from the database
	docs, err := h.Database.GetDocumentsByPrefix(c.Context(), chatID+"-")
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
			"id": doc.ID,
		})
	}
	
	return c.JSON(fiber.Map{
		"chat_id": chatID,
		"files": files,
	})
}

// POST_ChatMessage handles sending messages to the chat system
func (h *Handler) POST_ChatMessage(c *fiber.Ctx) error {
	// Parse request
	var req struct {
		ChatID string `json:"chat_id"`
		Query  string `json:"query"`
	}

	if err := c.BodyParser(&req); err != nil {
		// Try to get query from form data if JSON parsing fails
		req.Query = c.FormValue("query")
		req.ChatID = c.FormValue("chat_id", "default")
		
		if req.Query == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Query parameter is required",
			})
		}
	}

	// Get response from chat service
	response, err := h.ChatService.ChatWithID(c.Context(), req.ChatID, req.Query)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrNoDocumentsFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No documents found for this chat ID. Please upload some files first.",
				"chat_id": req.ChatID,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Chat error: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"response": response,
		"chat_id":  req.ChatID,
	})
} 