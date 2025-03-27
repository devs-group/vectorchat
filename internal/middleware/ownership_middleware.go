package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
)

// OwnershipMiddleware checks chatbot ownership
type OwnershipMiddleware struct {
	chatbotStore *db.ChatbotStore
}

// NewOwnershipMiddleware creates a new ownership middleware
func NewOwnershipMiddleware(chatbotStore *db.ChatbotStore) *OwnershipMiddleware {
	return &OwnershipMiddleware{
		chatbotStore: chatbotStore,
	}
}

// IsChatbotOwner verifies if the authenticated user owns the specified chatbot
func (m *OwnershipMiddleware) IsChatbotOwner(c *fiber.Ctx) error {
	// Get user from context (set by AuthMiddleware)
	user, ok := c.Locals("user").(*db.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Parse chatID from params
	chatID := c.Params("chatID")
	id, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chatbot ID format",
		})
	}

	// Verify ownership
	isOwner, err := m.chatbotStore.CheckChatbotOwnership(c.Context(), id, user.ID)
	if err != nil {
		slog.Error("Failed to verify chatbot ownership", "error", err, "chat_id", chatID, "user_id", user.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	if !isOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to access this chatbot",
		})
	}

	return c.Next()
}
