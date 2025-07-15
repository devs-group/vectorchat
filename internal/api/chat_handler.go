package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

// ChatHandler contains all the dependencies
type ChatHandler struct {
	ChatService        *services.ChatService
	UploadsDir         string
	AuthMiddleware     *middleware.AuthMiddleware
	OwershipMiddleware *middleware.OwnershipMiddleware
	CommonService      *services.CommonService
}

// NewChatHandler creates a new handler
func NewChatHandler(
	authMiddleware *middleware.AuthMiddleware,
	chatService *services.ChatService,
	uploadsDir string,
	ownershipMiddlware *middleware.OwnershipMiddleware,
	commonService *services.CommonService,
) *ChatHandler {
	return &ChatHandler{
		ChatService:        chatService,
		UploadsDir:         uploadsDir,
		AuthMiddleware:     authMiddleware,
		OwershipMiddleware: ownershipMiddlware,
		CommonService:      commonService,
	}
}

// RegisterRoutes registers all API routes
func (h *ChatHandler) RegisterRoutes(app *fiber.App) {
	// Health check endpoint (no auth required)
	app.Get("/health", h.GET_HealthCheck)

	chat := app.Group("/chat", h.AuthMiddleware.RequireAuth)

	// File upload and management
	chat.Post("/chatbot", h.POST_CreateChatbot)
	chat.Get("/chatbots", h.GET_ListChatbots)
	chat.Get("/chatbot/:chatID", h.OwershipMiddleware.IsChatbotOwner, h.GET_ChatbotByID)
	chat.Put("/chatbot/:chatID", h.OwershipMiddleware.IsChatbotOwner, h.PUT_UpdateChatbot)
	chat.Delete("/chatbot/:chatID", h.OwershipMiddleware.IsChatbotOwner, h.DELETE_Chatbot)
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
// @Success 200 {object} models.FileUploadResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/upload [post]
func (h *ChatHandler) POST_UploadFile(c *fiber.Ctx) error {
	chatID, err := h.ChatService.ParseChatID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, "No file uploaded", err, http.StatusBadRequest)
	}

	response, err := h.ChatService.ProcessFileUpload(c.Context(), chatID, file, h.UploadsDir)
	if err != nil {
		return ErrorResponse(c, "Failed to upload file", err)
	}

	return c.JSON(response)
}

// @Summary Delete chat file
// @Description Delete a file from a chat session
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param filename path string true "File name"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/files/{filename} [delete]
func (h *ChatHandler) DELETE_ChatFile(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	filename := c.Params("filename")

	if chatID == "" || filename == "" {
		return ErrorResponse(c, "Chat ID and filename are required", nil, http.StatusBadRequest)
	}

	if err := h.ChatService.ProcessFileDelete(c.Context(), chatID, filename, h.UploadsDir); err != nil {
		return ErrorResponse(c, "Failed to delete file", err)
	}

	return c.JSON(models.MessageResponse{
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
// @Success 200 {object} models.FileUploadResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/files/{filename} [put]
func (h *ChatHandler) PUT_UpdateFile(c *fiber.Ctx) error {
	chatID, err := h.CommonService.ParseUUID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}

	filename := c.Params("filename")
	if filename == "" {
		return ErrorResponse(c, "Filename is required", nil, http.StatusBadRequest)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, "No file uploaded", err, http.StatusBadRequest)
	}

	response, err := h.ChatService.ProcessFileUpdate(c.Context(), chatID, filename, file, h.UploadsDir)
	if err != nil {
		return ErrorResponse(c, "Failed to update file", err)
	}

	return c.JSON(response)
}

// @Summary List chat files
// @Description List all files in a chat session
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Success 200 {object} models.ChatFilesResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/files [get]
func (h *ChatHandler) GET_ChatFiles(c *fiber.Ctx) error {
	chatID, err := h.ChatService.ParseChatID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}

	response, err := h.ChatService.GetChatFilesFormatted(c.Context(), chatID)
	if err != nil {
		return ErrorResponse(c, "Failed to retrieve files", err)
	}

	return c.JSON(response)
}

// @Summary Get list of chatbots
// @Description Get a list of all chatbots owned by the current user
// @Tags chat
// @Accept json
// @Produce json
// @Success 200 {array} models.ChatbotsListResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbots [get]
func (h *ChatHandler) GET_ListChatbots(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	response, err := h.ChatService.ListChatbotsFormatted(c.Context(), user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to retrieve chatbots", err)
	}

	return c.JSON(response)
}

