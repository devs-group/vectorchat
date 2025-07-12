package services

import (
	"context"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// HomeService handles home-related operations
type HomeService struct {
	userRepo UserRepository
}

// NewHomeService creates a new home service
func NewHomeService(userRepo UserRepository) *HomeService {
	return &HomeService{
		userRepo: userRepo,
	}
}

// FindUserByID finds a user by ID (for session validation)
func (s *HomeService) FindUserByID(ctx context.Context, id string) (*User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// GetUserProfile gets user profile information
func (s *HomeService) GetUserProfile(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrUserNotFound, "user ID is required")
	}

	return s.userRepo.FindByID(ctx, userID)
}

// ValidateUserSession validates if a user session is valid
func (s *HomeService) ValidateUserSession(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	return user != nil, nil
}

// UpdateUserProfile updates user profile information
func (s *HomeService) UpdateUserProfile(ctx context.Context, userID, name, email string) (*User, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrUserNotFound, "user ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update user information
	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetDashboardData gets dashboard data for the home page
func (s *HomeService) GetDashboardData(ctx context.Context, userID string) (map[string]interface{}, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrUserNotFound, "user ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	dashboardData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"provider":   user.Provider,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
		"welcome_message": "Welcome to VectorChat!",
		"quick_actions": []string{
			"Create a new chatbot",
			"Upload documents",
			"Manage API keys",
			"View analytics",
		},
	}

	return dashboardData, nil
}
