package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/constants"
	stripe_sub "github.com/yourusername/vectorchat/pkg/stripe_sub"
)

// SubscriptionLimitsMiddleware middleware for checking plan-based limits
type SubscriptionLimitsMiddleware struct {
	svc         *stripe_sub.Service
	chatService *services.ChatService
}

// NewSubscriptionLimitsMiddleware creates a new subscription limits middleware
func NewSubscriptionLimitsMiddleware(svc *stripe_sub.Service, chatService *services.ChatService) *SubscriptionLimitsMiddleware {
	return &SubscriptionLimitsMiddleware{
		svc:         svc,
		chatService: chatService,
	}
}

// CheckLimit returns a middleware that validates against a specific plan limit
func (s *SubscriptionLimitsMiddleware) CheckLimit(limitKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*db.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		// Get user's plan
		plan, _, err := s.svc.GetUserPlan(c.Context(), &user.ID, user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve subscription",
			})
		}

		// Default to free plan limits if no active subscription
		limits := getDefaultLimits()
		if plan != nil {
			if features, ok := plan.PlanDefinition["features"].(map[string]interface{}); ok {
				limits = features
			}
		}

		// Check specific limit
		switch limitKey {
		case constants.LimitChatbots:
			return s.checkChatbotsLimit(c, user.ID, limits)
		case constants.LimitDataSources:
			return s.checkDataSourcesLimit(c, user.ID, limits)
		case constants.LimitTrainingData:
			return s.checkTrainingDataLimit(c, user.ID, limits)
		default:
			return c.Next()
		}
	}
}

func (s *SubscriptionLimitsMiddleware) checkChatbotsLimit(c *fiber.Ctx, userID string, limits map[string]interface{}) error {
	maxChatbots := getIntLimit(limits, constants.LimitChatbots, constants.DefaultChatbots)

	// Count existing chatbots
	chatbots, err := s.chatService.ListChatbots(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check chatbots limit",
		})
	}

	if len(chatbots) >= maxChatbots {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Chatbot limit reached. Please upgrade your plan.",
			"limit": maxChatbots,
			"used":  len(chatbots),
		})
	}

	return c.Next()
}

func (s *SubscriptionLimitsMiddleware) checkDataSourcesLimit(c *fiber.Ctx, userID string, limits map[string]interface{}) error {
	chatID := c.Params("chatID")
	if chatID == "" {
		return c.Next()
	}

	maxSources := getIntLimit(limits, constants.LimitDataSources, constants.DefaultDataSources)

	// Count existing data sources (files + websites)
	files, err := s.chatService.GetFilesByChatbotID(c.Context(), uuid.MustParse(chatID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check data sources limit",
		})
	}

	if len(files) >= maxSources {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Data sources limit reached. Please upgrade your plan.",
			"limit": maxSources,
			"used":  len(files),
		})
	}

	return c.Next()
}

func (s *SubscriptionLimitsMiddleware) checkTrainingDataLimit(c *fiber.Ctx, userID string, limits map[string]interface{}) error {
	chatID := c.Params("chatID")
	if chatID == "" {
		return c.Next()
	}

	// Parse training data limit
	maxDataStr := getStringLimit(limits, constants.LimitTrainingData, constants.DefaultTrainingData)
	maxBytes := parseDataSize(maxDataStr)

	// Get current usage
	files, err := s.chatService.GetFilesByChatbotID(c.Context(), uuid.MustParse(chatID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check training data limit",
		})
	}

	var totalBytes int64
	for _, file := range files {
		totalBytes += file.SizeBytes
	}

	if totalBytes >= maxBytes {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":      "Training data limit reached. Please upgrade your plan.",
			"limit":      maxDataStr,
			"used_bytes": totalBytes,
		})
	}

	return c.Next()
}

// Helper functions

func getDefaultLimits() map[string]interface{} {
	return map[string]interface{}{
		constants.LimitMessageCredits: constants.DefaultMessageCredits,
		constants.LimitTrainingData:   constants.DefaultTrainingData,
		constants.LimitChatbots:       constants.DefaultChatbots,
		constants.LimitDataSources:    constants.DefaultDataSources,
		constants.LimitEmbedWebsites:  true,
		constants.LimitAPIAccess:      true,
	}
}

func getIntLimit(limits map[string]interface{}, key string, defaultVal int) int {
	if val, ok := limits[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		case string:
			// Extract number from string like "5 data sources"
			var num int
			if _, err := parseStringWithNumber(v, &num); err == nil {
				return num
			}
		}
	}
	return defaultVal
}

func getStringLimit(limits map[string]interface{}, key string, defaultVal string) string {
	if val, ok := limits[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func parseDataSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	var multiplier int64 = 1
	if strings.Contains(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.Replace(sizeStr, "KB", "", 1)
	} else if strings.Contains(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.Replace(sizeStr, "MB", "", 1)
	} else if strings.Contains(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.Replace(sizeStr, "GB", "", 1)
	}

	var size float64
	if _, err := parseStringWithNumber(strings.TrimSpace(sizeStr), &size); err == nil {
		return int64(size * float64(multiplier))
	}

	// Default to 400KB if parsing fails
	return 400 * 1024
}

func parseStringWithNumber(s string, dest interface{}) (int, error) {
	switch d := dest.(type) {
	case *int:
		n, err := parseIntFromString(s)
		if err != nil {
			return 0, err
		}
		*d = n
		return 1, nil
	case *float64:
		f, err := parseFloatFromString(s)
		if err != nil {
			return 0, err
		}
		*d = f
		return 1, nil
	}
	return 0, nil
}

func parseIntFromString(s string) (int, error) {
	// Simple parser - extracts first number from string
	for _, part := range strings.Fields(s) {
		var n int
		if _, err := fmt.Sscanf(part, "%d", &n); err == nil {
			return n, nil
		}
	}
	return 0, fmt.Errorf("no number found")
}

func parseFloatFromString(s string) (float64, error) {
	for _, part := range strings.Fields(s) {
		if n, err := strconv.ParseFloat(part, 64); err == nil {
			return n, nil
		}
	}
	return 0, fmt.Errorf("no number found")
}

// CheckMessageCredits checks if user has message credits remaining
func (s *SubscriptionLimitsMiddleware) CheckMessageCredits() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*db.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		plan, _, err := s.svc.GetUserPlan(context.Background(), &user.ID, user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve subscription",
			})
		}

		limits := getDefaultLimits()
		if plan != nil {
			if features, ok := plan.PlanDefinition["features"].(map[string]interface{}); ok {
				limits = features
			}
		}

		maxCredits := getIntLimit(limits, constants.LimitMessageCredits, constants.DefaultMessageCredits)

		// TODO: Implement message credit tracking
		// For now, just pass through
		c.Locals("message_credits_limit", maxCredits)

		return c.Next()
	}
}