// @Summary Send chat message
// @Description Send a message and get a response with context from uploaded files
// @Tags chat
// @Accept json
// @Produce json
// @Param message body models.ChatMessageRequest true "Chat message"
// @Param chatID path string true "Chat session ID"
// @Success 200 {object} models.ChatResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/message [post]
func (h *ChatHandler) POST_ChatMessage(c *fiber.Ctx) error {
	var req models.ChatMessageRequest

	chatID := c.Params("chatID")
	if chatID == "" {
		return ErrorResponse(c, "Chat ID is required", nil, http.StatusBadRequest)
	}

	if err := c.BodyParser(&req); err != nil {
		req.Query = c.FormValue("query")
	}

	query, err := h.ChatService.ValidateAndParseQuery(&req, c.FormValue("query"))
	if err != nil {
		return ErrorResponse(c, "Invalid query parameter", err, http.StatusBadRequest)
	}

	user, err := GetUser(c)
	if err != nil {
		return err
	}

	response, err := h.ChatService.ChatWithChatbot(c.Context(), chatID, user.ID, query)
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
// @Param chatbot body models.ChatbotCreateRequest true "Chatbot configuration"
// @Success 201 {object} models.ChatbotResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot [post]
func (h *ChatHandler) POST_CreateChatbot(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	var req models.ChatbotCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	response, err := h.ChatService.ValidateAndCreateChatbot(c.Context(), user.ID, &req)
	if err != nil {
		return ErrorResponse(c, "Failed to create chatbot", err)
	}

	return c.Status(http.StatusCreated).JSON(response)
}

// @Summary Get chatbot by ID
// @Description Get details of a specific chatbot by ID
// @Tags chat
// @Accept json
// @Produce json
// @Param chatbotID path string true "Chatbot ID"
// @Success 200 {object} models.ChatbotResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot/{chatbotID} [get]
func (h *ChatHandler) GET_ChatbotByID(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	chatbot, err := h.ChatService.GetChatbotFormatted(c.Context(), c.Params("chatID"), user.ID)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrUnauthorizedChatbotAccess) {
			return ErrorResponse(c, "You don't have permission to access this chatbot", err, http.StatusForbidden)
		} else if apperrors.Is(err, apperrors.ErrChatbotNotFound) {
			return ErrorResponse(c, "Chatbot not found", err, http.StatusNotFound)
		}
		return ErrorResponse(c, "Failed to retrieve chatbot", err)
	}

	return c.JSON(fiber.Map{
		"chatbot": chatbot,
	})
}

// @Summary Update chatbot
// @Description Update chatbot configuration including name, description, system instructions, model settings
// @Tags chat
// @Accept json
// @Produce json
// @Param chatbotID path string true "Chatbot ID"
// @Param chatbot body models.ChatbotUpdateRequest true "Updated chatbot configuration"
// @Success 200 {object} models.ChatbotResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot/{chatbotID} [put]
func (h *ChatHandler) PUT_UpdateChatbot(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	var req models.ChatbotUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	chatbot, err := h.ChatService.UpdateChatbotFromRequest(c.Context(), c.Params("chatID"), user.ID, &req)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrUnauthorizedChatbotAccess) {
			return ErrorResponse(c, "You don't have permission to access this chatbot", err, http.StatusForbidden)
		} else if apperrors.Is(err, apperrors.ErrChatbotNotFound) {
			return ErrorResponse(c, "Chatbot not found", err, http.StatusNotFound)
		} else if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			return ErrorResponse(c, "Invalid chatbot parameters", err, http.StatusBadRequest)
		}
		return ErrorResponse(c, "Failed to update chatbot", err)
	}

	return c.JSON(fiber.Map{
		"chatbot": chatbot,
	})
}

// @Summary Delete chatbot
// @Description Delete a chatbot and all associated data including files, documents, and conversations
// @Tags chat
// @Accept json
// @Produce json
// @Param chatbotID path string true "Chatbot ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot/{chatbotID} [delete]
func (h *ChatHandler) DELETE_Chatbot(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	chatID := c.Params("chatID")
	if chatID == "" {
		return ErrorResponse(c, "Chat ID is required", nil, http.StatusBadRequest)
	}

	err = h.ChatService.DeleteChatbot(c.Context(), chatID, user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to delete chatbot", err)
	}

	return c.JSON(models.MessageResponse{
		Message: "Chatbot deleted successfully",
	})
}
