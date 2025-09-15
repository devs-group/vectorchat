package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

// ConversationHandler handles conversation-related endpoints
type ConversationHandler struct {
	authMiddleware *middleware.AuthMiddleware
	chatService    *services.ChatService
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(
	authMiddleware *middleware.AuthMiddleware,
	chatService *services.ChatService,
) *ConversationHandler {
	return &ConversationHandler{
		authMiddleware: authMiddleware,
		chatService:    chatService,
	}
}

// RegisterRoutes registers conversation routes
func (h *ConversationHandler) RegisterRoutes(app *fiber.App) {
	conversation := app.Group("/conversation", h.authMiddleware.RequireAuth)

	// Conversation management
	conversation.Get("/conversations/:chatbotID", h.GetConversations)
	conversation.Get("/conversations/:chatbotID/:sessionID", h.GetConversationMessages)

	// Revision management
	conversation.Get("/revisions/:chatbotID", h.GetRevisions)
	conversation.Post("/revisions", h.CreateRevision)
	conversation.Put("/revisions/:revisionID", h.UpdateRevision)
	conversation.Delete("/revisions/:revisionID", h.DeactivateRevision)
}

// GetConversations retrieves all conversations for a chatbot
// @Summary Get conversations for a chatbot
// @Description Retrieves all conversations (sessions) with their messages for conversation review
// @Tags conversation
// @Accept json
// @Produce json
// @Param chatbotID path string true "Chatbot ID"
// @Param limit query int false "Number of conversations to return" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} models.ConversationsResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /conversation/conversations/{chatbotID} [get]
func (h *ConversationHandler) GetConversations(c *fiber.Ctx) error {
	// Get chatbot ID from path
	chatbotIDStr := c.Params("chatbotID")
	chatbotID, err := uuid.Parse(chatbotIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chatbot ID format",
		})
	}

	// Get pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get user ID from context (set by auth middleware)
	user, ok := c.Locals("user").(*db.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Verify the user owns this chatbot
	isOwner, err := h.chatService.CheckChatbotOwnership(c.Context(), chatbotID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify ownership",
		})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have access to this chatbot",
		})
	}

	// Get conversations
	response, err := h.chatService.GetConversations(c.Context(), chatbotID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve conversations",
		})
	}
	return c.JSON(response)
}

// GetConversationMessages retrieves all messages for a specific conversation (session)
// @Summary Get conversation messages
// @Description Retrieves all messages for a specific conversation session
// @Tags conversation
// @Accept json
// @Produce json
// @Param chatbotID path string true "Chatbot ID"
// @Param sessionID path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /conversation/conversations/{chatbotID}/{sessionID} [get]
func (h *ConversationHandler) GetConversationMessages(c *fiber.Ctx) error {
	// Parse IDs
	chatbotIDStr := c.Params("chatbotID")
	sessionIDStr := c.Params("sessionID")

	chatbotID, err := uuid.Parse(chatbotIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid chatbot ID format"})
	}
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session ID format"})
	}

	// Auth user
	user, ok := c.Locals("user").(*db.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Ownership check
	isOwner, err := h.chatService.CheckChatbotOwnership(c.Context(), chatbotID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify ownership"})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You don't have access to this chatbot"})
	}

	// Fetch messages via service (includes session->chatbot validation)
	msgs, err := h.chatService.GetConversationMessages(c.Context(), chatbotID, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve messages"})
	}

	return c.JSON(fiber.Map{"messages": msgs})
}

// GetRevisions retrieves all revisions for a chatbot
// @Summary Get revisions for a chatbot
// @Description Retrieves all answer revisions for a specific chatbot
// @Tags conversation
// @Accept json
// @Produce json
// @Param chatbotID path string true "Chatbot ID"
// @Param includeInactive query bool false "Include inactive revisions" default(false)
// @Success 200 {object} models.RevisionsListResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /conversation/revisions/{chatbotID} [get]
func (h *ConversationHandler) GetRevisions(c *fiber.Ctx) error {
	// Get chatbot ID from path
	chatbotIDStr := c.Params("chatbotID")
	chatbotID, err := uuid.Parse(chatbotIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chatbot ID format",
		})
	}

	// Get includeInactive parameter
	includeInactive := c.Query("includeInactive") == "true"

	// Get user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Verify the user owns this chatbot
	isOwner, err := h.chatService.CheckChatbotOwnership(c.Context(), chatbotID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify ownership",
		})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have access to this chatbot",
		})
	}

	// Get revisions
	revisions, err := h.chatService.GetRevisions(c.Context(), chatbotID, includeInactive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve revisions",
		})
	}

	// Convert to response format
	var responseRevisions []models.RevisionResponse
	for _, rev := range revisions {
		responseRevisions = append(responseRevisions, models.RevisionResponse{
			ID:                rev.ID,
			ChatbotID:         rev.ChatbotID,
			OriginalMessageID: rev.OriginalMessageID,
			Question:          rev.Question,
			OriginalAnswer:    rev.OriginalAnswer,
			RevisedAnswer:     rev.RevisedAnswer,
			RevisionReason:    rev.RevisionReason,
			RevisedBy:         rev.RevisedBy,
			CreatedAt:         rev.CreatedAt,
			UpdatedAt:         rev.UpdatedAt,
			IsActive:          rev.IsActive,
		})
	}

	return c.JSON(models.RevisionsListResponse{
		Revisions: responseRevisions,
	})
}

