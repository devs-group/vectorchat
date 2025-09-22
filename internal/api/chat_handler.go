package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/constants"
	"github.com/yourusername/vectorchat/pkg/models"
)

type ChatHandler struct {
	ChatService        *services.ChatService
	AuthMiddleware     *middleware.AuthMiddleware
	OwershipMiddleware *middleware.OwnershipMiddleware
	CommonService      *services.CommonService
	SubscriptionLimits *middleware.SubscriptionLimitsMiddleware
}

func NewChatHandler(
	authMiddleware *middleware.AuthMiddleware,
	chatService *services.ChatService,
	ownershipMiddlware *middleware.OwnershipMiddleware,
	commonService *services.CommonService,
	subscriptionLimits *middleware.SubscriptionLimitsMiddleware,
) *ChatHandler {
	return &ChatHandler{
		ChatService:        chatService,
		AuthMiddleware:     authMiddleware,
		OwershipMiddleware: ownershipMiddlware,
		CommonService:      commonService,
		SubscriptionLimits: subscriptionLimits,
	}
}

func (h *ChatHandler) RegisterRoutes(app *fiber.App) {
	// Health check endpoint (no auth required)
	app.Get("/health", h.GET_HealthCheck)

	chat := app.Group("/chat", h.AuthMiddleware.RequireAuth)

	// File upload and management
	chat.Post("/chatbot", h.SubscriptionLimits.CheckLimit(constants.LimitChatbots), h.POST_CreateChatbot)
	chat.Get("/chatbots", h.GET_ListChatbots)
	chat.Get("/chatbot/:chatID", h.OwershipMiddleware.IsChatbotOwner, h.GET_ChatbotByID)
	chat.Put("/chatbot/:chatID", h.OwershipMiddleware.IsChatbotOwner, h.PUT_UpdateChatbot)
	chat.Patch("/chatbot/:chatID/toggle", h.OwershipMiddleware.IsChatbotOwner, h.PATCH_ToggleChatbot)
	chat.Delete("/chatbot/:chatID", h.OwershipMiddleware.IsChatbotOwner, h.DELETE_Chatbot)
	chat.Post("/:chatID/upload", h.OwershipMiddleware.IsChatbotOwner, h.SubscriptionLimits.CheckLimit(constants.LimitDataSources), h.SubscriptionLimits.CheckLimit(constants.LimitTrainingData), h.POST_UploadFile)
	chat.Post("/:chatID/text", h.OwershipMiddleware.IsChatbotOwner, h.SubscriptionLimits.CheckLimit(constants.LimitDataSources), h.SubscriptionLimits.CheckLimit(constants.LimitTrainingData), h.POST_UploadText)
	chat.Post("/:chatID/website", h.OwershipMiddleware.IsChatbotOwner, h.SubscriptionLimits.CheckLimit(constants.LimitDataSources), h.POST_UploadWebsite)
	chat.Get("/:chatID/text", h.OwershipMiddleware.IsChatbotOwner, h.GET_TextSources)
	chat.Delete("/:chatID/text/:id", h.OwershipMiddleware.IsChatbotOwner, h.DELETE_TextSource)
	chat.Delete("/:chatID/files/:filename", h.OwershipMiddleware.IsChatbotOwner, h.DELETE_ChatFile)
	chat.Get("/:chatID/files", h.OwershipMiddleware.IsChatbotOwner, h.GET_ChatFiles)

	// Chat
	chat.Post("/:chatID/message", h.OwershipMiddleware.IsChatbotOwner, h.SubscriptionLimits.CheckMessageCredits(), h.POST_ChatMessage)
	chat.Post("/:chatID/stream-message", h.OwershipMiddleware.IsChatbotOwner, h.SubscriptionLimits.CheckMessageCredits(), h.POST_StreamChatMessage)
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

	response, err := h.ChatService.ProcessFileUpload(c.Context(), chatID, file)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to upload file", err, status)
	}

	return c.JSON(response)
}

