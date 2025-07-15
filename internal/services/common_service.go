package services

import (
	"strconv"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// CommonService provides common utility functions that can be used across all services
type CommonService struct{}

// NewCommonService creates a new CommonService instance
func NewCommonService() *CommonService {
	return &CommonService{}
}

// ParseUUID parses a UUID string and returns a UUID or an error
func (s *CommonService) ParseUUID(uuidStr string) (uuid.UUID, error) {
	if uuidStr == "" {
		return uuid.Nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "UUID is required")
	}

	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(err, "invalid UUID format")
	}

	return parsedUUID, nil
}

// ParsePaginationParams parses and validates pagination parameters
func (s *CommonService) ParsePaginationParams(pageStr, limitStr string) (page, limit, offset int) {
	page = 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit = 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset = (page - 1) * limit
	return page, limit, offset
}