// CreateRevision creates a new answer revision
// @Summary Create a new answer revision
// @Description Creates a new revision to correct or improve an AI answer
// @Tags conversation
// @Accept json
// @Produce json
// @Param revision body models.CreateRevisionRequest true "Revision details"
// @Success 201 {object} models.RevisionResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /conversation/revisions [post]
func (h *ConversationHandler) CreateRevision(c *fiber.Ctx) error {
	// Parse request body
	var req models.CreateRevisionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Set the revised_by field to the current user
	req.RevisedBy = userID

	// Verify the user owns this chatbot
	isOwner, err := h.chatService.CheckChatbotOwnership(c.Context(), req.ChatbotID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify ownership",
		})
	}
	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have access to this chatbot",
		})
	}

	// Create the revision
	revision, err := h.chatService.CreateAnswerRevision(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create revision: " + err.Error(),
		})
	}

	// Return the created revision
	return c.Status(fiber.StatusCreated).JSON(models.RevisionResponse{
		ID:                revision.ID,
		ChatbotID:         revision.ChatbotID,
		OriginalMessageID: revision.OriginalMessageID,
		Question:          revision.Question,
		OriginalAnswer:    revision.OriginalAnswer,
		RevisedAnswer:     revision.RevisedAnswer,
		RevisionReason:    revision.RevisionReason,
		RevisedBy:         revision.RevisedBy,
		CreatedAt:         revision.CreatedAt,
		UpdatedAt:         revision.UpdatedAt,
		IsActive:          revision.IsActive,
	})
}

// UpdateRevision updates an existing answer revision
// @Summary Update an answer revision
// @Description Updates an existing revision's content or status
// @Tags conversation
// @Accept json
// @Produce json
// @Param revisionID path string true "Revision ID"
// @Param revision body models.UpdateRevisionRequest true "Updated revision details"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /conversation/revisions/{revisionID} [put]
func (h *ConversationHandler) UpdateRevision(c *fiber.Ctx) error {
	// Get revision ID from path
	revisionIDStr := c.Params("revisionID")
	revisionID, err := uuid.Parse(revisionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid revision ID format",
		})
	}

	// Parse request body
	var req models.UpdateRevisionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user ID from context - not needed for update since we don't verify ownership on individual revisions
	// The revision might belong to a different conversation user

	// Build updates map
	updates := make(map[string]interface{})
	if req.Question != nil {
		updates["question"] = *req.Question
	}
	if req.RevisedAnswer != nil {
		updates["revised_answer"] = *req.RevisedAnswer
	}
	if req.RevisionReason != nil {
		updates["revision_reason"] = *req.RevisionReason
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// Update the revision
	if err := h.chatService.UpdateRevision(c.Context(), revisionID, updates); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update revision: " + err.Error(),
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "Revision updated successfully",
	})
}

// DeactivateRevision deactivates an answer revision
// @Summary Deactivate an answer revision
// @Description Marks a revision as inactive (soft delete)
// @Tags conversation
// @Accept json
// @Produce json
// @Param revisionID path string true "Revision ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /conversation/revisions/{revisionID} [delete]
func (h *ConversationHandler) DeactivateRevision(c *fiber.Ctx) error {
	// Get revision ID from path
	revisionIDStr := c.Params("revisionID")
	revisionID, err := uuid.Parse(revisionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid revision ID format",
		})
	}

	// Deactivate the revision (no ownership check needed - any conversation can deactivate)
	if err := h.chatService.DeactivateRevision(c.Context(), revisionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deactivate revision: " + err.Error(),
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "Revision deactivated successfully",
	})
}