// @Summary Upload plain text
// @Description Upload plain text to be indexed for chat context
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param body body models.TextUploadRequest true "Text payload"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/text [post]
func (h *ChatHandler) POST_UploadText(c *fiber.Ctx) error {
	chatID, err := h.ChatService.ParseChatID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}

	var req models.TextUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Text) == "" {
		return ErrorResponse(c, "Text is required", nil, http.StatusBadRequest)
	}

	if err := h.ChatService.ProcessTextUpload(c.Context(), chatID, req.Text); err != nil {
		return ErrorResponse(c, "Failed to upload text", err)
	}

	return c.JSON(models.MessageResponse{Message: "Text processed successfully"})
}

// @Summary Add website
// @Description Crawl a website from a root URL and index its text content
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param body body models.WebsiteUploadRequest true "Website URL"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/website [post]
func (h *ChatHandler) POST_UploadWebsite(c *fiber.Ctx) error {
	chatID, err := h.ChatService.ParseChatID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}
	var req models.WebsiteUploadRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	if err := h.ChatService.AddWebsite(c.Context(), chatID, req.URL); err != nil {
		return ErrorResponse(c, "Failed to index website", err)
	}

	return c.JSON(models.MessageResponse{Message: "Website indexed successfully"})
}

// @Summary List text sources
// @Description List text sources indexed for a chat session
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Success 200 {object} models.TextSourcesResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/text [get]
func (h *ChatHandler) GET_TextSources(c *fiber.Ctx) error {
	chatID, err := h.ChatService.ParseChatID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}

	files, err := h.ChatService.GetTextSources(c.Context(), chatID)
	if err != nil {
		return ErrorResponse(c, "Failed to retrieve text sources", err)
	}

	sources := make([]models.TextSourceInfo, 0, len(files))
	for _, f := range files {
		sources = append(sources, models.TextSourceInfo{
			ID:         f.ID,
			Title:      f.Filename,
			Size:       f.SizeBytes,
			UploadedAt: f.UploadedAt,
		})
	}

	return c.JSON(models.TextSourcesResponse{ChatID: chatID, Sources: sources})
}

// @Summary Delete text source
// @Description Delete a text source and its associated chunks
// @Tags chat
// @Accept json
// @Produce json
// @Param chatID path string true "Chat session ID"
// @Param id path string true "Text source ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/text/{id} [delete]
func (h *ChatHandler) DELETE_TextSource(c *fiber.Ctx) error {
	chatID, err := h.ChatService.ParseChatID(c.Params("chatID"))
	if err != nil {
		return ErrorResponse(c, "Invalid chat ID", err, http.StatusBadRequest)
	}
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, "Text source ID is required", nil, http.StatusBadRequest)
	}

	if err := h.ChatService.DeleteTextSource(c.Context(), chatID, id); err != nil {
		return ErrorResponse(c, "Failed to delete text source", err)
	}
	return c.JSON(models.MessageResponse{Message: "Text source deleted successfully"})
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
	rawFilename := c.Params("filename")
	// Decode in case the filename contains URL-escaped characters (spaces, etc.)
	filename, _ := url.PathUnescape(rawFilename)

	if chatID == "" || filename == "" {
		return ErrorResponse(c, "Chat ID and filename are required", nil, http.StatusBadRequest)
	}

	if err := h.ChatService.ProcessFileDelete(c.Context(), chatID, filename); err != nil {
		return ErrorResponse(c, "Failed to delete file", err)
	}

	return c.JSON(models.MessageResponse{
		Message: "File deleted successfully",
	})
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
	ctxData, status, msg, err := h.prepareChatMessageContext(c)
	if err != nil {
		if status == 0 {
			return err
		}
		return ErrorResponse(c, msg, err, status)
	}

	response, sessionID, err := h.ChatService.ChatWithChatbot(c.Context(), ctxData.chatbot, ctxData.query, ctxData.sessionID)
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
		"response":   response,
		"session_id": sessionID,
	})
}

// @Summary Stream chat message
// @Description Send a message and receive a streamed response with context from uploaded files
// @Tags chat
// @Accept json
// @Produce text/event-stream
// @Param message body models.ChatMessageRequest true "Chat message"
// @Param chatID path string true "Chat session ID"
// @Success 200 {string} string "Streamed response"
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/{chatID}/stream-message [post]
func (h *ChatHandler) POST_StreamChatMessage(c *fiber.Ctx) error {
	ctxData, status, msg, err := h.prepareChatMessageContext(c)
	if err != nil {
		if status == 0 {
			return err
		}
		return ErrorResponse(c, msg, err, status)
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ctx := c.Context()

	type streamEvent struct {
		Type      string `json:"type"`
		Content   string `json:"content,omitempty"`
		SessionID string `json:"session_id,omitempty"`
		Error     string `json:"error,omitempty"`
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		send := func(event streamEvent) error {
			payload, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err = fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return err
			}
			return w.Flush()
		}

		clientClosed := false
		completion, sessionID, err := h.ChatService.ChatWithChatbotStream(ctx, ctxData.chatbot, ctxData.query, ctxData.sessionID, func(callCtx context.Context, chunk string) error {
			if chunk == "" {
				return nil
			}
			if err := send(streamEvent{Type: "chunk", Content: chunk}); err != nil {
				clientClosed = true
				return err
			}
			return nil
		})
		if err != nil {
			if clientClosed || errors.Is(err, context.Canceled) {
				return
			}
			_ = send(streamEvent{Type: "error", Error: err.Error()})
			return
		}
		_ = send(streamEvent{Type: "done", Content: completion, SessionID: sessionID})
	})

	return nil
}

type chatMessageContext struct {
	chatbot   *db.Chatbot
	query     string
	sessionID *string
}

func (h *ChatHandler) prepareChatMessageContext(c *fiber.Ctx) (*chatMessageContext, int, string, error) {
	var req models.ChatMessageRequest
	chatID := c.Params("chatID")
	if chatID == "" {
		return nil, http.StatusBadRequest, "Chat ID is required", nil
	}

	if err := c.BodyParser(&req); err != nil {
		// Fallback for form value if body parsing fails
		req.Query = c.FormValue("query")
	}

	query, err := h.ChatService.ValidateAndParseQuery(&req, c.FormValue("query"))
	if err != nil {
		return nil, http.StatusBadRequest, "Invalid query parameter", err
	}

	user, err := GetUser(c)
	if err != nil {
		return nil, 0, "", err
	}

	chatbot, err := h.ChatService.GetChatbotForUser(c.Context(), chatID, user.ID)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrChatbotNotFound) {
			return nil, http.StatusNotFound, "Chatbot not found", err
		}
		return nil, http.StatusInternalServerError, "Failed to load chatbot", err
	}

	return &chatMessageContext{
		chatbot:   chatbot,
		query:     query,
		sessionID: req.SessionID,
	}, 0, "", nil
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

// @Summary Toggle chatbot enabled state
// @Description Enable or disable a chatbot
// @Tags chatbot
// @Accept json
// @Produce json
// @Param chatID path string true "Chatbot ID"
// @Param request body models.ChatbotToggleRequest true "Toggle request"
// @Success 200 {object} models.ChatbotResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot/{chatID}/toggle [patch]
func (h *ChatHandler) PATCH_ToggleChatbot(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	if chatID == "" {
		return ErrorResponse(c, "Chatbot ID is required", nil, http.StatusBadRequest)
	}

	var req models.ChatbotToggleRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	user, err := GetUser(c)
	if err != nil {
		return err
	}

	chatbot, err := h.ChatService.ToggleChatbotEnabled(c.Context(), chatID, user.ID, req.IsEnabled)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrChatbotNotFound) {
			return ErrorResponse(c, "Chatbot not found", err, http.StatusNotFound)
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedChatbotAccess) {
			return ErrorResponse(c, "You don't have permission to modify this chatbot", err, http.StatusForbidden)
		}
		return ErrorResponse(c, "Failed to toggle chatbot state", err)
	}
	return c.JSON(chatbot)
}

// @Summary Delete chatbot
// @Description Delete a chatbot by ID
// @Tags chatbot
// @Accept json
// @Produce json
// @Param chatID path string true "Chatbot ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /chat/chatbot/{chatID} [delete]
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
